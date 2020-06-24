package service

import (
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Transport struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *Transport) Write(opt *parser.Option) error {
	protocol := parser.MustOption(opt.At("protocol"))
	switch protocol.Value.String() {
	case "http":
		return newTransportHTTP(w.ctx, w.w).Write(opt)
	}
	return nil
}

func newTransport(ctx serviceCtx, w *writer.Writer) *Transport {
	return &Transport{w: w, ctx: ctx}
}
