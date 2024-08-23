package sms

import "context"

type Service interface {
	Send(ctx context.Context, tplID string, args []string, number ...string) error
}
