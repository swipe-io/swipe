//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
)

type ServiceLoggingMiddleware struct {
	next   InterfaceB
	logger log.Logger
}

func (s *ServiceLoggingMiddleware) Create(ctx context.Context, newData Data, name string, data []byte) error {
	var (
		err error
	)
	defer func(now time.Time) {
		logErr := err
		if le, ok := err.(interface{ LogError() error }); ok {
			logErr = le.LogError()
		}
		s.logger.Log("method", "Create", "took", time.Since(now), "name", name, "data", len(data), "err", logErr)
	}(time.Now())
	err = s.next.Create(ctx, newData, name, data)
	return err
}

func (s *ServiceLoggingMiddleware) Delete(ctx context.Context, id uint) (string, string, error) {
	var (
		a   string
		b   string
		err error
	)
	a, b, err = s.next.Delete(ctx, id)
	return a, b, err
}

func (s *ServiceLoggingMiddleware) Get(ctx context.Context, id uuid.UUID, name string, fname string, price float32, n int, b int, cc int) (User, error) {
	var (
		result User
		err    error
	)
	defer func(now time.Time) {
		logErr := err
		if le, ok := err.(interface{ LogError() error }); ok {
			logErr = le.LogError()
		}
		s.logger.Log("method", "Get", "took", time.Since(now), "id", id, "err", logErr)
	}(time.Now())
	result, err = s.next.Get(ctx, id, name, fname, price, n, b, cc)
	return result, err
}

func (s *ServiceLoggingMiddleware) GetAll(ctx context.Context, members Members) ([]*User, error) {
	var (
		result []*User
		err    error
	)
	result, err = s.next.GetAll(ctx, members)
	return result, err
}

func (s *ServiceLoggingMiddleware) TestMethod(data map[string]interface{}, ss interface{}) (map[string]map[int][]string, error) {
	var (
		result map[string]map[int][]string
		err    error
	)
	result, err = s.next.TestMethod(data, ss)
	return result, err
}

func (s *ServiceLoggingMiddleware) TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error {
	var (
		err error
	)
	err = s.next.TestMethod2(ctx, ns, utype, user, restype, resource, permission)
	return err
}

func (s *ServiceLoggingMiddleware) TestMethodOptionals(ctx context.Context, ns string) error {
	var (
		err error
	)
	err = s.next.TestMethodOptionals(ctx, ns)
	return err
}

func NewLoggingServiceMiddleware(s InterfaceB, logger log.Logger) InterfaceB {
	return &ServiceLoggingMiddleware{next: s, logger: logger}
}
