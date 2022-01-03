package cmd

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	logFile     *os.File
	logFullPath string
)

func initLogger(verbose bool) {
	// create log dir if not exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	logDir := filepath.Join(homeDir, ".budich-cli")
	logFilename := "output.log"
	logFullPath = filepath.Join(logDir, logFilename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: time.RFC3339})
	if verbose {
		log.Level(zerolog.DebugLevel)
	} else {
		log.Level(zerolog.InfoLevel)
	}
}

func switchLogsOutput(toFile bool) {
	var out io.Writer
	if toFile {
		ensureLogFile()
		out = logFile
	} else {
		out = os.Stderr
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: out, NoColor: true, TimeFormat: time.RFC3339})
}

func ensureLogFile() {
	if logFile != nil {
		return
	}

	var err error
	logFile, err = os.OpenFile(logFullPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}
