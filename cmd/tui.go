package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"simple/client"
	"simple/config"
	"simple/ui"
)

// TUICmd represents the TUI command
type TUICmd struct{}

// Run executes the TUI command
func (t *TUICmd) Run(cfg *config.Config) error {
	// Set up debug logging if needed
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	// Create Plain API client
	client := client.NewPlainClient(cfg)

	// Create the main model
	model := ui.NewMainModel(cfg, client)

	// Create and run the program
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
