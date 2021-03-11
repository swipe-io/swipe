package app1

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
}

type App interface {
	// Create new item of item.
	Create(ctx context.Context, name string, data []byte) (err error)
}

type serviceApp struct {
}

func (s *serviceApp) Create(ctx context.Context, name string, data []byte) (err error) {
	return nil
}
