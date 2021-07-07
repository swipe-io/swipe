/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	goast "go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/annotation"
	"golang.org/x/tools/go/ast/astutil"
)

// optionsCmd represents the options command
var optionsCmd = &cobra.Command{
	Use:   "options",
	Short: "",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a directory options argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		wd := viper.GetString("work-dir")
		if wd == "" {
			wd, _ = cmd.Flags().GetString("work-dir")
		}
		if wd == "" {
			wd, err = os.Getwd()
			if err != nil {
				cmd.PrintErrf("failed to get working directory: %s", err)
				os.Exit(1)
			}
		}

		wd, _ = filepath.Abs(filepath.Join(wd, args[0]))

		cmd.Printf("Workdir: %s\n", wd)

		fset := token.NewFileSet()
		d, err := parser.ParseDir(fset, wd, nil, parser.ParseComments)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		for _, file := range d {
			p := doc.New(file, "./", 0)
			for _, t := range p.Types {
				if annotations, err := annotation.Parse(t.Doc); err == nil {
					if len(t.Decl.Specs) > 0 {
						if ts, ok := t.Decl.Specs[0].(*goast.TypeSpec); ok {
							if s, ok := ts.Type.(*goast.StructType); ok {
								buf := bytes.NewBuffer(nil)
								if a, err := annotations.Get("swipe"); err == nil {
									baseTypeName := a.Value() + "Option"

									buf.WriteString(fmt.Sprintf("// %s\n", a.Value()))
									buf.WriteString(fmt.Sprintf("func %s(opts ...%s) {}\n", a.Value(), baseTypeName))

									opts := getOpts(baseTypeName, s)
									optExists := map[string]struct{}{}
									optTypeExists := map[string]struct{}{}
									for _, opt := range opts {
										typeName := opt.typeName

										if !strings.HasSuffix(typeName, "Option") {
											typeName += "Option"
										}

										key := opt.name + ":" + typeName
										if _, ok := optTypeExists[typeName]; !ok {
											optTypeExists[typeName] = struct{}{}
											buf.WriteString(fmt.Sprintf("// %s ...\n", typeName))
											buf.WriteString(fmt.Sprintf("type %s string\n", typeName))
										}
										if _, ok := optExists[key]; !ok {
											optExists[key] = struct{}{}
											buf.WriteString(fmt.Sprintf("// %s ...\n", opt.name))
											if opt.isRepeat {
												buf.WriteString("// @type:\"repeat\"\n")
											}
											paramsStr := strings.Join(opt.params, ",")

											optsType := opt.optsType
											if optsType != "" {
												if !strings.HasSuffix(optsType, "Option") {
													optsType += "Option"
												}
												paramsStr += ",opts ..." + optsType
											}

											buf.WriteString(fmt.Sprintf("func %s(%s) %s { return \"implementation not generated, run swipe\" }\n", opt.name, paramsStr, typeName))
										}
									}

									data, err := format.Source(buf.Bytes())
									if err != nil {
										cmd.PrintErrf("failed generate: %s", err)
										os.Exit(1)
									}

									f, err := os.Create(filepath.Join(wd, file.Name+"_gen.go"))
									if err != nil {
										cmd.PrintErrf("failed generate: %s", err)
										os.Exit(1)
									}
									_, _ = f.WriteString("package " + p.Name + "\n")
									_, _ = f.WriteString(fmt.Sprintf("func (*%s) Options() []byte { return []byte(%s)}\n", ts.Name.Name, strconv.Quote(string(data))))
									_ = f.Close()
								}
							}
						}
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(optionsCmd)
}

func getFieldType(f *goast.Field) string {
	if i, ok := f.Type.(*goast.Ident); ok {
		if i.Obj != nil {
			if ts, ok := i.Obj.Decl.(*goast.TypeSpec); ok {
				return getExprType(ts.Type)
			}
		}
	}
	return getExprType(f.Type)
}

func getExprType(e goast.Expr) string {
	switch t := e.(type) {
	case *goast.Ident:
		return t.Name
	case *goast.ArrayType:
		lenStr := ""
		if t.Len != nil {
			lenStr = t.Len.(*goast.Ident).Name
		}
		return fmt.Sprintf("[%s]%s", lenStr, getExprType(t.Elt))
	case *goast.StarExpr:
		return getExprType(t.X)
	default:
		return "interface{}"
	}
}

func getOpt(optionName string, f *goast.Field, e goast.Expr, isRepeat bool) (result []optionFunc) {
	name, ok := getOptName(f)
	if !ok {
		return nil
	}
	switch t := e.(type) {
	case *goast.ArrayType:
		return getOpt(optionName, f, t.Elt, true)
	case *goast.StarExpr:
		return getOpt(optionName, f, t.X, isRepeat)
	case *goast.StructType:
		for _, ident := range f.Names {
			result = append(result, optionFunc{
				typeName: optionName,
				name:     ident.Name,
			})
		}
	case *goast.Ident:
		of := optionFunc{
			typeName: optionName,
			name:     name,
			isRepeat: isRepeat,
		}
		if ts, ok := t.Obj.Decl.(*goast.TypeSpec); ok {
			if s, ok := ts.Type.(*goast.StructType); ok {
				var hasOpts bool
				if s.Fields != nil {
					for _, f := range s.Fields.List {
						name, ok := getOptName(f)
						if !ok {
							continue
						}
						if isFiledOpt(f) {
							hasOpts = true
							expr := astutil.Unparen(f.Type)
							if e, ok := expr.(*goast.StarExpr); ok {
								expr = e.X
							}
							result = append(result, getOpt(ts.Name.Name, f, expr, false)...)
							continue
						}
						of.params = append(of.params, strcase.ToLowerCamel(name)+" "+getFieldType(f))
					}
				}
				if hasOpts {
					of.optsType = ts.Name.Name
				}
			}
		}
		result = append(result, of)
	}
	return
}

type optionFunc struct {
	params   []string
	typeName string
	name     string
	isRepeat bool
	optsType string
}

func getOpts(optionName string, s *goast.StructType) []optionFunc {
	if s.Fields == nil {
		return nil
	}
	var result []optionFunc
	for _, f := range s.Fields.List {
		expr := astutil.Unparen(f.Type)
		if e, ok := expr.(*goast.StarExpr); ok {
			expr = e.X
		}
		if opts := getOpt(optionName, f, expr, false); len(opts) > 0 {
			result = append(result, opts...)
		}

	}
	return result
}

func isFiledOpt(f *goast.Field) bool {
	if f.Tag != nil {
		if tags, err := structtag.Parse(strings.Trim(f.Tag.Value, "`")); err == nil {
			if t, err := tags.Get("swipe"); err == nil {
				if t.Value() == "option" {
					return true
				}
			}
		}
	}
	return false
}

func getOptName(f *goast.Field) (name string, ok bool) {
	name = f.Names[0].Name
	ok = true
	if f.Tag != nil {
		tags, err := structtag.Parse(strings.Trim(f.Tag.Value, "`"))
		if err == nil {
			if tag, err := tags.Get("mapstructure"); err == nil {
				if tag.Value() == "-" {
					ok = false
				}
				name = tag.Value()
			}
		}
	}
	return
}
