package v2

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	goast "go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/google/subcommands"
	"github.com/gookit/color"
	"github.com/pterm/pterm"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/annotation"
	"github.com/swipe-io/swipe/v2/internal/ast"
	"github.com/swipe-io/swipe/v2/internal/format"
	"github.com/swipe-io/swipe/v2/internal/gitattributes"
	_ "github.com/swipe-io/swipe/v2/internal/plugin/config"
	_ "github.com/swipe-io/swipe/v2/internal/plugin/gokit"
	"github.com/swipe-io/swipe/v2/swipe"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	colorSuccess = color.Green.Render
	colorAccent  = color.Cyan.Render
	colorFail    = color.Red.Render
)

func Main() {
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&genCmd{}, "")
	subcommands.Register(&genOptionsCmd{}, "")
	subcommands.Register(&initCmd{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	allCmds := map[string]bool{
		"commands": true,
		"gen-opts": true,
		"init":     true,
		"help":     true,
		"flags":    true,
		"gen":      true,
		//"fix-comment": true,
	}

	header := pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgWhite))
	pterm.Println(header.Sprint("Swipe - " + swipe.Version))

	var code int
	if args := flag.Args(); len(args) == 0 || !allCmds[args[0]] {
		genCmd := &genCmd{}
		code = int(genCmd.Execute(context.Background(), flag.CommandLine))
	} else {
		code = int(subcommands.Execute(context.Background()))
	}
	os.Exit(code)
}

type initCmd struct {
	verbose  bool
	init     bool
	swipePkg string
	wd       string
}

func (*initCmd) Name() string { return "init" }
func (*initCmd) Synopsis() string {
	return ""
}
func (*initCmd) Usage() string {
	return ``
}

func (cmd *initCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.wd, "w", "", "")
}

func (cmd *initCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var err error
	if cmd.wd == "" {
		cmd.wd, err = os.Getwd()
		if err != nil {
			log.Println(colorFail("failed to get working directory: "), colorFail(err))
			return subcommands.ExitFailure
		}
	}
	log.Printf("%s: %s\n", color.Yellow.Render("Workdir"), cmd.wd)
	for name, data := range swipe.Options() {
		buf := bytes.NewBuffer(nil)
		path := filepath.Join(cmd.wd, "pkg", "swipe", name)
		if err := os.MkdirAll(path, 0775); err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		buf.WriteString("package " + name + "\n\n")
		buf.Write(data)

		filename := filepath.Join(path, "swipe.go")
		if err := os.Remove(filename); err != nil {
			//fmt.Println(err)
			//return subcommands.ExitFailure
		}
		if err := os.WriteFile(filename, buf.Bytes(), 0755); err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
	}
	return subcommands.ExitSuccess
}

type genCmd struct {
	verbose  bool
	init     bool
	swipePkg string
	wd       string
}

func (*genCmd) Name() string { return "gen" }
func (*genCmd) Synopsis() string {
	return "generate the *_gen.go file for each package"
}
func (*genCmd) Usage() string {
	return `swipe [packages]
  Given one or more packages, gen creates the config.go file for each.
  If no packages are listed, it defaults to ".".
`
}

func (cmd *genCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.verbose, "v", false, "show verbose output")
	f.BoolVar(&cmd.init, "init", false, "initial swipe project")
	f.StringVar(&cmd.swipePkg, "swipe-pkg", "", "package for generating swipe options file")
	f.StringVar(&cmd.wd, "w", "", "")
}

func (cmd *genCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	pterm.DefaultBox.Println("Thanks for using Swipe")
	//pterm.Println("")
	//progressbar, _ := pterm.DefaultProgressbar.WithTotal(10).Start()
	//
	//progressbar.Title = "Generate"
	//progressbar.Increment()
	//log.Printf("%s %s", color.Yellow.Render("Thanks for using"), color.LightBlue.Render("swipe"))

	log.Println(color.Yellow.Render("Please wait the command is running, it may take some time"))

	var err error

	if cmd.wd == "" {
		cmd.wd, err = os.Getwd()
		if err != nil {
			log.Println(colorFail("failed to get working directory: "), colorFail(err))
			return subcommands.ExitFailure
		}
	}
	log.Printf("%s: %s\n", color.Yellow.Render("Workdir"), cmd.wd)

	packages := f.Args()
	if data, err := ioutil.ReadFile(filepath.Join(cmd.wd, "pkgs")); err == nil {
		packages = append(packages, strings.Split(string(data), "\n")...)
	}
	log.Printf("%s: %s\n", color.Yellow.Render("Packages"), strings.Join(packages, ", "))

	loader, errs := ast.NewLoader(cmd.wd, os.Environ(), packages)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(colorFail(err))
		}
		return subcommands.ExitFailure
	}
	cfg, err := swipe.GetConfig(loader)
	if err != nil {
		log.Println(colorFail(err))
		return subcommands.ExitFailure
	}

	// clear all before generated files.
	_ = filepath.Walk(loader.WorkDir(), func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, "/vendor/") {
			if !info.IsDir() {
				if strings.Contains(info.Name(), "swipe_gen_") {
					_ = os.Remove(path)
				}
			}
		}
		return nil
	})

	result, errs := swipe.Generate(cfg)
	success := true

	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(colorFail(err))
		}
		success = false
	}

	diffExcludes := make([]string, 0, len(result))

	for _, g := range result {
		if len(g.Errs) > 0 {
			for _, err := range g.Errs {
				log.Println(colorFail(err))
			}
			success = false
			continue
		}
		if len(g.Content) == 0 {
			continue
		}

		diffExcludes = append(diffExcludes, strings.Replace(g.OutputPath, cfg.WorkDir+"/", "", -1))

		dirPath := filepath.Dir(g.OutputPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Printf("%s: failed to create dir %s: %v\n", colorSuccess(g.PkgPath), colorAccent(dirPath), colorFail(err))
			return subcommands.ExitFailure
		}
		err := ioutil.WriteFile(g.OutputPath, g.Content, 0755)
		if err == nil {
			if cmd.verbose {
				log.Printf("%s: wrote %s\n", colorSuccess(g.PkgPath), colorAccent(g.OutputPath))
			}
		} else {
			log.Printf("%s: failed to write %s: %v\n", colorSuccess(g.PkgPath), colorAccent(g.OutputPath), colorFail(err))
			success = false
		}
	}

	if !success {
		return subcommands.ExitFailure
	}

	if err := gitattributes.Generate(cfg.WorkDir, diffExcludes); err != nil {
		log.Println(colorFail(err))
		return subcommands.ExitFailure
	}

	log.Println(color.LightGreen.Render("Command execution completed successfully"))

	return subcommands.ExitSuccess
}

type genOptionsCmd struct {
}

func (cmd *genOptionsCmd) Name() string { return "gen-opts" }

func (cmd *genOptionsCmd) Synopsis() string { return "generating a plugin options" }

func (cmd *genOptionsCmd) Usage() string {
	return ``
}

func (cmd *genOptionsCmd) SetFlags(set *flag.FlagSet) {
}

func (cmd *genOptionsCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	wd, err := filepath.Abs(f.Arg(0))
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}
	fset := token.NewFileSet()
	d, err := parser.ParseDir(fset, wd, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
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
										buf.WriteString(fmt.Sprintf("// %s\n", typeName))
										buf.WriteString(fmt.Sprintf("type %s string\n", typeName))
									}
									if _, ok := optExists[key]; !ok {
										optExists[key] = struct{}{}
										buf.WriteString(fmt.Sprintf("// %s\n", opt.name))
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
									log.Println(colorFail("failed generate: "), colorFail(err))
									return subcommands.ExitFailure
								}

								f, err := os.Create(filepath.Join(wd, file.Name+"_gen.go"))
								if err != nil {
									log.Println(colorFail("failed generate: "), colorFail(err))
									return subcommands.ExitFailure
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
	return subcommands.ExitSuccess
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
