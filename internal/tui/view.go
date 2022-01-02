package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io.github.binatory/busich-cli/internal/domain"
	"io.github.binatory/busich-cli/internal/tui/model"
)

type view struct {
	model           *model.Model
	onSelectSong    func(domain.Song)
	onSwitchPage    func(model.PageEnum)
	onPauseOrResume func()
	onSearch        func()

	// ui components
	appView        *tview.Application
	playerView     *tview.TextView
	searchFormView *tview.Form
	pagesView      *tview.Pages
	songsListView  *tview.Table
}

func NewView(m *model.Model, onSelectSong func(domain.Song), onSwitchPage func(enum model.PageEnum), onPauseOrResume func(), onSearch func()) *view {
	return &view{
		model:           m,
		onSelectSong:    onSelectSong,
		onSwitchPage:    onSwitchPage,
		onPauseOrResume: onPauseOrResume,
		onSearch:        onSearch,
	}
}

func (v *view) StartView() error {
	v.songsListView = tview.NewTable().SetBorders(false).SetSelectable(true, false)
	v.songsListView.SetSelectedFunc(func(row, _ int) {
		song := v.songsListView.GetCell(row, 0).GetReference().(domain.Song)
		go v.onSelectSong(song)
	})

	v.playerView = tview.NewTextView().SetTextAlign(tview.AlignCenter)

	v.searchFormView = tview.NewForm()

	v.pagesView = tview.NewPages()
	v.pagesView.AddPage(model.PageSearch.String(), v.searchFormView, true, true)
	v.pagesView.AddPage(model.PageList.String(), v.songsListView, true, false)

	grid := tview.NewGrid().
		SetRows(0, 3).
		SetBorders(true).
		AddItem(v.pagesView, 0, 0, 1, 1, 0, 0, false).
		AddItem(v.playerView, 1, 0, 1, 1, 0, 0, false)

	v.appView = tview.NewApplication()
	v.appView.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyF4:
			go v.onSwitchPage(model.PageSearch)
			return nil
		case tcell.KeyF9:
			go v.onPauseOrResume()
			return nil
			//case tcell.KeyRune:
			//	switch ev.Rune() {
			//	case 'p':
			//		go v.onPauseOrResume()
			//		return nil
			//	}
		}
		return ev
	})

	v.appView.SetRoot(grid, true)
	v.updateViews()

	return v.appView.Run()
}

func (v *view) updaters() []func(bool) {
	return []func(bool){
		v.updateSearchFormView,
		v.updatePlayerView,
		v.switchPage,
		v.updateSongsListView,
	}
}

func (v *view) updateViews() {
	for _, f := range v.updaters() {
		f(false)
	}
}

func (v *view) updateViewsAsync() {
	v.executeUpdate(true, v.updateViews)
}

func (v *view) updateSearchFormView(async bool) {
	if v.model.CurrentPage != model.PageSearch {
		return
	}

	v.executeUpdate(async, func() {
		v.searchFormView.Clear(true)
		v.searchFormView.AddDropDown("Type", []string{"Song", "Artist", "Playlist"}, 0, func(option string, _ int) {
			v.model.Search.SelectedType = option
		})
		v.searchFormView.AddDropDown("Connector", v.model.Search.ConnectorNames, 0, func(option string, _ int) {
			v.model.Search.SelectedConnector = option
		})
		v.searchFormView.AddInputField("Term", "", 20, nil, func(term string) {
			v.model.Search.Term = term
		})
		v.searchFormView.AddButton("Search", func() {
			go v.onSearch()
		})
		v.searchFormView.AddButton("Reset", func() {
			v.appView.QueueUpdate(func() {
				v.searchFormView.Clear(true)
			})
		})
	})
}

func (v *view) executeUpdate(async bool, f func()) {
	if async {
		v.appView.QueueUpdateDraw(f)
	} else {
		f()
	}
}

func (v *view) updatePlayerView(async bool) {
	v.executeUpdate(async, func() {
		if v.model.Player.IsInitialized {
			playerModel := &v.model.Player
			v.playerView.SetText(fmt.Sprintf("%s - %s\nCurrent state (%s): %s/%s",
				playerModel.SongName, playerModel.ArtistsName, playerModel.Status.State, playerModel.Status.Pos, playerModel.Status.Len))
		} else {
			v.playerView.SetText("N/A")
		}
	})
}

func (v *view) updateSongsListView(async bool) {
	if v.model.CurrentPage != model.PageList {
		return
	}

	v.executeUpdate(async, func() {
		// clear
		v.songsListView.Clear()

		// set headers
		headers := []string{"Id", "Name", "Artists", "Duration"}
		v.songsListView.SetFixed(1, len(headers))
		for col, header := range headers {
			v.songsListView.SetCell(0, col, tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter).SetSelectable(false))
		}

		// set data
		for row, song := range v.model.SongsList {
			for col, text := range []string{song.Id, song.Name, song.Artists, song.Duration.String()} {
				cell := tview.NewTableCell(text).SetTextColor(tcell.ColorWhite)
				if col == 0 {
					cell.SetReference(song)
				}
				v.songsListView.SetCell(row+1, col, cell)
			}
		}
	})
}

func (v *view) switchPage(async bool) {
	v.executeUpdate(async, func() {
		v.pagesView.SwitchToPage(v.model.CurrentPage.String())
		v.appView.SetFocus(v.pagesView)
	})
}
