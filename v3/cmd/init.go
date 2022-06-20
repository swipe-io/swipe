package cmd

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/swipe-io/swipe/v3/swipe"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a swipe config file",
	Long:  ``,
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

		pkgName := viper.GetString("pkg")

		if pkgName == "" {
			pkgName = "pkg"
		}

		cmd.Printf("Workdir: %s\n", wd)
		cmd.Printf("Package: %s\n", pkgName)

		for name, data := range swipe.Options() {
			buf := bytes.NewBuffer(nil)
			path := filepath.Join(wd, pkgName, "swipe", name)
			if err := os.MkdirAll(path, 0775); err != nil {
				cmd.PrintErrf("Error: %s", err)
				os.Exit(1)
			}

			buf.WriteString("package " + name + "\n\n")
			buf.Write(data)

			filename := filepath.Join(path, "swipe.go")

			if _, err := os.Stat(filename); err == nil {
				if err := os.Remove(filename); err != nil {
					cmd.PrintErrf("Error: %s", err)
					os.Exit(1)
				}
			} else if !os.IsNotExist(err) {
				cmd.PrintErrf("Error: %s", err)
				os.Exit(1)
			}
			if err := os.WriteFile(filename, buf.Bytes(), 0755); err != nil {
				cmd.PrintErrf("Error: %s", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("pkg", "p", "pkg", "Package name")
	initCmd.Flags().StringP("work-dir", "w", "", "Swipe work directory")

	_ = viper.BindPFlag("pkg", initCmd.Flags().Lookup("pkg"))
	_ = viper.BindPFlag("work-dir", initCmd.Flags().Lookup("work-dir"))
}
