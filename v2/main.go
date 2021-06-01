package v2

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/gookit/color"
	"github.com/pterm/pterm"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/ast"
	"github.com/swipe-io/swipe/v2/internal/gitattributes"
	_ "github.com/swipe-io/swipe/v2/internal/plugin/configenv"
	_ "github.com/swipe-io/swipe/v2/internal/plugin/gokit"
	"github.com/swipe-io/swipe/v2/internal/stcreator"
	"github.com/swipe-io/swipe/v2/swipe"
)

const Version = "v2.0.0-rc13"

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
	subcommands.Register(&genTplCmd{}, "")

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

	header := pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgWhite))
	pterm.Println(header.Sprint("Swipe - " + Version))

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
