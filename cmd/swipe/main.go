package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/swipe-io/swipe/v2"

	"github.com/google/subcommands"
	"github.com/gookit/color"
	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/interface/executor"
	"github.com/swipe-io/swipe/v2/internal/interface/factory"
	"github.com/swipe-io/swipe/v2/internal/interface/finder"
	"github.com/swipe-io/swipe/v2/internal/interface/frame"
	"github.com/swipe-io/swipe/v2/internal/interface/registry"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/stcreator"

	"golang.org/x/mod/modfile"
)

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
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&genCmd{}, "")
	subcommands.Register(&genTplCmd{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	allCmds := map[string]bool{
		"commands": true,
		"gen-tpl":  true,
		"version":  true,
		"help":     true,
		"flags":    true,
		"gen":      true,
		"show":     true,
	}

	log.Printf("%s %s", color.LightBlue.Render("Swipe"), color.Yellow.Render(swipe.Version))
	log.Printf("%s %s", color.Yellow.Render("Thanks for using"), color.LightBlue.Render("swipe"))
	log.Println(color.Yellow.Render("Please wait the command is running, it may take some time"))

	startCmd := time.Now()

	var code int
	if args := flag.Args(); len(args) == 0 || !allCmds[args[0]] {
		genCmd := &genCmd{}
		code = int(genCmd.Execute(context.Background(), flag.CommandLine))
	} else {
		code = int(subcommands.Execute(context.Background()))
	}
	if code == 0 {
		log.Println(color.LightGreen.Render("Command execution completed successfully"))
	}
	log.Printf("%s %s", color.LightBlue.Render("Time"), color.Yellow.Render(time.Now().Sub(startCmd).String()))
	os.Exit(code)
}

type versionCmd struct {
}

// Name returns the name of the command.
func (c *versionCmd) Name() string {
	return "version"
}

// Synopsis returns a short string (less than one line) describing the command.
func (c *versionCmd) Synopsis() string {
	return "version"
}

// Usage returns a long string explaining the command and giving usage
// information.
func (c *versionCmd) Usage() string {
	return "version"
}

// SetFlags adds the flags for this command to the specified set.
func (c *versionCmd) SetFlags(_ *flag.FlagSet) {

}

// Execute executes the command and returns an ExitStatus.
func (c *versionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	log.Println(swipe.Version)
	return subcommands.ExitSuccess
}

type genCmd struct {
	verbose bool
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
}

func (cmd *genCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}
	modBytes, err := ioutil.ReadFile(filepath.Join(wd, "go.mod"))
	if err != nil {
		log.Println(colorFail("failed read go.mod file: "), colorFail(err))
		return subcommands.ExitFailure
	}
	mod, err := modfile.Parse("go.mod", modBytes, nil)
	if err != nil {
		log.Println(colorFail("failed parse go.mod file: "), colorFail(err))
		return subcommands.ExitFailure
	}

	if mod.Module.Mod.Path != "github.com/swipe-io/swipe/v2" {
		foundReplace := false
		for _, replace := range mod.Replace {
			if replace.Old.Path == "github.com/swipe-io/swipe/v2" {
				foundReplace = true
				break
			}
		}
		if !foundReplace {
			for _, require := range mod.Require {
				if require.Mod.Path == "github.com/swipe-io/swipe/v2" && require.Mod.Version != swipe.Version {
					log.Println(colorFail("swipe cli version (" + swipe.Version + ") does not match package version (" + require.Mod.Version + ")"))
					return subcommands.ExitFailure
				}
			}
		}
	}

	l := option.NewLoader()
	fi := finder.NewServiceFinder(l)
	r := registry.NewRegistry(fi)
	i := factory.NewImporterFactory()
	ff := frame.NewFrameFactory(swipe.Version)
	ge := executor.NewGenerationExecutor(r, i, ff, l)

	ge.Cleanup(wd) // clear all before generated files.

	results, errs := ge.Execute(wd, os.Environ(), packages(f))

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
	srcPath := filepath.Join(build.Default.GOPATH, "src")
	wd, err := os.Getwd()
	if err != nil {
		log.Println(colorFail("failed to get working directory: "), colorFail(err))
		return subcommands.ExitFailure
	}
	basePkgName := strings.Replace(wd, srcPath+"/", "", -1)
	projectName := f.Arg(0)
	if projectName == "" {
		log.Println(colorFail("project name required"))
		return subcommands.ExitFailure
	}

	projectID := strcase.ToKebab(projectName)
	pkgName := filepath.Join(basePkgName, projectID)
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
	_, err = stl.Process(templatePath, cmd.configFilepath)
	if err != nil {
		log.Println(colorFail(err.Error()))
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func packages(f *flag.FlagSet) []string {
	pkgs := f.Args()
	if len(pkgs) == 0 {
		pkgs = []string{"."}
	}
	return pkgs
}

func logErrors(errs []error) {
	for _, err := range errs {
		log.Println(strings.Replace(err.Error(), "\n", "\n\t", -1))
	}
}
