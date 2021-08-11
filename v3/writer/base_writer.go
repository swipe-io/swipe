package writer

import (
	"bytes"
	"fmt"
)

type TextWriter struct {
	bytes.Buffer
}

func (w *TextWriter) Line() {
	w.W("\n")
}

func (w *TextWriter) W(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&w.Buffer, format, args...)
}
