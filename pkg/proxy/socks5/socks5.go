package socks5

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/net/context"

	"github.com/liamylian/lsocks/pkg/proxy"
)

const (
	socks5Version = uint8(5)
)

// Config is used to setup and configure a Server
type Config struct {
	// AuthMethods 鉴权方式，默认无需鉴权
	AuthMethods []Authenticator

	// Credentials 用户名密码鉴权配置
	// 如果设置，启用用户名密码认证
	// 如果未设置，且 AuthMethods 为空，无需鉴权
	Credentials proxy.CredentialStore

	// Resolver 自定义 DNS 解析器，默认为 DNSResolver
	Resolver proxy.NameResolver

	// Rules 用于自定义授权命令，默认为 PermitAll
	Rules RuleSet

	// Rewriter 用于透明重写地址，默认不重写
	// 在 RuleSet 之前调用
	Rewriter AddressRewriter

	// RequestReporter 用于统计转发请求数据
	RequestReporter proxy.TrafficReporter

	// ResponseReporter 用于统计转发响应数据
	ResponseReporter proxy.TrafficReporter

	// RequestCopier 用于统计转发请求数据
	RequestCopier proxy.Copier

	// ResponseCopier 用于统计转发请求数据
	ResponseCopier proxy.Copier

	// BindIP 用于 bind 和 udp associate 命令
	BindIP net.IP

	// Logger 自定义日志，默认为标准输出
	Logger *log.Logger

	// Dial 可选拨号函数
	Dial func(ctx context.Context, network, addr string) (net.Conn, error)
}

// Server 接收并处理 SOCKS5 请求
type Server struct {
	config      *Config
	authMethods map[uint8]Authenticator
}

func New(conf *Config) (*Server, error) {
	// 至少需要支持一种鉴权方式
	if len(conf.AuthMethods) == 0 {
		if conf.Credentials != nil {
			conf.AuthMethods = []Authenticator{&UserPassAuthenticator{conf.Credentials}}
		} else {
			conf.AuthMethods = []Authenticator{&NoAuthAuthenticator{}}
		}
	}

	// 确保有 DNS 解析器
	if conf.Resolver == nil {
		conf.Resolver = proxy.DNSResolver{}
	}

	// 确保有授权器
	if conf.Rules == nil {
		conf.Rules = PermitAll()
	}

	// 确保有日志
	if conf.Logger == nil {
		conf.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// 确保有数据转发器
	if conf.RequestCopier == nil {
		conf.RequestCopier = proxy.NewSimpleCopier()
	}
	if conf.ResponseCopier == nil {
		conf.ResponseCopier = proxy.NewSimpleCopier()
	}

	server := &Server{
		config:      conf,
		authMethods: make(map[uint8]Authenticator),
	}

	for _, a := range conf.AuthMethods {
		server.authMethods[a.GetCode()] = a
	}
	return server, nil
}

func (s *Server) ListenAndServe(network, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

//  1. Method Negotiation Request:
//     +----+----------+----------+
//     |VER | NMETHODS | METHODS  |
//     +----+----------+----------+
//     | 1  |    1     | 1 to 255 |
//     +----+----------+----------+
//
//  2. Method Negotiation Response:
//     +----+--------+
//     |VER | METHOD |
//     +----+--------+
//     | 1  |   1    |
//     +----+--------+
//
//  3. Request:
//     +----+-----+-------+------+----------+----------+
//     |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
//     +----+-----+-------+------+----------+----------+
//     | 1  |  1  | X'00' |  1   | Variable |    2     |
//     +----+-----+-------+------+----------+----------+
//
//  4. Response:
//     +----+-----+-------+------+----------+----------+
//     |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
//     +----+-----+-------+------+----------+----------+
//     | 1  |  1  | X'00' |  1   | Variable |    2     |
//     +----+-----+-------+------+----------+----------+
//
// 5. CONNECT / BIND / UDP ASSOCIATE
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	bufConn := bufio.NewReader(conn)

	// 读取版本
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		s.config.Logger.Printf("[ERR] socks: Failed to get version byte: %v", err)
		return
	}

	// 检查兼容性
	if version[0] != socks5Version {
		err := fmt.Errorf("unsupported SOCKS version: %v", version)
		s.config.Logger.Printf("[ERR] socks: %v", err)
		return
	}

	// 认证请求
	authContext, err := s.authenticate(conn, bufConn)
	if err != nil {
		err = fmt.Errorf("failed to authenticate: %v", err)
		s.config.Logger.Printf("[ERR] socks: %v", err)
		return
	}

	request, err := NewRequest(bufConn)
	if err != nil {
		if err == unrecognizedAddrType {
			if err := sendReply(conn, ReplyAddrTypeNotSupported, nil); err != nil {
				s.config.Logger.Printf("[ERR] failed to send reply: %v", err)
				return
			}
		}

		s.config.Logger.Printf("[ERR] failed to read destination address: %v", err)
		return
	}
	request.AuthContext = authContext
	if client, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		request.RemoteAddr = &AddrSpec{IP: client.IP, Port: client.Port}
	}

	// 处理请求
	if err := s.handleRequest(request, conn); err != nil {
		err = fmt.Errorf("failed to handle request: %v", err)
		s.config.Logger.Printf("[ERR] socks: %v", err)
		return
	}
}

func noAcceptableAuth(conn io.Writer) error {
	conn.Write([]byte{socks5Version, MethodNotAcceptable})
	return NoSupportedAuth
}

func readMethods(r io.Reader) ([]byte, error) {
	header := []byte{0}
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(r, methods, numMethods)
	return methods, err
}

func (s *Server) authenticate(conn io.Writer, bufConn io.Reader) (*AuthContext, error) {
	// Get the methods
	methods, err := readMethods(bufConn)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth methods: %v", err)
	}

	// Select a usable method
	for _, method := range methods {
		cator, found := s.authMethods[method]
		if found {
			return cator.Authenticate(bufConn, conn)
		}
	}

	// No usable method found
	return nil, noAcceptableAuth(conn)
}
