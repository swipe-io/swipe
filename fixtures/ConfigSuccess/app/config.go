package app

import "time"

type Config struct {
	FldDuration   time.Duration     `env:",desc:Test duration required description"`
	FldTime       time.Time         `env:",desc:Test time required description"`
	FldString     string            `env:",desc:Test string required description"`
	FldBool       bool              `env:",desc:Test bool required description"`
	FldInt        int               `env:",desc:Test int required description"`
	FldInt8       int8              `env:",desc:Test int8 required description"`
	FldInt16      int16             `env:",desc:Test int16 required description"`
	FldInt32      int32             `env:",desc:Test int32 required description"`
	FldInt64      int64             `env:",desc:Test int64 required description"`
	FldUInt       uint              `env:",desc:Test uint required description"`
	FldUInt8      uint8             `env:",desc:Test uint8 required description"`
	FldUInt16     uint16            `env:",desc:Test uint16 required description"`
	FldUInt32     uint32            `env:",desc:Test uint32 required description"`
	FldUInt64     uint64            `env:",desc:Test uint64 required description"`
	FldFloat64    float64           `env:",desc:Test int required description"`
	FldFloat32    float32           `env:",desc:Test int required description"`
	FldStrings    []string          `env:",desc:Test []string required description"`
	FldMap        map[string]string `env:",desc:Test map[string]string required description"`
	ID3Ver        string            `env:",desc:Test number env name required description"`
	TestNumber123 string            `env:",desc:Test number env name required description"`
}
