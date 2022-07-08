package swipe

import "context"

type Generator interface {
	Generate(ctx context.Context) []byte
	OutputPath() string
	Filename() string
}

type GeneratorPackage interface {
	Package() string
}
