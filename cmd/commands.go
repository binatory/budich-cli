package cmd

import "github.com/spf13/cobra"

var (
	// search cmd flags
	connectorFlag string
)

var searchCmd = &cobra.Command{
	Use:   "search <search_term>",
	Short: "search for songs/playlists/artists by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executor.Search(connectorFlag, args[0])
	},
}

var playCmd = &cobra.Command{
	Use:   "play <song_id>",
	Short: "play a song by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executor.Play(args[0])
	},
}

func init() {
	// setup searchCmd
	searchCmd.Flags().StringVarP(&connectorFlag, "connector", "c", "", "connector name (required)")
	searchCmd.MarkFlagRequired("connector")

	// add sub commands to root
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(playCmd)
}
