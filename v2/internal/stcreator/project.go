package stcreator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/achiku/varfmt"
	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/format"

	"gopkg.in/yaml.v3"
)

type FormatError struct {
	Err  error
	Data []byte
}

func (e FormatError) Error() string {
	var lines []string
	for i, b := range bytes.Split(e.Data, []byte("\n")) {
		lines = append(lines, fmt.Sprintf("%d  %s", i+1, string(b)))
	}
	return fmt.Sprintf("%s:\n%s", e.Err.Error(), strings.Join(lines, "\n"))
}

var funcs = template.FuncMap{
	"ToLowerCamel":  strcase.ToLowerCamel,
	"ToCamel":       strcase.ToCamel,
	"ToSnake":       strcase.ToSnake,
	"ToKebab":       strcase.ToKebab,
	"PublicVarName": varfmt.PublicVarName,
	"Add": func(v, n int) int {
		return v + n
	},
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

func (l *ProjectLoader) loadEntities(loaders Loaders) (result []StructMetadata, err error) {
	for _, loader := range loaders {
		structs, err := loader.Process()
		if err != nil {
			return nil, err
		}
		result = append(result, structs...)
	}
	return
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
		fmtData, err := format.Source(data)
		if err != nil {
			return fmt.Errorf("filename: %s: %v", filename, FormatError{
				Err:  err,
				Data: data,
			})
		}
		data = fmtData
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

func (l *ProjectLoader) executeTemplate(name string, data []byte, varsMap interface{}) ([]byte, error) {
	var buf bytes.Buffer
	t, err := template.New(name).Funcs(funcs).Parse(string(data))
	if err != nil {
		return nil, err
	}
	if err := t.Execute(&buf, varsMap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *ProjectLoader) loadConfig(configFilepath string) (*Config, error) {
	var cfg Config
	if configFilepath != "" {
		configData, err := ioutil.ReadFile(configFilepath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(configData, &cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func (l *ProjectLoader) Process(dir, configFilepath string) (*Project, error) {
	cfg, err := l.loadConfig(configFilepath)
	if err != nil {
		return nil, err
	}
	structs, err := l.loadEntities(cfg.Loaders)
	if err != nil {
		return nil, err
	}
	wd := filepath.Join(l.wd, l.projectID)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		outputPath := filepath.Join(wd, strings.Replace(path, dir, "", -1))
		if !info.IsDir() {
			fileData, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), ".tpl") {
				normalizeName := l.normalizeName(info.Name())
				if strings.HasPrefix(info.Name(), "$struct") {
					for i, st := range structs {
						data, err := l.executeTemplate(st.Name, fileData, map[string]interface{}{
							"Structs":     structs,
							"Struct":      st,
							"Index":       i,
							"PkgName":     l.pkgName,
							"ProjectName": l.projectName,
							"ProjectID":   l.projectID,
						})
						if err != nil {
							return err
						}
						filename := strings.Replace(normalizeName, "$struct", strcase.ToSnake(st.Name), -1)
						if err := l.createFile(filepath.Join(filepath.Dir(outputPath), filename), data); err != nil {
							return err
						}
					}
				} else {
					data, err := l.executeTemplate(info.Name(), fileData, map[string]interface{}{
						"Structs":     structs,
						"PkgName":     l.pkgName,
						"ProjectName": l.projectName,
						"ProjectID":   l.projectID,
					})
					if err != nil {
						return err
					}
					if err := l.createFile(filepath.Join(filepath.Dir(outputPath), normalizeName), data); err != nil {
						return err
					}
				}
				return nil
			} else {
				if err := l.createFile(filepath.Join(filepath.Dir(outputPath), info.Name()), fileData); err != nil {
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
	for _, command := range cfg.Commands {
		command = strings.TrimSpace(command)
		parts := strings.Split(command, " ")
		if len(parts) > 0 {
			name := parts[0]
			rawArgs := parts[1:]
			args := make([]string, len(rawArgs))
			for i, arg := range rawArgs {
				args[i] = strings.TrimSpace(arg)
			}
			if name != "" {
				cmd := exec.Command(name, args...)
				cmd.Dir = wd
				stderr, err := cmd.StderrPipe()
				if err != nil {
					return nil, err
				}
				if err := cmd.Start(); err != nil {
					return nil, err
				}
				out, err := ioutil.ReadAll(stderr)
				if err != nil {
					return nil, err
				}
				fmt.Println(string(out))
				if err := cmd.Wait(); err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, nil
}

func NewProjectLoader(projectName, projectID, pkgName, wd string) *ProjectLoader {
	return &ProjectLoader{projectName: projectName, projectID: projectID, pkgName: pkgName, wd: wd}
}
