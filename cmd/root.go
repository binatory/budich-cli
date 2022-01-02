package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"

	"io.github.binatory/busich-cli/internal/domain"
)

var (
	// search cmd flags
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
		songs, err := app.Search(connectorFlag, args[0])
		if err != nil {
			return err
		}

		tw := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		defer tw.Flush()

		fmt.Fprint(tw, "Id\tBài hát\tCa sĩ\n")
		fmt.Fprint(tw, "----------\t----------\t----------\n")
		for _, s := range songs {
			fmt.Fprintf(tw, "%s.%s\t%s\t%s\n", connectorFlag, s.Id, s.Name, s.Artists)
		}

		return nil
	},
}

var playCmd = &cobra.Command{
	Use:   "play <song_id>",
	Short: "play a song by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]
		parts := strings.SplitN(input, ".", 2)
		if len(parts) != 2 {
			return errors.Errorf("invalid id %s", input)
		}
		cName, id := parts[0], parts[1]
		player, err := app.Play2(domain.Song{
			Id:        id,
			Connector: cName,
		})
		if err != nil {
			return errors.Errorf("error getting player: %s", err)
		}

		go func() {
			for {
				select {
				case <-time.After(time.Second):
					report := player.Report()
					fmt.Fprintf(os.Stdout, "Current state (%s): %s/%s\n", report.State, report.Pos, report.Len)
				}
			}
		}()

		return player.Start()
	},
}

func init() {
	cobra.OnInitialize(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339})

		if _, ok := os.LookupEnv("BD_PROFILING"); ok {
			runtime.SetBlockProfileRate(1)
			runtime.SetMutexProfileFraction(10)

			go func() {
				if err := http.ListenAndServe("localhost:30888", nil); err != nil {
					panic(err)
				}
			}()
		}
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
