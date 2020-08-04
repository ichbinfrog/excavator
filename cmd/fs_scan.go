package cmd

import (
	"github.com/ichbinfrog/excavator/pkg/scan"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var fsScanCmd = &cobra.Command{
	Use:   "fs",
	Short: "scan a directory in the filesystem",
	Long: `Command to scan a local directory in the filesystem.
Will loop through each file to verify for possible password,
access tokens (JWT, aws, gcp, ...) leaks.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setVerbosity()
		log.Debug().
			Str("repo", args[0]).
			Str("rules", rules).
			Str("format", format).
			Int("concurrent", concurrent).
			Msg("Scan initiated with configuration")

		var s *scan.FsScanner
		if format == "yaml" {
			s = scan.NewFsScanner(args[0], rules, &scan.YamlReport{}, true)
		} else {
			s = scan.NewFsScanner(args[0], rules, &scan.HTMLReport{}, true)
		}
		s.Scan(concurrent)
	},
}

func init() {
	rootCmd.AddCommand(fsScanCmd)
}
