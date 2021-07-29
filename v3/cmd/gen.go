package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/swipe-io/swipe/v3/internal/ast"
	"github.com/swipe-io/swipe/v3/internal/gitattributes"
	"github.com/swipe-io/swipe/v3/swipe"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate code",
	Long:  ``,
	Args: func(cmd *cobra.Command, packages []string) error {
		if len(viper.GetStringSlice("packages")) == 0 && len(packages) < 1 {
			return errors.New("requires a packages argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, packages []string) {
		var err error

		cmd.Println("Please wait the command is running, it may take some time")

		if len(packages) == 0 {
			packages = viper.GetStringSlice("packages")
		}

		wd := viper.GetString("work-dir")
		if wd == "" {
			wd, _ = cmd.Flags().GetString("work-dir")
		}
		verbose, _ := cmd.Flags().GetBool("verbose")

		if wd == "" {
			wd, err = os.Getwd()
			if err != nil {
				cmd.PrintErrf("failed to get working directory: ", err)
				os.Exit(1)
			}
		}

		cmd.Printf("Workdir: %s\n", wd)

		if data, err := ioutil.ReadFile(filepath.Join(wd, "pkgs")); err == nil {
			packages = append(packages, strings.Split(string(data), "\n")...)
		}
		cmd.Printf("Packages: %s\n", strings.Join(packages, ", "))

		loader, errs := ast.NewLoader(wd, os.Environ(), packages)
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err)
			}
			os.Exit(1)
		}
		cfg, err := swipe.GetConfig(loader)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
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
				cmd.PrintErrln(err)
			}
			success = false
		}

		diffExcludes := make([]string, 0, len(result))

		for _, g := range result {
			if len(g.Errs) > 0 {
				for _, err := range g.Errs {
					cmd.PrintErrln(err)
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
				cmd.PrintErrf("%s: failed to create dir %s: %v\n", g.PkgPath, dirPath, err)
				os.Exit(1)
			}
			err := ioutil.WriteFile(g.OutputPath, g.Content, 0755)
			if err == nil {
				if verbose {
					log.Printf("%s: wrote %s\n", g.PkgPath, g.OutputPath)
				}
			} else {
				log.Printf("%s: failed to write %s: %v\n", g.PkgPath, g.OutputPath, err)
				success = false
			}
		}
		if !success {
			os.Exit(1)
		}
		if err := gitattributes.Generate(cfg.WorkDir, diffExcludes); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		cmd.Println("Command execution completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("work-dir", "w", "", "Workdir")
	genCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}
