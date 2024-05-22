package workwork

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/cli/v2/pkg/cmd/project/shared/queries"
	"github.com/cli/cli/v2/pkg/iostreams"
	ghAPI "github.com/cli/go-gh/v2/pkg/api"
	"github.com/williammartin/gh-workwork/list"
	"github.com/williammartin/gh-workwork/maps"
	"github.com/williammartin/gh-workwork/remotedata"
	"github.com/williammartin/gh-workwork/slices"
)

type Item struct {
	id   uint
	repo string
	name string
}

func (i Item) String() string {
	return fmt.Sprintf("%s #%d\n\n%s", i.repo, i.id, i.name)
}

type Column struct {
	Name  string
	Items []Item
}

func viewColumn(c Column, width int) string {
	title := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).
		MarginBottom(1)

	enumeratorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		MarginRight(1)

	itemStyle := lipgloss.NewStyle().
		Padding(1, 1).
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Foreground(lipgloss.Color("255")).
		MarginRight(1)

	l := list.New(slices.Map(c.Items, func(i Item) list.Item { return i })...).
		Enumerator(list.None).
		EnumeratorStyle(enumeratorStyle).
		ItemStyle(itemStyle)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(width).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			title.Render(c.Name),
			l.String(),
		))
}

type Board struct {
	Columns []Column
}

func viewBoard(b Board, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinHorizontal(
			lipgloss.Left,
			slices.Map(b.Columns, func(c Column) string {
				return viewColumn(c, width/len(b.Columns))
			})...,
		))
}

func loadBoard(owner string, projectNumber int32) tea.Cmd {
	return func() tea.Msg {
		// This is all copied and pasted from the CLI cause I ain't writing that from scratch
		ios := &iostreams.IOStreams{}
		ios.SetStdoutTTY(false)
		ios.SetStderrTTY(false)
		ios.SetStdinTTY(false)

		ghClient, err := ghAPI.NewHTTPClient(ghAPI.ClientOptions{})
		if err != nil {
			return loadBoardResult{
				Err: fmt.Errorf("could not create HTTP client: %v", err),
			}
		}

		client := queries.NewClient(ghClient, "github.com", ios)
		projectOwner, err := client.NewOwner(false, owner)
		if err != nil {
			return loadBoardResult{
				Err: fmt.Errorf("could not work with the provided owner %q: %v", owner, err),
			}
		}

		project, err := client.ProjectItems(projectOwner, projectNumber, 500)
		if err != nil {
			return loadBoardResult{
				Err: fmt.Errorf("could not work with the provided project %d: %v", projectNumber, err),
			}
		}

		// Ok now let's group each item into columns
		buckets := map[string]Column{}

		// Iterate over all the items
		for _, node := range project.Items.Nodes {
			// Iterate over all the fields of those items
			for _, fields := range node.FieldValues.Nodes {
				// If the field is a SingeSelect (i.e. a dropdown) and has the name "Status"
				if fields.ProjectV2ItemFieldSingleSelectValue.Field.Name() == "Status" {
					// We get the value of that Status
					status := fields.ProjectV2ItemFieldSingleSelectValue.Name
					// And we bucket this item into the column with the same name as the status
					var column Column
					column, ok := buckets[status]
					// If there was no existing columnn we create it
					if !ok {
						column = Column{Name: status}
					}
					// We append the new item to the column
					column.Items = append(column.Items, Item{
						id:   uint(node.Number()),
						repo: node.Repo(),
						name: node.Title(),
					})

					// And since we're working on values rather than pointers, just set the column back in the map
					buckets[status] = column
				}
			}
		}

		return loadBoardResult{
			Err: nil,
			Board: Board{
				Columns: slices.Sort(maps.Values(buckets), byStatus),
			},
		}
	}
}

// lol wow this is a terrible sorting function
func byStatus(i, j Column) int {
	if i.Name == "Prioritized" {
		return -1
	}

	if i.Name == "Done" {
		return 1
	}

	if j.Name == "Prioritized" {
		return 1
	}

	return -1
}

type loadBoardResult struct {
	Err   error
	Board Board
}

func InitialModel(owner string, projectNumber int32) Model {
	return Model{
		Owner:         owner,
		ProjectNumber: projectNumber,

		Board: remotedata.NotAsked{},
		Width: 0,
	}
}

type Model struct {
	Owner         string
	ProjectNumber int32

	Board remotedata.State[Board]
	Width int
}

func (m Model) Init() tea.Cmd {
	return loadBoard(m.Owner, m.ProjectNumber)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		return Model{
			Board: m.Board,
			Width: msg.Width,
		}, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case loadBoardResult:
		if msg.Err != nil {
			return Model{
				Board: remotedata.Failure{Err: msg.Err},
			}, nil
		}

		return Model{
			Board: remotedata.Success[Board]{Data: msg.Board},
		}, nil
	}

	return m, nil
}

func (m Model) View() string {
	s := remotedata.Match(m.Board,
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
			return viewBoard(s.Data, m.Width)
		},
	)

	return s
}
