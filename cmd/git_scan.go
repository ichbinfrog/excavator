package cmd

import (
	"github.com/ichbinfrog/excavator/pkg/scan"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var gitScanCmd = &cobra.Command{
	Use:   "git",
	Short: "scan a git repository",
	Long: `Command to scan a local or remote git repository.
Will loop through each commit to verify for possible password,
access tokens (JWT, aws, gcp, ...) leaks.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setVerbosity()
		log.Debug().
			Str("path", path).
			Str("repo", args[0]).
			Str("rules", rules).
			Str("format", format).
			Int("concurrent", concurrent).
			Msg("Scan initiated with configuration")

		var s *scan.GitScanner
		if format == "yaml" {
			s = scan.NewGitScanner(args[0], path, rules, &scan.YamlReport{}, true)
		} else {
			s = scan.NewGitScanner(args[0], path, rules, &scan.HTMLReport{}, true)
		}
		s.Scan(concurrent)
	},
}

func init() {
	rootCmd.AddCommand(gitScanCmd)
	gitScanCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "temporary local path to store the git repository (only applies to remote repository)")
}
