package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type Aggregate struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
}

func (g *Aggregate) Generate(ctx context.Context) []byte {
	for _, iface := range g.Interfaces {
		g.writeAggregate(iface)
	}
	return g.w.Bytes()
}

func (g *Aggregate) writeAggregate(iface *config.Interface) {
	ifaceType := iface.Named.Type.(*option.IfaceType)
	for _, m := range ifaceType.Methods {
		mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

		if mopt.Aggregate != nil {
			for _, aggregate := range mopt.Aggregate {
				fmt.Println("OK", aggregate)
			}
		}
	}
}

func (g *Aggregate) OutputDir() string {
	return ""
}

func (g *Aggregate) Filename() string {
	return "aggregate.go"
}
