package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	path, rules, format string
	concurrent          int
)

var rootCmd = &cobra.Command{
	Use:   "excavator",
	Short: "small cli to scan a git repository for potential leaks",
}

// Execute attempts to run the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

var (
	dbhost, dbuser, dbpasswd, dbname, dbsslmode string
	dbport, verbosity                           int
	nobackend                                   bool
)

func setVerbosity() {
	switch verbosity {
	case 0:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case 1:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		break
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	flags := rootCmd.PersistentFlags()
	flags.CountVarP(&verbosity, "verbosity", "v", "logging verbosity (default : warning)")
	flags.StringVarP(&rules, "rules", "r", "", "location of the rule declaration (defaults to internal)")
	flags.StringVarP(&format, "format", "f", "html", "output format of the scan results")
	flags.IntVarP(&concurrent, "concurrent", "c", 1, "number of concurrent executions (any number below 0 is considered as a single routine execution)")
}
