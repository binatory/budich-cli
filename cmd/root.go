package cmd

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"io.github.binatory/nhac-cli/internal/domain"
)

var app *domain.App

var rootCmd = &cobra.Command{
	Use:           "busich",
	Short:         "busich is a TUI and CLI music player for vietnamese. For more info: https://github.com/binatory/busich-cli",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// f, err := os.OpenFile("/tmp/logg", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0744)
		// if err != nil {
		// 	panic(err)
		// }
		// defer f.Close()

		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339})

		app = domain.NewApp(domain.NewConnectorZingMp3(http.DefaultClient), os.Stdout)
		return app.Init()
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search for songs/playlists/artists by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Search(args[0])
	},
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "play a song by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Play(args[0])
	},
}

func init() {
	searchCmd.Flags().StringP("type", "t", "", "type (required)")
	searchCmd.MarkFlagRequired("type")

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(playCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
}
