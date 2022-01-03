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

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"io.github.binatory/busich-cli/internal/domain"
)

var (
	// root flags
	verboseFlag bool

	// search cmd flags
	connectorFlag string

	app *domain.App
)

var rootCmd = &cobra.Command{
	Use:          "busich",
	Short:        "busich is a TUI and CLI music player for vietnamese. For more info: https://github.com/binatory/busich-cli",
	SilenceUsage: true,
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
		player, err := app.Play(domain.Song{
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
		// setup logger
		initLogger(verboseFlag)

		// setup profiler
		if _, ok := os.LookupEnv("BD_PROFILING"); ok {
			pprofAddr := "localhost:30888"
			runtime.SetBlockProfileRate(1)
			runtime.SetMutexProfileFraction(10)
			log.Info().Msgf("Launch pprof server at http://%s", pprofAddr)

			go func() {
				if err := http.ListenAndServe(pprofAddr, nil); err != nil {
					log.Panic().Stack().Err(err).Send()
				}
			}()
		}

		// setup core
		app = domain.DefaultApp()
		if err := app.Init(); err != nil {
			panic(err)
		}
	})

	// setup searchCmd
	searchCmd.Flags().StringVarP(&connectorFlag, "connector", "c", "", "connector name (required)")
	searchCmd.MarkFlagRequired("connector")

	// setup rootCmd
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(playCmd)
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable logs verbosity")
}

func Execute() error {
	defer func() {
		if err := recover(); err != nil {
			log.Panic().Msgf("%+v", err)
		}
	}()

	return rootCmd.Execute()
}
