package workwork

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williammartin/gh-workwork/remotedata"
)

type Item struct {
	Name string
}

type Column struct {
	Name  string
	Items []Item
}

type Board struct {
	Columns []Column
}

func loadBoard() tea.Msg {
	if os.Getenv("FORCE_BOARD_ERROR") != "" {
		return loadBoardResult{
			Err: fmt.Errorf("forced error"),
		}
	}

	return loadBoardResult{
		Err: nil,
		Board: Board{
			Columns: []Column{
				{
					Name: "Prioritised",
					Items: []Item{
						{Name: "Foo"},
					},
				},
				{
					Name: "In Progress",
					Items: []Item{
						{Name: "Bar"},
					},
				},
				{
					Name: "Done",
					Items: []Item{
						{Name: "Baz"},
					},
				},
			},
		},
	}

}

type loadBoardResult struct {
	Err   error
	Board Board
}

func InitialModel() Model {
	return Model{
		State: remotedata.NotAsked{},
	}
}

type Model struct {
	remotedata.State[Board]
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return loadBoard
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case loadBoardResult:
		if msg.Err != nil {
			return Model{
				State: remotedata.Failure{Error: msg.Err},
			}, nil
		}

		return Model{
			State: remotedata.Success[Board]{Data: msg.Board},
		}, nil
	}

	return m, nil
}

func (m Model) View() string {
	s, _ := remotedata.Match(m.State,
		func(remotedata.NotAsked) (string, error) {
			return "Not asked", nil
		},
		func(remotedata.Loading) (string, error) {
			return "Loading", nil
		},
		func(f remotedata.Failure) (string, error) {
			return fmt.Sprintf("Failed: %v", f.Error), nil
		},
		func(s remotedata.Success[Board]) (string, error) {
			return fmt.Sprintf("Success: %v", s.Data), nil
		},
	)

	return s
}
