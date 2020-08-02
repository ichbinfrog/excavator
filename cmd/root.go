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

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	flags := rootCmd.PersistentFlags()
	flags.IntVarP(&verbosity, "verbosity", "v", 3, "logging verbosity (0: Fatal, 1: Error, 2: Warning, 3: Info, 4: Debug, 5: Trace)")
	flags.StringVarP(&rules, "rules", "r", "", "location of the rule declaration (defaults to internal)")
	flags.StringVarP(&format, "format", "f", "html", "output format of the scan results")
	flags.IntVarP(&concurrent, "concurrent", "c", 1, "number of concurrent executions (any number below 0 is considered as a single routine execution)")
}
