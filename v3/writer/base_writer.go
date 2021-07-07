package writer

import (
	"bytes"
	"fmt"
)

type BaseWriter struct {
	bytes.Buffer
}

func (w *BaseWriter) Line() {
	w.W("\n")
}

func (w *BaseWriter) W(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(&w.Buffer, format, args...)
}
