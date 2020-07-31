package cmd

import (
	"github.com/ichbinfrog/excavator/pkg/scan"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var fsScanCmd = &cobra.Command{
	Use:   "fsScan",
	Short: "scan a directory in the filesystem",
	Long: `Command to scan a local directory in the filesystem.
Will loop through each file to verify for possible password,
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
			Str("repo", args[0]).
			Str("rules", rules).
			Str("format", format).
			Int("concurrent", concurrent).
			Msg("Scan initiated with configuration")

		s := &scan.FsScanner{}
		if format == "yaml" {
			s.New(args[0], rules, &scan.YamlReport{}, true)
		} else {
			s.New(args[0], rules, &scan.HTMLReport{}, true)
		}
		s.Scan(concurrent)
	},
}

func init() {
	rootCmd.AddCommand(fsScanCmd)
}
