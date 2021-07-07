package gitattributes

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

var startGitAttrPattern = []byte("\n# /swipe gen\n")
var endGitAttrPattern = []byte("# swipe gen/\n")

func Generate(wd string, diffExcludes []string) error {
	gitAttributesPath := filepath.Join(wd, ".gitattributes")
	var (
		f   *os.File
		err error
	)
	if _, err = os.Stat(gitAttributesPath); os.IsNotExist(err) {
		f, err = os.Create(gitAttributesPath)
		if err != nil {
			return err
		}
		f.Close()
	}
	data, err := ioutil.ReadFile(gitAttributesPath)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	start := bytes.Index(data, startGitAttrPattern)
	end := bytes.Index(data, endGitAttrPattern)

	if start == -1 && end != -1 {
		return err
	}
	if start != -1 && end == -1 {
		return err
	}

	if start != -1 && end != -1 {
		buf.Write(data[:start])
		buf.Write(data[end+len(endGitAttrPattern):])
	}

	sort.Strings(diffExcludes)

	buf.Write(startGitAttrPattern)
	for _, exclude := range diffExcludes {
		buf.WriteString(exclude + " -diff\n")
	}
	buf.Write(endGitAttrPattern)

	if err := ioutil.WriteFile(gitAttributesPath, buf.Bytes(), 0755); err != nil {
		return err
	}
	return nil
}
