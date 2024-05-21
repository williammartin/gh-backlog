package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	workwork "github.com/williammartin/gh-workwork"
)

func main() {
	p := tea.NewProgram(workwork.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
