package user

import (
	"time"

	"github.com/pborman/uuid"
)

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
