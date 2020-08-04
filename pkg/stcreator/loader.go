package stcreator

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	"gopkg.in/yaml.v3"
)

var funcs = template.FuncMap{
	"ToLowerCamel": strcase.ToLowerCamel,
	"ToCamel":      strcase.ToCamel,
	"ToSnake":      strcase.ToSnake,
	"ToKebab":      strcase.ToKebab,
}

type yamlData struct {
	Type   string    `yaml:"type"`
	Params yaml.Node `yaml:"params"`
}

type Loader struct {
	Params LoaderParams
}

type StructParam struct {
	Name, LowerName, ColumnName string
	Type, RawType               string
	Primary                     bool
	NotNull                     bool
	Default                     string
}

type Imports []StructImport

func (i Imports) At(p StructParam) string {
	for _, structImport := range i {
		if structImport.Param == p {
			return strconv.Quote(structImport.Pkg)
		}
	}
	return ""
}

type StructImport struct {
	Pkg   string
	Param StructParam
}

type StructMetadata struct {
	Name, LowerName string
	Primary         StructParam
	Params          []StructParam
	Imports         Imports
}

type LoaderParams interface {
	Name() string
	Process() ([]StructMetadata, error)
}

func (l *Loader) MarshalYAML() (interface{}, error) {
	return struct {
		Type   string       `yaml:"type"`
		Params LoaderParams `yaml:"params"`
	}{Type: l.Params.Name(), Params: l.Params}, nil
}

func (l *Loader) UnmarshalYAML(node *yaml.Node) error {
	var tmpData yamlData
	var dt LoaderParams
	if err := node.Decode(&tmpData); err != nil {
		return errors.New(err.Error())
	}
	f, ok := LoaderFactories[tmpData.Type]
	if !ok {
		return fmt.Errorf("could not find condition type %s", tmpData.Type)
	}
	dt = f()
	if err := tmpData.Params.Decode(dt); err != nil {
		return errors.New(err.Error())
	}
	l.Params = dt
	return nil
}

var LoaderFactories = map[string]func() LoaderParams{
	new(MongoLoader).Name(): func() LoaderParams {
		return new(MongoLoader)
	},
	new(PostgresLoader).Name(): func() LoaderParams {
		return new(PostgresLoader)
	},
}

type Entity struct {
	Name string
}

type Data struct {
	Structure []Node `yaml:"structure"`
}

type Import struct {
	Resource string `yaml:"resource"`
}

type Project struct {
	Structs []StructMetadata
}

type Node struct {
	Name     string                 `yaml:"name"`
	Template string                 `yaml:"template"`
	Data     map[string]interface{} `yaml:"data"`
	Children []Node                 `yaml:"children"`
}

type ProjectLoader struct {
	projectID   string
	projectName string
	pkgName     string
	wd          string
}

func (l *ProjectLoader) loadEntities(file string) ([]StructMetadata, error) {
	var data struct {
		Loader Loader `yaml:"loader"`
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	if data.Loader.Params != nil {
		structs, err := data.Loader.Params.Process()
		if err != nil {
			return nil, err
		}
		return structs, nil
	}
	return nil, nil
}

func (l *ProjectLoader) exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l *ProjectLoader) createDirIfNeeded(path string) error {
	if !l.exists(path) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (l *ProjectLoader) createFile(filename string, data []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()
	if filepath.Ext(filename) == ".go" {
		data, err = format.Source(data)
		if err != nil {
			return err
		}
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (l *ProjectLoader) normalizeName(filename string) string {
	return strings.TrimSuffix(filename, ".tpl")
}

func (l *ProjectLoader) executeTemplate(data []byte, varsMap interface{}) ([]byte, error) {
	var buf bytes.Buffer
	t, err := template.New("template").Funcs(funcs).Parse(string(data))
	if err != nil {
		return nil, err
	}
	if err := t.Execute(&buf, varsMap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *ProjectLoader) Load(dir, entitiesFile string) (*Project, error) {
	structs, err := l.loadEntities(entitiesFile)
	if err != nil {
		return nil, err
	}
	wd := filepath.Join(l.wd, l.projectID)
	varsMap := map[string]interface{}{
		"Structs":     structs,
		"PkgName":     l.pkgName,
		"ProjectName": l.projectName,
		"ProjectID":   l.projectID,
	}
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		outputPath := filepath.Join(wd, strings.Replace(path, dir, "", -1))
		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".tpl") {
				normalizeName := l.normalizeName(info.Name())
				if strings.HasPrefix(info.Name(), "$struct") {
					for _, st := range structs {
						varsMap["Struct"] = st
						data, err = l.executeTemplate(data, varsMap)
						if err != nil {
							return err
						}
						filename := strings.Replace(normalizeName, "$struct", strcase.ToSnake(st.Name), -1)
						if err := l.createFile(filepath.Join(filepath.Dir(outputPath), filename), data); err != nil {
							return err
						}
					}
				} else {
					data, err := l.executeTemplate(data, varsMap)
					if err != nil {
						return err
					}
					if err := l.createFile(filepath.Join(filepath.Dir(outputPath), normalizeName), data); err != nil {
						return err
					}
				}
				return nil
			} else {
				if err := l.createFile(filepath.Join(filepath.Dir(outputPath), info.Name()), data); err != nil {
					return err
				}
			}
		} else {
			if err := l.createDirIfNeeded(outputPath); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func NewProjectLoader(projectName, projectID, pkgName, wd string) *ProjectLoader {
	return &ProjectLoader{projectName: projectName, projectID: projectID, pkgName: pkgName, wd: wd}
}
