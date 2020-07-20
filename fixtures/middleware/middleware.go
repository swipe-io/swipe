package middleware

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/swipe-io/swipe/fixtures/user"
)

func AuthMiddleware() endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			return nil, GetErrUnauthorized()
		}
	}
}

func FirstGetErrUnauthorized() error {
	return GetErrUnauthorized()
}

func GetErrUnauthorized() error {
	return NestedGetErrUnauthorized()
}

func NestedGetErrUnauthorized() error {
	return user.ErrForbidden{}
}
