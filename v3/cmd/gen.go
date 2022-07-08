package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/swipe-io/swipe/v3/frame"
	"github.com/swipe-io/swipe/v3/internal/ast"
	"github.com/swipe-io/swipe/v3/internal/gitattributes"
	"github.com/swipe-io/swipe/v3/swipe"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen [dir]",
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

		cmd.Println("Please wait the command is running, it may take some time\n")

		if len(packages) == 0 {
			packages = viper.GetStringSlice("packages")
		}

		prefix := viper.GetString("prefix")
		swipePkg := viper.GetString("swipe-pkg")
		verbose, _ := cmd.Flags().GetBool("verbose")
		wd := viper.GetString("work-dir")
		useDoNotEdit := viper.GetBool("dn-edit")

		if wd == "" {
			wd, _ = cmd.Flags().GetString("work-dir")
		}

		swipeSysFilepath := filepath.Join(wd, ".swipe")

		if wd == "" {
			wd, err = os.Getwd()
			if err != nil {
				cmd.PrintErrf("failed to get working directory: ", err)
				os.Exit(1)
			}
		}

		basePath := path.Dir(wd)

		cmd.Printf("Workdir: %s\n", wd)

		// clear all before generated files.
		data, err := ioutil.ReadFile(swipeSysFilepath)
		if err == nil {
			genOldFiles := strings.Split(string(data), "\n")

			for _, filepath := range genOldFiles {
				if err := os.Remove(basePath + filepath); err != nil {
					if verbose {
						cmd.Printf("Remove generated file %s error: %s", filepath, err)
					}
				}
			}
		}

		if data, err := ioutil.ReadFile(filepath.Join(wd, "pkgs")); err == nil {
			packages = append(packages, strings.Split(string(data), "\n")...)
		}
		cmd.Printf("Packages: %s\n", strings.Join(packages, ", "))
		cmd.Printf("Swipe Package: %s\n", swipePkg)

		cmd.Println()

		packages = append(packages, filepath.Join(wd, swipePkg, "swipe", "..."))
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

		result, errs := swipe.Generate(cfg, prefix)
		success := true
		if len(errs) > 0 {
			for _, err := range errs {
				cmd.PrintErrln(err)
			}
			success = false
		}

		diffExcludes := make([]string, 0, len(result))
		generatedFiles := bytes.NewBuffer(nil)

		if verbose {
			cmd.Println("Generated files")
		}

		var i int
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

			filename := filepath.Base(g.OutputPath)
			f := frame.NewFrame(cmd.Version, filename, g.Imports, g.PkgName, useDoNotEdit)
			data, err := f.Frame(g.Content)
			if err != nil {
				cmd.PrintErrf("%s: failed to write %s: %v\n", g.PkgPath, g.OutputPath, err)
				os.Exit(1)
			}

			outputPath := strings.Replace(g.OutputPath, basePath, "", -1)

			if i > 0 {
				generatedFiles.WriteString("\n")
			}
			generatedFiles.WriteString(outputPath)

			diffExcludes = append(diffExcludes, strings.Replace(g.OutputPath, cfg.WorkDir+"/", "", -1))

			dirPath := filepath.Dir(g.OutputPath)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				cmd.PrintErrf("%s: failed to create dir %s: %v\n", g.PkgPath, dirPath, err)
				os.Exit(1)
			}
			err = ioutil.WriteFile(g.OutputPath, data, 0755)
			if err == nil {
				if verbose {
					cmd.Printf("%s: wrote %s\n", g.PkgPath, g.OutputPath)
				}
			} else {
				cmd.PrintErrf("%s: failed to write %s: %v\n", g.PkgPath, g.OutputPath, err)
				success = false
			}
			i++
		}
		if !success {
			os.Exit(1)
		}
		if err := gitattributes.Generate(cfg.WorkDir, diffExcludes); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(swipeSysFilepath, generatedFiles.Bytes(), 0755); err != nil {
			cmd.PrintErrf("Failed to create system file: %s", err)
			os.Exit(1)
		}

		cmd.Println("\n\nCommand execution completed successfully.")
	},
}

func init() {
	genCmd.Flags().StringP("swipe-pkg", "p", "pkg", "Swipe package name")
	genCmd.Flags().StringP("work-dir", "w", "", "Work directory")
	genCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	genCmd.Flags().StringP("prefix", "x", "swipe_gen_", "Prefix for generated file names")
	genCmd.Flags().BoolP("dn-edit", "d", true, "Generate a 'DO NOT EDIT' warning")

	_ = viper.BindPFlag("swipe-pkg", genCmd.Flags().Lookup("swipe-pkg"))
	_ = viper.BindPFlag("work-dir", genCmd.Flags().Lookup("work-dir"))
	_ = viper.BindPFlag("verbose", genCmd.Flags().Lookup("verbose"))
	_ = viper.BindPFlag("prefix", genCmd.Flags().Lookup("prefix"))
	_ = viper.BindPFlag("dn-edit", genCmd.Flags().Lookup("dn-edit"))

	rootCmd.AddCommand(genCmd)
}
