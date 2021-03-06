package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io.github.binatory/budich-cli/internal/cli"
	"io.github.binatory/budich-cli/metadata"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"io.github.binatory/budich-cli/internal/domain"
)

var (
	// root flags
	verboseFlag bool

	app      domain.App
	executor *cli.CLI
)

var rootCmd = &cobra.Command{
	Use:           "budich",
	Short:         "budich is a TUI and CLI music player for vietnamese. For more info: https://github.com/binatory/budich-cli",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       metadata.Version.String(),
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

		// check for updates
		updateStatus, err := app.CheckForUpdate()
		if err != nil {
			log.Error().Msgf("Unable to check for updates: %s", err)
		} else if !updateStatus.IsUpToDate {
			fmt.Fprintf(os.Stderr, "New release has been found, current version %s, latest version %s (%s)",
				metadata.VersionRaw, updateStatus.LatestVersion, updateStatus.LatestVersionUrl)
			fmt.Fprintln(os.Stderr)
		}

		// init core
		if err := app.Init(); err != nil {
			panic(err)
		}

		// setup CLI implementation
		executor = cli.New(os.Stdout, app)
	})

	// setup rootCmd
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable logs verbosity")
}

func Execute() error {
	defer func() {
		if err := recover(); err != nil {
			log.Panic().Msgf("%+v", err)
		}
	}()

	err := rootCmd.Execute()
	if err != nil {
		log.Error().Msgf("%+v", err)
	}
	return err
}
