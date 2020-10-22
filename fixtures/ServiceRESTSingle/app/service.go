package app

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ErrUnauthorized unauthorized.
type ErrUnauthorized struct{}

func (ErrUnauthorized) Error() string {
	return "unauthorized"
}

// StatusCode error value implements StatusCoder,
// the StatusCode will be used when encoding the error.
func (ErrUnauthorized) StatusCode() int {
	return 401
}

// ErrorCode error value implements ErrorCoder,
// the ErrorCode will be used when encoding the error.
func (ErrUnauthorized) ErrorCode() int {
	return -32001
}

// ErrForbidden forbidden.
type ErrForbidden struct{}

func (ErrForbidden) Error() string {
	return "forbidden"
}

// StatusCode error value implements StatusCoder,
// the StatusCode will be used when encoding the error.
func (ErrForbidden) StatusCode() int {
	return 403
}

// ErrorCode error value implements ErrorCoder,
// the ErrorCode will be used when encoding the error.
func (ErrForbidden) ErrorCode() int {
	return -32002
}

type Member struct {
	ID string `json:"id"`
}

type Members []*Member

type Data map[string]interface{}

type AliasData = Data

type GeoJSON struct {
	Type        string    `json:"-"`
	Coordinates []float64 `json:"coordinates200"`
}

type Profile struct {
	Phone string `json:"phone"`
}

type Recurse struct {
	Name    string     `json:"name"`
	Recurse []*Recurse `json:"recurse"`
}

type Kind string

type User struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Password  string     `json:"password"`
	Point     GeoJSON    `json:"point"`
	LastSeen  time.Time  `json:"last_seen"`
	Data      AliasData  `json:"data"`
	Photo     []byte     `json:"photo"`
	User      *User      `json:"user"`
	Profile   *Profile   `json:"profile"`
	Recurse   *Recurse   `json:"recurse"`
	Kind      Kind       `json:"kind"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AppInterface interface {
	InterfaceB
}

type InterfaceB interface {
	// Create new item of item.
	Create(ctx context.Context, newData AliasData, name string, data []byte, date time.Time) (err error)
	// Get item.
	Get(ctx context.Context, id int, name, fname string, price float32, n, b, cc int) (data User, err error)
	// GetAll more comment and more and more comment and more and more comment and more.
	// New line comment.
	GetAll(ctx context.Context, members Members) ([]*User, error)
	Delete(ctx context.Context, id uint) (a string, b string, err error)
	TestMethod(data map[string]interface{}, ss interface{}) (states map[string]map[int][]string, err error)
	TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error
}

type serviceB struct {
}

func (s *serviceB) Create(ctx context.Context, newData AliasData, name string, data []byte) (err error) {
	return &ErrUnauthorized{}
}

func (s *serviceB) Get(ctx context.Context, id int, name, fname string, price float32, n, b, cc int) (data User, err error) {
	panic("implement me")
}

func (s *serviceB) GetAll(ctx context.Context, members Members) ([]*User, error) {
	panic("implement me")
}

func (s *serviceB) Delete(ctx context.Context, id uint) (a string, b string, err error) {
	panic("implement me")
}

func (s *serviceB) TestMethod(data map[string]interface{}, ss interface{}) (states map[string]map[int][]string, err error) {
	panic("implement me")
}

func (s *serviceB) TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error {
	panic("implement me")
}
