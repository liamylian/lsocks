package socks5

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"golang.org/x/net/context"
)

const (
	CommandConnect      = uint8(1) // 连接命令
	CommandBind         = uint8(2) // 绑定命令
	CommandUDPAssociate = uint8(3) // UDP 关联命令

	AddressIPV4 = uint8(1) // IPV4 地址
	AddressFQDN = uint8(3) // FQDN 地址
	AddressIPV6 = uint8(4) // IPV6 地址
)

const (
	ReplySuccess              uint8 = iota // 响应成功
	ReplyServerFailure                     // 响应失败
	ReplyRuleFailure                       // 授权失败
	ReplyNetworkUnreachable                // 网络不可达
	ReplyHostUnreachable                   // 主机不可达
	ReplyConnectionRefused                 // 连接拒绝
	ReplyTTLExpired                        // 超时
	ReplyCommandNotSupported               // 命令不支持
	ReplyAddrTypeNotSupported              // 地址类型不支持
)

var (
	unrecognizedAddrType = fmt.Errorf("unrecognized address type")
)

// AddressRewriter 用于透明重写目标地址
type AddressRewriter interface {
	Rewrite(ctx context.Context, request *Request) (context.Context, *AddrSpec)
}

// AddrSpec 用于描述目标地址，可能为IPv4、IPv6、或 FQDN
type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// Address 返回用于拨号的地址，优先使用 IP 地址，回退用 FQDN
func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

// Request 请求
type Request struct {
	// 协议版本
	Version uint8
	// 请求命令
	Command uint8
	// 协商认证上下文
	AuthContext *AuthContext
	// RemoteAddr 请求者地址
	RemoteAddr *AddrSpec
	// DestAddr 目标地址
	DestAddr *AddrSpec
	// realDestAddr 实际目标地址（可能被重写）
	realDestAddr *AddrSpec
	bufConn      io.Reader
}

type conn interface {
	Write([]byte) (int, error)
	RemoteAddr() net.Addr
}

// NewRequest creates a new Request from the tcp connection
func NewRequest(bufConn io.Reader) (*Request, error) {
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		return nil, fmt.Errorf("failed to get command version: %v", err)
	}

	// Ensure we are compatible
	if header[0] != socks5Version {
		return nil, fmt.Errorf("unsupported command version: %v", header[0])
	}

	// Read in the destination address
	dest, err := readAddrSpec(bufConn)
	if err != nil {
		return nil, err
	}

	request := &Request{
		Version:  socks5Version,
		Command:  header[1],
		DestAddr: dest,
		bufConn:  bufConn,
	}

	return request, nil
}

// handleRequest 处理认证成功后的请求
func (s *Server) handleRequest(req *Request, conn conn) error {
	ctx := context.Background()

	// Resolve the address if we have a FQDN
	dest := req.DestAddr
	if dest.FQDN != "" {
		ctx_, addr, err := s.config.Resolver.Resolve(ctx, dest.FQDN)
		if err != nil {
			if err := sendReply(conn, ReplyHostUnreachable, nil); err != nil {
				return fmt.Errorf("failed to send reply: %v", err)
			}
			return fmt.Errorf("failed to resolve destination '%v': %v", dest.FQDN, err)
		}
		ctx = ctx_
		dest.IP = addr
	}

	// Apply any address rewrites
	req.realDestAddr = req.DestAddr
	if s.config.Rewriter != nil {
		ctx, req.realDestAddr = s.config.Rewriter.Rewrite(ctx, req)
	}

	// Switch on the command
	switch req.Command {
	case CommandConnect:
		return s.handleConnect(ctx, conn, req)
	case CommandBind:
		return s.handleBind(ctx, conn, req)
	case CommandUDPAssociate:
		return s.handleAssociate(ctx, conn, req)
	default:
		if err := sendReply(conn, ReplyCommandNotSupported, nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("unsupported command: %v", req.Command)
	}
}

// handleConnect 处理 Connect 命令
func (s *Server) handleConnect(ctx context.Context, conn conn, req *Request) error {
	// Check if this is allowed
	if ctx_, ok := s.config.Rules.Allow(ctx, req); !ok {
		if err := sendReply(conn, ReplyRuleFailure, nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v blocked by rules", req.DestAddr)
	} else {
		ctx = ctx_
	}

	// Attempt to connect
	dial := s.config.Dial
	if dial == nil {
		dial = func(ctx context.Context, net_, addr string) (net.Conn, error) {
			return net.Dial(net_, addr)
		}
	}
	target, err := dial(ctx, "tcp", req.realDestAddr.Address())
	if err != nil {
		msg := err.Error()
		resp := ReplyHostUnreachable
		if strings.Contains(msg, "refused") {
			resp = ReplyConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = ReplyNetworkUnreachable
		}
		if err := sendReply(conn, resp, nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v failed: %v", req.DestAddr, err)
	}
	defer target.Close()

	// Send success
	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, ReplySuccess, &bind); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}

	// Start proxying
	errCh := make(chan error, 2)
	go s.forwardRequest(req, target, req.bufConn, errCh)
	go s.forwardResponse(req, conn, target, errCh)

	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

// handleBind 处理 Connect 命令
func (s *Server) handleBind(ctx context.Context, conn conn, req *Request) error {
	// Check if this is allowed
	if ctx_, ok := s.config.Rules.Allow(ctx, req); !ok {
		if err := sendReply(conn, ReplyRuleFailure, nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("bind to %v blocked by rules", req.DestAddr)
	} else {
		ctx = ctx_
	}

	// TODO: Support bind
	if err := sendReply(conn, ReplyCommandNotSupported, nil); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}
	return nil
}

// handleAssociate 处理 Connect 命令
func (s *Server) handleAssociate(ctx context.Context, conn conn, req *Request) error {
	// Check if this is allowed
	if ctx_, ok := s.config.Rules.Allow(ctx, req); !ok {
		if err := sendReply(conn, ReplyRuleFailure, nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("associate to %v blocked by rules", req.DestAddr)
	} else {
		ctx = ctx_
	}

	// TODO: Support associate
	if err := sendReply(conn, ReplyCommandNotSupported, nil); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}
	return nil
}

// forwardRequest 转发请求数据
func (s *Server) forwardRequest(req *Request, dst io.Writer, src io.Reader, errCh chan error) {
	n, err := s.config.RequestCopier.Copy(dst, src)
	if s.config.RequestReporter != nil {
		_ = s.config.RequestReporter.Report(req.AuthContext.UserIdentifier, n)
	}

	type closeWriter interface {
		CloseWrite() error
	}
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}

	errCh <- err
}

// forwardResponse 转发响应数据
func (s *Server) forwardResponse(req *Request, dst io.Writer, src io.Reader, errCh chan error) {
	n, err := s.config.ResponseCopier.Copy(dst, src)
	if s.config.ResponseReporter != nil {
		_ = s.config.ResponseReporter.Report(req.AuthContext.UserIdentifier, n)
	}

	type closeWriter interface {
		CloseWrite() error
	}
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}

	errCh <- err
}

// readAddrSpec 读取请求中的地址
// 期望为 地址类型（一个字节）+ 请求地址 + 端口
func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	d := &AddrSpec{}

	// Get the address type
	addrType := []byte{0}
	if _, err := r.Read(addrType); err != nil {
		return nil, err
	}

	// Handle on a per type basis
	switch addrType[0] {
	case AddressIPV4:
		addr := make([]byte, 4)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case AddressIPV6:
		addr := make([]byte, 16)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case AddressFQDN:
		if _, err := r.Read(addrType); err != nil {
			return nil, err
		}
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadAtLeast(r, fqdn, addrLen); err != nil {
			return nil, err
		}
		d.FQDN = string(fqdn)

	default:
		return nil, unrecognizedAddrType
	}

	// Read the port
	port := []byte{0, 0}
	if _, err := io.ReadAtLeast(r, port, 2); err != nil {
		return nil, err
	}
	d.Port = (int(port[0]) << 8) | int(port[1])

	return d, nil
}

// sendReply 用于响应相信
func sendReply(w io.Writer, resp uint8, addr *AddrSpec) error {
	// Format the address
	var addrType uint8
	var addrBody []byte
	var addrPort uint16
	switch {
	case addr == nil:
		addrType = AddressIPV4
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = AddressFQDN
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = AddressIPV4
		addrBody = []byte(addr.IP.To4())
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = AddressIPV6
		addrBody = []byte(addr.IP.To16())
		addrPort = uint16(addr.Port)

	default:
		return fmt.Errorf("failed to format address: %v", addr)
	}

	// Format the message
	msg := make([]byte, 6+len(addrBody))
	msg[0] = socks5Version
	msg[1] = resp
	msg[2] = 0 // Reserved
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	// Send the message
	_, err := w.Write(msg)
	return err
}
