package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"simple/client"
	"simple/config"
)

// App represents the terminal UI application
type App struct {
	config *config.Config
	client *client.PlainClient
}

// AppState represents the current application state
type AppState int

const (
	StateThreads AppState = iota
	StateDashboard
)

// MainModel represents the main application model
type MainModel struct {
	config        *config.Config
	client        *client.PlainClient
	state         AppState
	threadsView   *ThreadsView
	dashboardView *DashboardView
	quitting      bool
	width         int
	height        int
}

// NewApp creates a new terminal UI application
func NewApp(cfg *config.Config) (*App, error) {
	client := client.NewPlainClient(cfg)

	return &App{
		config: cfg,
		client: client,
	}, nil
}

// NewMainModel creates a new main model
func NewMainModel(cfg *config.Config, client *client.PlainClient) *MainModel {
	threadsView := NewThreadsView(cfg, client)
	dashboardView := NewDashboardView(cfg, client)

	return &MainModel{
		config:        cfg,
		client:        client,
		state:         StateThreads,
		threadsView:   threadsView,
		dashboardView: dashboardView,
		quitting:      false,
	}
}

// Init initializes the main model
func (m *MainModel) Init() tea.Cmd {
	return m.threadsView.Init()
}

// Update handles messages and updates the model
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update sub-models with new dimensions
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.threadsView, cmd = m.threadsView.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.dashboardView, cmd = m.dashboardView.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == StateThreads && m.threadsView.IsInDetailView() {
				// If in detail view, go back to list
				var cmd tea.Cmd
				m.threadsView, cmd = m.threadsView.Update(msg)
				return m, cmd
			} else if m.state == StateDashboard {
				// Go back to threads view from dashboard
				m.state = StateThreads
				return m, nil
			} else {
				// Otherwise quit the application
				m.quitting = true
				return m, tea.Quit
			}
		case "d":
			// Switch to dashboard if not already there
			if m.state != StateDashboard {
				m.state = StateDashboard
				return m, m.dashboardView.Init()
			}
		case "t":
			// Switch to threads view if not already there
			if m.state != StateThreads {
				m.state = StateThreads
				return m, nil
			}
		}

		// Handle messages for the current view
		switch m.state {
		case StateThreads:
			var cmd tea.Cmd
			m.threadsView, cmd = m.threadsView.Update(msg)
			return m, cmd
		case StateDashboard:
			var cmd tea.Cmd
			m.dashboardView, cmd = m.dashboardView.Update(msg)
			return m, cmd
		}
	}

	// Handle messages for the current view
	switch m.state {
	case StateThreads:
		var cmd tea.Cmd
		m.threadsView, cmd = m.threadsView.Update(msg)
		return m, cmd
	case StateDashboard:
		var cmd tea.Cmd
		m.dashboardView, cmd = m.dashboardView.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the main application view
func (m *MainModel) View() string {
	if m.quitting {
		return ""
	}

	var content string

	switch m.state {
	case StateThreads:
		content = m.threadsView.View()
	case StateDashboard:
		content = m.dashboardView.View()
	}

	// Add some styling
	style := lipgloss.NewStyle().
		Padding(1).
		Width(m.width).
		Height(m.height)

	return style.Render(content)
}
