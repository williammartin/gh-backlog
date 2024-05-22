package main

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	workwork "github.com/williammartin/gh-workwork"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: workwork <owner> <project number>")
		os.Exit(1)
	}

	owner := os.Args[1]
	projectNumber, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "project number must be a number")
		os.Exit(1)
	}

	p := tea.NewProgram(workwork.InitialModel(owner, int32(projectNumber)))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
