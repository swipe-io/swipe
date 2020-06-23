package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/swipe-io/swipe/pkg/gen"

	"github.com/google/subcommands"
)

const version = "v1.13.4"

func main() {
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&genCmd{}, "")

	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("swipe: ")
	log.SetOutput(os.Stderr)

	allCmds := map[string]bool{
		"commands": true,
		"version":  true,
		"help":     true,
		"flags":    true,
		"gen":      true,
		"show":     true,
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
		log.Println("failed to get working directory: ", err)
		return subcommands.ExitFailure
	}

	s := gen.NewSwipe(ctx, version, wd, os.Environ(), packages(f))

	results, errs := s.Generate()

	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
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
			log.Printf("%s: generate failed\n", g.PkgPath)
			success = false
		}
		if len(g.Content) == 0 {
			continue
		}
		err := ioutil.WriteFile(g.OutputPath, g.Content, 0755)
		if err == nil {
			log.Printf("%s: wrote %s\n", g.PkgPath, g.OutputPath)
		} else {
			log.Printf("%s: failed to write %s: %v\n", g.PkgPath, g.OutputPath, err)
			success = false
		}
	}
	if !success {
		log.Println("at least one generate failure")
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
