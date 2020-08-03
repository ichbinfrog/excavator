package cmd

import (
	"github.com/ichbinfrog/excavator/pkg/scan"
	"github.com/rs/zerolog"
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
		switch verbosity {
		case 0:
			zerolog.SetGlobalLevel(zerolog.FatalLevel)
			break
		case 1:
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			break
		case 3:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			break
		case 4:
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			break
		case 5:
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
			break
		default:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		}
		log.Debug().
			Str("path", path).
			Str("repo", args[0]).
			Str("rules", rules).
			Str("format", format).
			Int("concurrent", concurrent).
			Msg("Scan initiated with configuration")

		s := &scan.GitScanner{}

		if format == "yaml" {
			s.New(args[0], path, rules, &scan.YamlReport{}, true)
		} else {
			s.New(args[0], path, rules, &scan.HTMLReport{}, true)
		}
		s.Scan(concurrent)
	},
}

func init() {
	rootCmd.AddCommand(gitScanCmd)
	gitScanCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "temporary local path to store the git repository (only applies to remote repository)")
}
