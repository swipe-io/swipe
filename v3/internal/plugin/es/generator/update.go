package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type UpdateGenerator struct {
	w      writer.GoWriter
	Entity *option.NamedType
}

func (g *UpdateGenerator) Generate(ctx context.Context) []byte {
	g.w.W("type %sUpdate struct {}\n", g.Entity.Name)
	return g.w.Bytes()
}

func (g *UpdateGenerator) OutputDir() string {
	return ""
}

func (g *UpdateGenerator) Filename() string {
	return "update.go"
}
