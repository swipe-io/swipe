package app

import "time"

type Config struct {
	FldDurationReq time.Duration     `env:",required,desc:Test duration required description"`
	FldTimeReq     time.Time         `env:",required,desc:Test time required description"`
	FldStringReq   string            `env:",required,desc:Test string required description"`
	FldBoolReq     bool              `env:",required,desc:Test bool required description"`
	FldIntReq      int               `env:",required,desc:Test int required description"`
	FldInt8Req     int8              `env:",required,desc:Test int8 required description"`
	FldInt16Req    int16             `env:",required,desc:Test int16 required description"`
	FldInt32Req    int32             `env:",required,desc:Test int32 required description"`
	FldInt64Req    int64             `env:",required,desc:Test int64 required description"`
	FldUIntReq     uint              `env:",required,desc:Test uint required description"`
	FldUInt8Req    uint8             `env:",required,desc:Test uint8 required description"`
	FldUInt16Req   uint16            `env:",required,desc:Test uint16 required description"`
	FldUInt32Req   uint32            `env:",required,desc:Test uint32 required description"`
	FldUInt64Req   uint64            `env:",required,desc:Test uint64 required description"`
	FldFloat64Req  float64           `env:",required,desc:Test int required description"`
	FldFloat32Req  float32           `env:",required,desc:Test int required description"`
	FldStringsReq  []string          `env:",required,desc:Test []string required description"`
	FldMapReq      map[string]string `env:",required,desc:Test map[string]string required description"`
	FldMapIntReq   map[string]int    `env:",required,desc:Test map[string]int required description"`
}
