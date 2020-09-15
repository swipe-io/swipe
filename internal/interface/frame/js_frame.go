package frame

import (
	"bytes"
	"os/exec"

	"github.com/swipe-io/swipe/v2/internal/usecase/frame"
)

type jsFrame struct {
	version string
}

func (f *jsFrame) Frame(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString("// Code generated by Swipe " + f.version + ". DO NOT EDIT.\n\n")
	buf.Write(data)

	cmd := exec.Command("prettier", "--stdin-filepath", "prettier.js")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		defer stdin.Close()
		_, _ = stdin.Write(buf.Bytes())
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NewJSFrame(version string) frame.Frame {
	return &jsFrame{version: version}

}