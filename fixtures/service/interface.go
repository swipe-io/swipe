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

// ErrorCode error value implements ErrorCoder,
// the ErrorCode will be used when encoding the error.
func (*ErrUnauthorized) ErrorCode() int {
	return -32001
}

type Interface interface {
	Create(ctx context.Context, name string, data []byte) (err error)
	Get(ctx context.Context, id int, name, fname string, price float32, n int) (data user.User, err error)
	GetAll(ctx context.Context) ([]*user.User, error)
	Delete(ctx context.Context, id uint) (a string, b string, err error)
	TestMethod(data map[string]interface{}, ss interface{}) (states map[string]map[int][]string, err error)
}
