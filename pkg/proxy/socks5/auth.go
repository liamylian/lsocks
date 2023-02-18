package socks5

import (
	"fmt"
	"io"

	"github.com/liamylian/lsocks/pkg/proxy"
)

const (
	MethodNoAuth        = uint8(0)   // 无需鉴权
	MethodGSSAPI        = uint8(1)   // GSSAPI 鉴权
	MethodUserPassAuth  = uint8(2)   // 用户名密码鉴权
	MethodNotAcceptable = uint8(255) // 不支持的Method

	userAuthVersion = uint8(1)
	authSuccess     = uint8(0)
	authFailure     = uint8(1)
)

var (
	UserAuthFailed  = fmt.Errorf("user authentication failed")
	NoSupportedAuth = fmt.Errorf("no supported authentication mechanism")
)

// AuthContext 协商鉴权请求
type AuthContext struct {
	Method         uint8             // 认证方法
	UserIdentifier string            // 用户标识
	Payload        map[string]string // 认证过程载荷，对于 Method = MethodUserPassAuth，为用户名和密码
}

// Authenticator 鉴权器
type Authenticator interface {
	Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error)
	GetCode() uint8
}

// NoAuthAuthenticator 无需鉴权
type NoAuthAuthenticator struct{}

func (a NoAuthAuthenticator) GetCode() uint8 {
	return MethodNoAuth
}

func (a NoAuthAuthenticator) Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error) {
	_, err := writer.Write([]byte{socks5Version, MethodNoAuth})
	return &AuthContext{MethodNoAuth, "", nil}, err
}

// UserPassAuthenticator 用户名密码鉴权
type UserPassAuthenticator struct {
	Credentials proxy.CredentialStore
}

func (a UserPassAuthenticator) GetCode() uint8 {
	return MethodUserPassAuth
}

func (a UserPassAuthenticator) Authenticate(reader io.Reader, writer io.Writer) (*AuthContext, error) {
	// Tell the client to use user/pass auth
	if _, err := writer.Write([]byte{socks5Version, MethodUserPassAuth}); err != nil {
		return nil, err
	}

	// Get the version and username length
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return nil, err
	}

	// Ensure we are compatible
	if header[0] != userAuthVersion {
		return nil, fmt.Errorf("unsupported auth version: %v", header[0])
	}

	// Get the user name
	userLen := int(header[1])
	user := make([]byte, userLen)
	if _, err := io.ReadAtLeast(reader, user, userLen); err != nil {
		return nil, err
	}

	// Get the password length
	if _, err := reader.Read(header[:1]); err != nil {
		return nil, err
	}

	// Get the password
	passLen := int(header[0])
	pass := make([]byte, passLen)
	if _, err := io.ReadAtLeast(reader, pass, passLen); err != nil {
		return nil, err
	}

	// Verify the password
	if a.Credentials.Valid(string(user), string(pass)) {
		if _, err := writer.Write([]byte{userAuthVersion, authSuccess}); err != nil {
			return nil, err
		}
	} else {
		if _, err := writer.Write([]byte{userAuthVersion, authFailure}); err != nil {
			return nil, err
		}
		return nil, UserAuthFailed
	}

	// Done
	return &AuthContext{MethodUserPassAuth, string(user), map[string]string{"Username": string(user)}}, nil
}
