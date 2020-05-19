package stcreator

import (
	"bytes"
	"html/template"
	"os"
)

// Visitor of create concrete type.
type Visitor interface {
	VisitDir(d *D, path string) error
	VisitFile(f *F, path string) error
	Exists(path string) bool
}

// I of concrete type file or dir.
type I interface {
	Accept(v Visitor, path string) error
}

type visitor struct {
	data     interface{}
	basePath string
}

func (v *visitor) createDirIfNeeded(path string) error {
	if !v.Exists(path) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (v *visitor) VisitDir(d *D, path string) error {
	path = path + "/" + d.Name
	if err := v.createDirIfNeeded(path); err != nil {
		return err
	}
	for _, i := range d.Children {
		if err := i.Accept(Visitor(v), path); err != nil {
			return err
		}
	}
	return nil
}

func (v *visitor) VisitFile(f *F, path string) (err error) {
	path = path + "/" + f.Name
	fi, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		fi.Close()
		if err != nil {
			os.Remove(fi.Name())
		}
	}()

	switch t := f.Template.(type) {
	case string:
		_, err = fi.WriteString(t)
		if err != nil {
			return err
		}
	case *template.Template:
		buf := new(bytes.Buffer)

		var data struct {
			Global interface{}
			Data   interface{}
		}

		data.Data = f.Data
		data.Global = v.data

		if err := t.Execute(buf, data); err != nil {
			return err
		}
		_, err = fi.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *visitor) Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// New new.
func New(data interface{}) Visitor {
	return &visitor{data: data}
}
