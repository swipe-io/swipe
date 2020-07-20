package service

import (
	"context"

	"github.com/swipe-io/swipe/fixtures/user"
)

type Interface interface {
	// Create new item of item.
	Create(ctx context.Context, name string, data []byte) (err error)
	// Get item.
	Get(ctx context.Context, id int, name, fname string, price float32, n, b, c int) (data user.User, err error)
	// GetAll more comment and more and more comment and more and more comment and more.
	// New line comment.
	GetAll(ctx context.Context) ([]*user.User, error)
	Delete(ctx context.Context, id uint) (a string, b string, err error)
	TestMethod(data map[string]interface{}, ss interface{}) (states map[string]map[int][]string, err error)
	TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error
}
