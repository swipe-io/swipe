package user

import (
	"time"

	"github.com/pborman/uuid"
)

type GeoJSON struct {
	Type        string    `json:"-"`
	Coordinates []float64 `json:"coordinates"`
}

type Profile struct {
	Phone string
}

type User struct {
	ID       uuid.UUID
	Name     string
	Password string
	Point    GeoJSON
	LastSeen time.Time
	Photo    []byte
	Profile  Profile
}
