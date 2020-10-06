package format

import (
	"fmt"
	"go/format"
	"os/exec"
)

func Source(src []byte) ([]byte, error) {
	var useGoimports bool
	_, err := exec.LookPath("goimports")
	if err == nil {
		useGoimports = true
	}
	if useGoimports {
		cmd := exec.Command("goimports")
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
			return nil, fmt.Errorf("error: %w", err)
		}
		return out, nil
	}
	return format.Source(src)
}
