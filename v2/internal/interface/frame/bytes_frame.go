package frame

import (
	"github.com/swipe-io/swipe/v2/internal/usecase/frame"
)

type bytesFrame struct {
}

func (f *bytesFrame) Frame(data []byte) ([]byte, error) {
	return data, nil
}

func NewBytesFrame() frame.Frame {
	return &bytesFrame{}

}
