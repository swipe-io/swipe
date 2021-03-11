//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadConfig() (cfg *Config, errs []error) {
	cfg = &Config{}
	fldDurationReqTmp, ok := os.LookupEnv("FLD_DURATION_REQ")
	if ok {
		fldDurationReqResult, err := time.ParseDuration(fldDurationReqTmp)
		if err != nil {
			errs = append(errs, err)
		}
		cfg.FldDurationReq = fldDurationReqResult
	} else {
		errs = append(errs, errors.New("env FLD_DURATION_REQ required"))
	}
	fldTimeReqTmp, ok := os.LookupEnv("FLD_TIME_REQ")
	if ok {
		fldTimeReqResult, err := time.Parse(time.RFC3339, fldTimeReqTmp)
		if err != nil {
			errs = append(errs, err)
		}
		cfg.FldTimeReq = fldTimeReqResult
	} else {
		errs = append(errs, errors.New("env FLD_TIME_REQ required"))
	}
	fldStringReqTmp, ok := os.LookupEnv("FLD_STRING_REQ")
	if ok {
		cfg.FldStringReq = fldStringReqTmp
	} else {
		errs = append(errs, errors.New("env FLD_STRING_REQ required"))
	}
	fldBoolReqTmp, ok := os.LookupEnv("FLD_BOOL_REQ")
	if ok {
		fldboolreqBool, err := strconv.ParseBool(fldBoolReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_BOOL_REQ error: %w", err))
		}
		cfg.FldBoolReq = bool(fldboolreqBool)
	} else {
		errs = append(errs, errors.New("env FLD_BOOL_REQ required"))
	}
	fldIntReqTmp, ok := os.LookupEnv("FLD_INT_REQ")
	if ok {
		fldintreqInt, err := strconv.Atoi(fldIntReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_INT_REQ error: %w", err))
		}
		cfg.FldIntReq = int(fldintreqInt)
	} else {
		errs = append(errs, errors.New("env FLD_INT_REQ required"))
	}
	fldInt8ReqTmp, ok := os.LookupEnv("FLD_INT8_REQ")
	if ok {
		fldint8reqInt8, err := strconv.Atoi(fldInt8ReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_INT8_REQ error: %w", err))
		}
		cfg.FldInt8Req = int8(fldint8reqInt8)
	} else {
		errs = append(errs, errors.New("env FLD_INT8_REQ required"))
	}
	fldInt16ReqTmp, ok := os.LookupEnv("FLD_INT16_REQ")
	if ok {
		fldint16reqInt16, err := strconv.Atoi(fldInt16ReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_INT16_REQ error: %w", err))
		}
		cfg.FldInt16Req = int16(fldint16reqInt16)
	} else {
		errs = append(errs, errors.New("env FLD_INT16_REQ required"))
	}
	fldInt32ReqTmp, ok := os.LookupEnv("FLD_INT32_REQ")
	if ok {
		fldint32reqInt32, err := strconv.Atoi(fldInt32ReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_INT32_REQ error: %w", err))
		}
		cfg.FldInt32Req = int32(fldint32reqInt32)
	} else {
		errs = append(errs, errors.New("env FLD_INT32_REQ required"))
	}
	fldInt64ReqTmp, ok := os.LookupEnv("FLD_INT64_REQ")
	if ok {
		fldint64reqInt64, err := strconv.Atoi(fldInt64ReqTmp)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_INT64_REQ error: %w", err))
		}
		cfg.FldInt64Req = int64(fldint64reqInt64)
	} else {
		errs = append(errs, errors.New("env FLD_INT64_REQ required"))
	}
	fldUIntReqTmp, ok := os.LookupEnv("FLD_U_INT_REQ")
	if ok {
		flduintreqUint, err := strconv.ParseUint(fldUIntReqTmp, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_U_INT_REQ error: %w", err))
		}
		cfg.FldUIntReq = uint(flduintreqUint)
	} else {
		errs = append(errs, errors.New("env FLD_U_INT_REQ required"))
	}
	fldUInt8ReqTmp, ok := os.LookupEnv("FLD_U_INT8_REQ")
	if ok {
		flduint8reqUint8, err := strconv.ParseUint(fldUInt8ReqTmp, 10, 8)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_U_INT8_REQ error: %w", err))
		}
		cfg.FldUInt8Req = uint8(flduint8reqUint8)
	} else {
		errs = append(errs, errors.New("env FLD_U_INT8_REQ required"))
	}
	fldUInt16ReqTmp, ok := os.LookupEnv("FLD_U_INT16_REQ")
	if ok {
		flduint16reqUint16, err := strconv.ParseUint(fldUInt16ReqTmp, 10, 16)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_U_INT16_REQ error: %w", err))
		}
		cfg.FldUInt16Req = uint16(flduint16reqUint16)
	} else {
		errs = append(errs, errors.New("env FLD_U_INT16_REQ required"))
	}
	fldUInt32ReqTmp, ok := os.LookupEnv("FLD_U_INT32_REQ")
	if ok {
		flduint32reqUint32, err := strconv.ParseUint(fldUInt32ReqTmp, 10, 32)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_U_INT32_REQ error: %w", err))
		}
		cfg.FldUInt32Req = uint32(flduint32reqUint32)
	} else {
		errs = append(errs, errors.New("env FLD_U_INT32_REQ required"))
	}
	fldUInt64ReqTmp, ok := os.LookupEnv("FLD_U_INT64_REQ")
	if ok {
		flduint64reqUint64, err := strconv.ParseUint(fldUInt64ReqTmp, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_U_INT64_REQ error: %w", err))
		}
		cfg.FldUInt64Req = uint64(flduint64reqUint64)
	} else {
		errs = append(errs, errors.New("env FLD_U_INT64_REQ required"))
	}
	fldFloat64ReqTmp, ok := os.LookupEnv("FLD_FLOAT64_REQ")
	if ok {
		fldfloat64reqFloat64, err := strconv.ParseFloat(fldFloat64ReqTmp, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_FLOAT64_REQ error: %w", err))
		}
		cfg.FldFloat64Req = float64(fldfloat64reqFloat64)
	} else {
		errs = append(errs, errors.New("env FLD_FLOAT64_REQ required"))
	}
	fldFloat32ReqTmp, ok := os.LookupEnv("FLD_FLOAT32_REQ")
	if ok {
		fldfloat32reqFloat32, err := strconv.ParseFloat(fldFloat32ReqTmp, 32)
		if err != nil {
			errs = append(errs, fmt.Errorf("convert FLD_FLOAT32_REQ error: %w", err))
		}
		cfg.FldFloat32Req = float32(fldfloat32reqFloat32)
	} else {
		errs = append(errs, errors.New("env FLD_FLOAT32_REQ required"))
	}
	fldStringsReqTmp, ok := os.LookupEnv("FLD_STRINGS_REQ")
	if ok {
		cfg.FldStringsReq = strings.Split(fldStringsReqTmp, ",")
	} else {
		errs = append(errs, errors.New("env FLD_STRINGS_REQ required"))
	}
	fldMapReqTmp, ok := os.LookupEnv("FLD_MAP_REQ")
	if ok {
		partsfldmapreq := strings.Split(fldMapReqTmp, ",")
		cfg.FldMapReq = make(map[string]string, len(partsfldmapreq))
		for _, s := range partsfldmapreq {
			kv := strings.Split(s, "=")
			if len(kv) == 2 {
				cfg.FldMapReq[kv[0]] = kv[1]
			}
		}
	} else {
		errs = append(errs, errors.New("env FLD_MAP_REQ required"))
	}
	return
}

func (cfg *Config) String() string {
	out := `
FLD_DURATION_REQ=` + fmt.Sprintf("%v", cfg.FldDurationReq) + ` ; Test duration required description
FLD_TIME_REQ=` + fmt.Sprintf("%v", cfg.FldTimeReq) + ` ; Test time required description
FLD_STRING_REQ=` + fmt.Sprintf("%v", cfg.FldStringReq) + ` ; Test string required description
FLD_BOOL_REQ=` + fmt.Sprintf("%v", cfg.FldBoolReq) + ` ; Test bool required description
FLD_INT_REQ=` + fmt.Sprintf("%v", cfg.FldIntReq) + ` ; Test int required description
FLD_INT8_REQ=` + fmt.Sprintf("%v", cfg.FldInt8Req) + ` ; Test int8 required description
FLD_INT16_REQ=` + fmt.Sprintf("%v", cfg.FldInt16Req) + ` ; Test int16 required description
FLD_INT32_REQ=` + fmt.Sprintf("%v", cfg.FldInt32Req) + ` ; Test int32 required description
FLD_INT64_REQ=` + fmt.Sprintf("%v", cfg.FldInt64Req) + ` ; Test int64 required description
FLD_U_INT_REQ=` + fmt.Sprintf("%v", cfg.FldUIntReq) + ` ; Test uint required description
FLD_U_INT8_REQ=` + fmt.Sprintf("%v", cfg.FldUInt8Req) + ` ; Test uint8 required description
FLD_U_INT16_REQ=` + fmt.Sprintf("%v", cfg.FldUInt16Req) + ` ; Test uint16 required description
FLD_U_INT32_REQ=` + fmt.Sprintf("%v", cfg.FldUInt32Req) + ` ; Test uint32 required description
FLD_U_INT64_REQ=` + fmt.Sprintf("%v", cfg.FldUInt64Req) + ` ; Test uint64 required description
FLD_FLOAT64_REQ=` + fmt.Sprintf("%v", cfg.FldFloat64Req) + ` ; Test int required description
FLD_FLOAT32_REQ=` + fmt.Sprintf("%v", cfg.FldFloat32Req) + ` ; Test int required description
FLD_STRINGS_REQ=` + fmt.Sprintf("%v", cfg.FldStringsReq) + ` ; Test []string required description
FLD_MAP_REQ=` + fmt.Sprintf("%v", cfg.FldMapReq) + ` ; Test map[string]string required description
`
	return out
}
