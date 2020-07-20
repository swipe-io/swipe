package user

import (
	"time"

	"github.com/pborman/uuid"
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

type GeoJSON struct {
	Type        string    `json:"-"`
	Coordinates []float64 `json:"coordinates200"`
}

type Profile struct {
	Phone string `json:"phone"`
}

type User struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Password  string     `json:"password"`
	Point     GeoJSON    `json:"point"`
	LastSeen  time.Time  `json:"last_seen"`
	Photo     []byte     `json:"photo"`
	Profile   *Profile   `json:"profile"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
