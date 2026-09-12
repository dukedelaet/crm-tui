package main

import (
	"fmt"
	"os"

	"github.com/dukedelaet/sophie/model"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model.NewModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
