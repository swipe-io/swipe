package frame

type BasicFrame struct {
}

func (f *BasicFrame) Frame(data []byte) ([]byte, error) {
	return data, nil
}

func NewBytesFrame() *BasicFrame {
	return &BasicFrame{}

}
