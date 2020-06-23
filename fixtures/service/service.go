package service

import (
	"context"

	"github.com/swipe-io/swipe/fixtures/user"
)

var _ Interface = new(Service)

type Service struct {
}

func (s *Service) Create(ctx context.Context, name string, data []byte) (err error) {
	return nil
}

func (s *Service) Get(ctx context.Context, id int, name, fname string, price float32, n int) (u user.User, err error) {
	return user.User{}, &ErrUnauthorized{}
}

func (s *Service) GetAll(ctx context.Context) (users []*user.User, err error) {
	return []*user.User{}, nil
}

func (s *Service) Delete(ctx context.Context, id uint) (string, string, error) {
	return "", "", nil
}

func (s *Service) TestMethod(data map[string]interface{}, ss interface{}) (map[string]map[int][]string, error) {
	return nil, nil
}
