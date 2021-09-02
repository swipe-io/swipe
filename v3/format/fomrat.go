package format

import (
	"fmt"
	"go/format"
	"os/exec"
)

func Source(src []byte) ([]byte, error) {
	var useGoImports bool
	_, err := exec.LookPath("gofmt")
	if err == nil {
		useGoImports = true
	}
	if useGoImports {
		cmd := exec.Command("gofmt")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		go func() {
			defer stdin.Close()
			_, _ = stdin.Write(src)
		}()
		out, err := cmd.Output()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				return nil, fmt.Errorf("error: %s %w", string(err.Stderr), err)
			}
			return nil, err
		}
		return out, nil
	}
	return format.Source(src)
}
