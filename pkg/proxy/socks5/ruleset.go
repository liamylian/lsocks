package socks5

import (
	"golang.org/x/net/context"
)

// RuleSet 用于授权
type RuleSet interface {
	Allow(ctx context.Context, req *Request) (context.Context, bool)
}

// PermitAll 允许所有
func PermitAll() RuleSet {
	return &PermitCommand{true, true, true}
}

// PermitNone 拒绝所有
func PermitNone() RuleSet {
	return &PermitCommand{false, false, false}
}

// PermitCommand 授权实现
type PermitCommand struct {
	EnableConnect   bool
	EnableBind      bool
	EnableAssociate bool
}

func (p *PermitCommand) Allow(ctx context.Context, req *Request) (context.Context, bool) {
	switch req.Command {
	case CommandConnect:
		return ctx, p.EnableConnect
	case CommandBind:
		return ctx, p.EnableBind
	case CommandUDPAssociate:
		return ctx, p.EnableAssociate
	}

	return ctx, false
}
