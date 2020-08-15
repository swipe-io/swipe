package main

import (
	"context"
	"flag"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/gookit/color"
	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/astloader"
	"github.com/swipe-io/swipe/pkg/gen"
	"github.com/swipe-io/swipe/pkg/stcreator"

	"golang.org/x/mod/modfile"
)

const version = "v1.25.7"

var (
	colorSuccess = color.Green.Render
	colorAccent  = color.Cyan.Render
	colorFail    = color.Red.Render
)

func main() {
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&genCmd{}, "")
	subcommands.Register(&crudServiceCmd{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("swipe: ")
	log.SetOutput(os.Stderr)

	allCmds := map[string]bool{
		"commands":     true,
		"crud-service": true,
		"version":      true,
		"help":         true,
		"flags":        true,
		"gen":          true,
		"show":         true,
	}
	if args := flag.Args(); len(args) == 0 || !allCmds[args[0]] {
		genCmd := &genCmd{}
		os.Exit(int(genCmd.Execute(context.Background(), flag.CommandLine)))
	}
	os.Exit(int(subcommands.Execute(context.Background())))
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
	log.Println(version)
	return subcommands.ExitSuccess
}

type genCmd struct {
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

	if mod.Module.Mod.Path != "github.com/swipe-io/swipe" {
		foundReplace := false
		for _, replace := range mod.Replace {
			if replace.Old.Path == "github.com/swipe-io/swipe" {
				foundReplace = true
				break
			}
		}
		if !foundReplace {
			for _, require := range mod.Require {
				if require.Mod.Path == "github.com/swipe-io/swipe" && require.Mod.Version != version {
					log.Println(colorFail("swipe cli version (" + version + ") does not match package version (" + require.Mod.Version + ")"))
					return subcommands.ExitFailure
				}
			}
		}
	}
	l := astloader.NewLoader(wd, os.Environ(), packages(f))
	s := gen.NewSwipe(ctx, version, l)
	results, errs := s.Generate()
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

	for _, g := range results {
		if len(g.Errs) > 0 {
			logErrors(g.Errs)
			log.Printf("%s: %s\n", g.PkgPath, colorFail("generate failed"))
			success = false
		}
		if len(g.Content) == 0 {
			continue
		}
		err := ioutil.WriteFile(g.OutputPath, g.Content, 0755)
		if err == nil {
			log.Printf("%s: wrote %s\n", colorSuccess(g.PkgPath), colorAccent(g.OutputPath))
		} else {
			log.Printf("%s: failed to write %s: %v\n", colorSuccess(g.PkgPath), colorAccent(g.OutputPath), colorFail(err))
			success = false
		}
	}
	if !success {
		log.Println(colorFail("at least one generate failure"))
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type crudServiceCmd struct {
	configFilepath string
}

func (cmd *crudServiceCmd) Name() string { return "crud-service" }

func (cmd *crudServiceCmd) Synopsis() string { return "generate CRUD service structure" }

func (cmd *crudServiceCmd) Usage() string {
	return `swipe crud-service [-config] projectName templatesPath`
}

func (cmd *crudServiceCmd) SetFlags(set *flag.FlagSet) {
	set.StringVar(&cmd.configFilepath, "config", "", "config file path")
}

func (cmd *crudServiceCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
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
