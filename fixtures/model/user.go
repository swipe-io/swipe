package model

type Point struct {
	T           string    `json:"-"`
	Coordinates []float64 `json:"coordinates"`
}

type User struct {
	Name     string
	Password string
	Point    Point
}
