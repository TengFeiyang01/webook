package grpc

import "golang.org/x/net/context"

type UserService interface {
	GetById(ctx context.Context, id int64) (User error)
}

//type User struct{}
