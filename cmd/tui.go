package cmd

import (
	"github.com/spf13/cobra"
	"io.github.binatory/budich-cli/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the text-based user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		switchLogsOutput(true)
		return tui.New(app).Start()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
