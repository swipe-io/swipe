package app

import "time"

type Config struct {
	FldDurationReq time.Duration     `env:",required,use_zero,desc:Test duration required description"`
	FldTimeReq     time.Time         `env:",required,use_zero,desc:Test time required description"`
	FldStringReq   string            `env:",required,use_zero,desc:Test string required description"`
	FldBoolReq     bool              `env:",required,use_zero,desc:Test bool required description"`
	FldIntReq      int               `env:",required,use_zero,desc:Test int required description"`
	FldInt8Req     int8              `env:",required,use_zero,desc:Test int8 required description"`
	FldInt16Req    int16             `env:",required,use_zero,desc:Test int16 required description"`
	FldInt32Req    int32             `env:",required,use_zero,desc:Test int32 required description"`
	FldInt64Req    int64             `env:",required,use_zero,desc:Test int64 required description"`
	FldUIntReq     uint              `env:",required,use_zero,desc:Test uint required description"`
	FldUInt8Req    uint8             `env:",required,use_zero,desc:Test uint8 required description"`
	FldUInt16Req   uint16            `env:",required,use_zero,desc:Test uint16 required description"`
	FldUInt32Req   uint32            `env:",required,use_zero,desc:Test uint32 required description"`
	FldUInt64Req   uint64            `env:",required,use_zero,desc:Test uint64 required description"`
	FldFloat64Req  float64           `env:",required,use_zero,desc:Test int required description"`
	FldFloat32Req  float32           `env:",required,use_zero,desc:Test int required description"`
	FldStringsReq  []string          `env:",required,use_zero,desc:Test []string required description"`
	FldMapReq      map[string]string `env:",required,use_zero,desc:Test map[string]string required description"`
}
