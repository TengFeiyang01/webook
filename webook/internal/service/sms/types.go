package sms

import "context"

type Service interface {
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}

type NameArg struct {
	Val  string
	Name string
}
