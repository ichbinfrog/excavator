package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	viper.AutomaticEnv()

	flags := rootCmd.PersistentFlags()
	flags.IntVarP(&verbosity, "verbosity", "v", 3, "logging verbosity (0: Fatal, 1: Error, 2: Warning, 3: Info, 4: Debug, 5: Trace)")
}
