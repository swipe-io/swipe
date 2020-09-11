package generator

import (
	"context"
)

type Generator interface {
	Prepare(ctx context.Context) error
	Process(ctx context.Context) error
	Bytes() []byte
	PkgName() string
	OutputDir() string
	Filename() string
}
