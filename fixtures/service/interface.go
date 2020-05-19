package service

import (
	"context"

	"github.com/swipe-io/swipe/fixtures/user"
)

// ErrUnauthorized unauthorized.
type ErrUnauthorized struct{}

func (*ErrUnauthorized) Error() string {
	return "unauthorized"
}

// StatusCode error value implements StatusCoder,
// the StatusCode will be used when encoding the error.
func (*ErrUnauthorized) StatusCode() int {
	return 403
}

type Interface interface {
	Create(ctx context.Context, name string) (err error)
	Get(ctx context.Context, id int, name, fname string, price float32, n int) (data user.User, err error)
	GetAll(ctx context.Context) (data []user.User, err error)
}
