package frame

import (
	"bytes"
	"fmt"
	"os/exec"
)

type JSFrame struct {
	version string
}

func (f *JSFrame) Frame(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString("// Code generated by Swipe " + f.version + ". DO NOT EDIT.\n\n")
	buf.Write(data)
	cmd := exec.Command("prettier", "--stdin-filepath", "prettier.js", "--trailing-comma", "none", "--no-config")
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
		return nil, fmt.Errorf("error: %w\n ***\n%s\n***\n\n***%s***\n\n", err, string(out), string(data))
	}
	return out, nil
}

func NewJSFrame(version string) *JSFrame {
	return &JSFrame{version: version}

}
