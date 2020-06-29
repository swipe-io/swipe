package generator

import (
	"context"
)

type Generator interface {
	Process(ctx context.Context) error
	Bytes() []byte
	PkgName() string
	OutputDir() string
	Filename() string
}
