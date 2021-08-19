package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"io.github.binatory/busich-cli/internal/domain"
)

var (
	connectorFlag string

	app *domain.App
)

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

		app = domain.DefaultApp()
		return app.Init()
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <search_term>",
	Short: "search for songs/playlists/artists by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Search(connectorFlag, args[0])
	},
}

var playCmd = &cobra.Command{
	Use:   "play <song_id>",
	Short: "play a song by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Play(args[0])
	},
}

func init() {
	cobra.OnInitialize(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339})
	})

	searchCmd.Flags().StringVarP(&connectorFlag, "connector", "c", "", "connector name (required)")
	searchCmd.MarkFlagRequired("connector")

	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(playCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
}