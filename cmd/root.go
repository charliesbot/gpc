package cmd

import (
	"fmt"
	"os"

	"github.com/charliesbot/gpc/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagPackage string
	flagKeyFile string
	flagOutput  string
	flagQuiet   bool
	flagVerbose bool
	appVersion  string
)

func SetVersion(v string) {
	appVersion = v
}

var rootCmd = &cobra.Command{
	Use:   "gpc",
	Short: "Google Play Console CLI",
	Long:  "A command-line interface for the Google Play Developer API.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       "dev",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.Banner(appVersion))
		cmd.Help()
	},
}

func Execute() error {
	rootCmd.Version = appVersion
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagPackage, "package", "p", "", "Android package name (e.g. com.example.app)")
	rootCmd.PersistentFlags().StringVar(&flagKeyFile, "key-file", "", "path to service account JSON key file")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "json", "output format: json or table")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "enable verbose output")
}

func getPackage() (string, error) {
	if flagPackage != "" {
		return flagPackage, nil
	}
	if pkg := os.Getenv("GPC_PACKAGE"); pkg != "" {
		return pkg, nil
	}
	return "", fmt.Errorf("package name required: use --package flag or set GPC_PACKAGE env var")
}

func getFormatter() string {
	return flagOutput
}
