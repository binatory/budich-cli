package cli

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io.github.binatory/budich-cli/internal/domain"
	"strings"
	"text/tabwriter"
	"time"
)

type CLI struct {
	out            io.Writer
	app            domain.App
	reportInterval time.Duration
}

func New(out io.Writer, app domain.App) *CLI {
	return &CLI{out, app, time.Second}
}

func (c *CLI) Search(connector, term string) error {
	songs, err := c.app.Search(connector, term)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(c.out, 1, 1, 5, ' ', 0)
	defer tw.Flush()

	fmt.Fprint(tw, "Id\tBài hát\tCa sĩ")
	fmt.Fprintln(tw)
	fmt.Fprint(tw, "----------\t----------\t----------")
	fmt.Fprintln(tw)
	for _, s := range songs {
		fmt.Fprintf(tw, "%s.%s\t%s\t%s", connector, s.Id, s.Name, s.Artists)
		fmt.Fprintln(tw)
	}

	return nil
}

func (c *CLI) Play(input string) error {
	parts := strings.SplitN(input, ".", 2)
	if len(parts) != 2 {
		return errors.Errorf("invalid id %s", input)
	}
	cName, id := parts[0], parts[1]
	player, err := c.app.Play(id, cName)
	if err != nil {
		return errors.Errorf("error getting player: %s", err)
	}

	song := player.Report().Song
	fmt.Fprintf(c.out, "Playing %s (%s), duration %s", song.Name, song.Artists, song.Duration)
	fmt.Fprintln(c.out)

	go func() {
		isLoading := false

		for {
			select {
			case <-time.After(c.reportInterval):
				report := player.Report()
				switch report.State {
				case domain.StateNotInitialized:
					fallthrough
				case domain.StateLoading:
					if !isLoading {
						isLoading = true
						fmt.Fprintln(c.out, "Loading...")
					}
				case domain.StatePlaying:
					fmt.Fprintf(c.out, "Playing: %s/%s", report.Pos, report.Len)
					fmt.Fprintln(c.out)
				case domain.StatePaused:
					fmt.Fprintf(c.out, "Paused: %s/%s", report.Pos, report.Len)
					fmt.Fprintln(c.out)
				default:
					return
				}
			}
		}
	}()

	return player.Start()
}
