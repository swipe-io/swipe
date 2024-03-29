package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/subcommands"
	"github.com/gookit/color"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/fixcomment"
	"github.com/swipe-io/swipe/v2/internal/git"
	"github.com/swipe-io/swipe/v2/internal/interface/executor"
	"github.com/swipe-io/swipe/v2/internal/interface/factory"
	"github.com/swipe-io/swipe/v2/internal/interface/frame"
	"github.com/swipe-io/swipe/v2/internal/interface/gateway"
	"github.com/swipe-io/swipe/v2/internal/interface/processor"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/stcreator"
	up "github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

const Version = "v2.0.0-rc13"

var (
	colorSuccess = color.Green.Render
	colorAccent  = color.Cyan.Render
	colorFail    = color.Red.Render
)

var startGitAttrPattern = []byte("\n# /swipe gen\n")
var endGitAttrPattern = []byte("# swipe gen/\n")

func main() {
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&genCmd{}, "")
	subcommands.Register(&genTplCmd{}, "")
	subcommands.Register(&fixComment{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	allCmds := map[string]bool{
		"commands":    true,
		"gen-tpl":     true,
		"help":        true,
		"flags":       true,
		"gen":         true,
		"fix-comment": true,
	}

	log.Printf("%s %s", color.LightBlue.Render("Swipe"), color.Yellow.Render(Version))

	var code int
	if args := flag.Args(); len(args) == 0 || !allCmds[args[0]] {
		genCmd := &genCmd{}
		code = int(genCmd.Execute(context.Background(), flag.CommandLine))
	} else {
		code = int(subcommands.Execute(context.Background()))
	}
	os.Exit(code)
}

type genCmd struct {
	verbose  bool
	init     bool
	swipePkg string
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
}

func (cmd *genCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	log.Printf("%s %s", color.Yellow.Render("Thanks for using"), color.LightBlue.Render("swipe"))
	log.Println(color.Yellow.Render("Please wait the command is running, it may take some time"))

	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}

	packages := f.Args()
	if data, err := ioutil.ReadFile(filepath.Join(wd, "pkgs")); err == nil {
		packages = append(packages, strings.Split(string(data), "\n")...)
	}

	log.Printf("%s: %s\n", color.Yellow.Render("Packages"), strings.Join(packages, ", "))
	log.Printf("%s: %s\n", color.Yellow.Render("Workdir"), wd)

	if cmd.swipePkg == "" {
		cmd.swipePkg = "pkg/swipe"
	}

	swipePkgPath := filepath.Join(wd, cmd.swipePkg)
	swipePkgFilePath := filepath.Join(swipePkgPath, "swipe.go")
	parts := strings.Split(swipePkgPath, "/")
	pkgName := parts[len(parts)-1]

	gt := git.NewGIT()

	processorFactory := processor.NewFactory()

	processorFactory.Register(new(processor.Service).Name(), func(o *option.ResultOption, opr *option.Result) (up.Processor, error) {
		serviceGateway, err := gateway.NewServiceGateway(o.Pkg, opr.Data.Module.Dir, o.Option, opr.Data.GraphTypes, opr.Data.CommentFuncs, opr.Data.CommentFields, opr.Data.Enums, opr.Data.WorkDir, opr.ExternalOptions)
		if err != nil {
			return nil, err
		}
		return &processor.Service{
			ServiceGateway: serviceGateway,
			GIT:            gt,
		}, nil
	}, processor.ServiceOptions)

	processorFactory.Register(new(processor.Config).Name(), func(o *option.ResultOption, opr *option.Result) (up.Processor, error) {
		configGateway := gateway.NewConfigGateway(o.Option)
		return &processor.Config{
			ConfigGateway: configGateway,
		}, nil
	}, processor.ConfigOptions)

	astLoader := astloader.NewLoader(wd, os.Environ(), packages)
	optionLoader := option.NewLoader(astLoader)
	importerFactory := factory.NewImporterFactory()
	frameFactory := frame.NewFrameFactory(Version)
	generationExecutor := executor.NewGenerationExecutor(processorFactory, importerFactory, frameFactory, optionLoader)

	if cmd.init {
		if err := os.MkdirAll(swipePkgPath, 0755); err != nil {
			log.Println(colorFail(err))
			return subcommands.ExitFailure
		}

		var w writer.GoLangWriter

		w.W("// Code generated by Swipe %s. DO NOT EDIT.\n\n", Version)
		w.W("package %s\n\n", pkgName)

		w.W("// A Option is an option for a Swipe.\ntype Option string\n\n")
		w.W("// Inject the basic option for defining the generation.\nfunc Inject(Option) {}\n\n")

		for _, name := range processorFactory.Names() {
			f, ok := processorFactory.GetOptGen(name)
			if !ok {
				continue
			}
			_, _ = w.Write(f())
		}

		goSrc := w.Bytes()
		fmtSrc, err := format.Source(goSrc)
		if err != nil {
			log.Println(colorFail(fmt.Sprintf("error: %s\n ***\n%s\n***\n\n", err, string(goSrc))))
			return subcommands.ExitFailure
		}
		if err := ioutil.WriteFile(swipePkgFilePath, fmtSrc, 0755); err != nil {
			log.Println(colorFail(err))
			return subcommands.ExitFailure
		}

		fmt.Printf("%s: %s", colorSuccess("swipe options generate success"), colorAccent(swipePkgFilePath))

		return subcommands.ExitSuccess
	}

	if _, err := os.Stat(swipePkgFilePath); err != nil {
		log.Printf("%s %s\n", colorFail("swipe options file \""+swipePkgFilePath+"\" not found, run:"), color.Yellow.Render("swipe gen --init"))
		return subcommands.ExitFailure
	}

	// clear all before generated files.
	_ = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(path, "/vendor/") {
			if !info.IsDir() {
				if strings.Contains(info.Name(), "_gen") {
					_ = os.Remove(path)
				}
			}
		}
		return nil
	})

	results, errs := generationExecutor.Execute()

	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(colorFail(err))
		}
		return subcommands.ExitFailure
	}
	if len(results) == 0 {
		return subcommands.ExitSuccess
	}
	success := true

	diffExcludes := make([]string, 0, len(results))

	for _, g := range results {
		if len(g.Errs) > 0 {
			logErrors(g.Errs)
			log.Printf("%s: %s\n", g.PkgPath, colorFail("generate failed"))
			fmt.Println(string(g.Content))
			success = false
		}
		if len(g.Content) == 0 {
			continue
		}
		diffExcludes = append(diffExcludes, strings.Replace(g.OutputPath, wd+"/", "", -1))

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
		log.Println(colorFail("at least one generate failure"))
		return subcommands.ExitFailure
	} else {
		gitAttributesPath := filepath.Join(wd, ".gitattributes")
		var (
			f   *os.File
			err error
		)
		if _, err = os.Stat(gitAttributesPath); os.IsNotExist(err) {
			f, err = os.Create(gitAttributesPath)
			if err != nil {
				log.Println(colorFail("create .gitattributes fail: ", err))
				return subcommands.ExitFailure
			}
			f.Close()
		}
		data, err := ioutil.ReadFile(gitAttributesPath)
		if err != nil {
			log.Println(colorFail("read .gitattributes fail: ", err))
			return subcommands.ExitFailure
		}

		buf := new(bytes.Buffer)

		start := bytes.Index(data, startGitAttrPattern)
		end := bytes.Index(data, endGitAttrPattern)

		if start == -1 && end != -1 {
			log.Println(colorFail("corrupted .gitattributes not found start swipe patter"))
			return subcommands.ExitFailure
		}

		if start != -1 && end == -1 {
			log.Println(colorFail("corrupted .gitattributes not found end swipe patter"))
			return subcommands.ExitFailure
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
			log.Println(colorFail("fail write .gitattributes: ", err))
			return subcommands.ExitFailure
		}
	}

	log.Println(color.LightGreen.Render("Command execution completed successfully"))

	return subcommands.ExitSuccess
}

type genTplCmd struct {
	configFilepath string
}

func (cmd *genTplCmd) Name() string { return "gen-tpl" }

func (cmd *genTplCmd) Synopsis() string { return "generating a project through templates" }

func (cmd *genTplCmd) Usage() string {
	return `swipe gen-tpl [--config] 'projectName' templatesPath`
}

func (cmd *genTplCmd) SetFlags(set *flag.FlagSet) {
	set.StringVar(&cmd.configFilepath, "config", "", "config YAML path")
}

func (cmd *genTplCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}

	pkgName := f.Arg(0)
	if pkgName == "" {
		log.Println(colorFail("package name required"))
		return subcommands.ExitFailure
	}

	parts := strings.Split(pkgName, "/")

	projectID := parts[len(parts)-1]
	projectName := strcase.ToCamel(projectID)
	templatePath := f.Arg(1)
	if templatePath == "" {
		log.Println(colorFail("template path required"))
		return subcommands.ExitFailure
	}

	if cmd.configFilepath != "" {
		cmd.configFilepath, err = filepath.Abs(cmd.configFilepath)
		if err != nil {
			log.Println(colorFail(err.Error()))
			return subcommands.ExitFailure
		}
	}
	templatePath, err = filepath.Abs(templatePath)
	if err != nil {
		log.Println(colorFail(err.Error()))
		return subcommands.ExitFailure
	}
	stl := stcreator.NewProjectLoader(projectName, projectID, pkgName, wd)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Println(colorFail("template path do not exists: ", templatePath))
		return subcommands.ExitFailure
	}

	log.Println(colorAccent("config file: ", cmd.configFilepath))

	_, err = stl.Process(templatePath, cmd.configFilepath)
	if err != nil {
		log.Println(colorFail(err.Error()))
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type fixComment struct {
}

func (c *fixComment) Name() string {
	return "fix-comment"
}

func (c *fixComment) Synopsis() string {
	return ""
}

func (c *fixComment) Usage() string {
	return ""
}

func (c *fixComment) SetFlags(set *flag.FlagSet) {

}

func (c *fixComment) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}
	fixComment := fixcomment.NewFixComment(wd, os.Environ(), f.Args())
	fixes, err := fixComment.Execute()
	if err != nil {
		log.Println(colorFail("failed to fix comments: "), colorFail(err))
		return subcommands.ExitFailure
	}
	for _, fix := range fixes {
		err := ioutil.WriteFile(fix.Filepath, fix.Content, 0755)
		if err == nil {
			log.Printf("wrote %s\n", colorSuccess(fix.Filepath))
		}
	}
	return subcommands.ExitSuccess
}

func logErrors(errs []error) {
	for _, err := range errs {
		log.Println(strings.Replace(err.Error(), "\n", "\n\t", -1))
	}
}
