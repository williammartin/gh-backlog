package workwork

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williammartin/gh-workwork/remotedata"
	"github.com/williammartin/gh-workwork/slices"
)

type Item struct {
	id   uint
	repo string
	name string
}

// implement the list.Item interface.
func (i Item) FilterValue() string {
	return i.name
}

func (i Item) Title() string {
	return fmt.Sprintf("%s #%d", i.repo, i.id)
}

func (i Item) Description() string {
	return i.name
}

type Column struct {
	Name  string
	Items []Item

	List list.Model
}

func (c Column) View() string {
	// TODO: I think bubbletea kind of expects messages to be sent to nested models, so
	// for example, list would be a field on Column. For the moment I'll just construct
	// the list each time though.

	// TODO: handle window resizing
	list := list.New(
		slices.Map(c.Items, func(i Item) list.Item { return i }),
		list.NewDefaultDelegate(), 500, 5)
	list.Title = c.Name
	list.SetShowHelp(false)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Render(list.View())
}

type Board struct {
	Columns []Column
}

func (b Board) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		slices.Map(b.Columns, Column.View)...,
	)
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
						{id: 354, repo: "cli", name: "GitHub Project Experiment"},
						{id: 339, repo: "cli", name: "Assess GitHub CLI primer design for screen reader accessibility"},
					},
				},
				{
					Name: "In Progress",
					Items: []Item{
						{id: 393, repo: "cli", name: "Initial FAQ Documentation"},
					},
				},
				{
					Name: "Done",
					Items: []Item{
						{id: 402, repo: "cli", name: "Document proposed retrospective experiment"},
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
				State: remotedata.Failure{Err: msg.Err},
			}, nil
		}

		return Model{
			State: remotedata.Success[Board]{Data: msg.Board},
		}, nil
	}

	return m, nil
}

func (m Model) View() string {
	s := remotedata.Match(m.State,
		func(remotedata.NotAsked) string {
			return ""
		},
		func(remotedata.Loading) string {
			return "Loading..."
		},
		func(f remotedata.Failure) string {
			return fmt.Sprintf("Uh oh ya done goofed because: %v", f.Err)
		},
		func(s remotedata.Success[Board]) string {
			return s.Data.View()
		},
	)

	return s
}
