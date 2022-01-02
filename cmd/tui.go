package cmd

import (
	"github.com/spf13/cobra"
	"io.github.binatory/busich-cli/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the text-based user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.New(app).Start()

	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
