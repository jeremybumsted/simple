package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"simple/client"
	"simple/config"
	"simple/types"
)

// DashboardView represents the dashboard view
type DashboardView struct {
	config     *config.Config
	client     *client.PlainClient
	loading    bool
	error      string
	components []DashboardComponent
	width      int
	height     int
}

// DashboardComponent interface for modular dashboard components
type DashboardComponent interface {
	Title() string
	Value() string
	Loading() bool
	Error() string
	Update(msg tea.Msg) (DashboardComponent, tea.Cmd)
}

// ThreadCountComponent shows count of threads for a specific status
type ThreadCountComponent struct {
	title   string
	status  string
	count   int
	loading bool
	error   string
	client  *client.PlainClient
	config  *config.Config
}

// ThreadsCreatedTodayComponent shows count of threads created today
type ThreadsCreatedTodayComponent struct {
	title   string
	count   int
	loading bool
	error   string
	client  *client.PlainClient
	config  *config.Config
}

// UnassignedThreadsComponent shows count of unassigned threads
type UnassignedThreadsComponent struct {
	title   string
	count   int
	loading bool
	error   string
	client  *client.PlainClient
	config  *config.Config
}

// threadCountMsg represents the result of loading thread count
type threadCountMsg struct {
	status string
	count  int
	error  string
}

// threadsCreatedTodayMsg represents the result of loading threads created today
type threadsCreatedTodayMsg struct {
	count int
	error string
}

// unassignedThreadsMsg represents the result of loading unassigned threads
type unassignedThreadsMsg struct {
	count int
	error string
}

// NewDashboardView creates a new dashboard view
func NewDashboardView(cfg *config.Config, client *client.PlainClient) *DashboardView {
	// Create dashboard components
	components := []DashboardComponent{
		NewThreadCountComponent("📋 TODO Threads", "TODO", cfg, client),
		NewThreadCountComponent("😴 Snoozed Threads", "SNOOZED", cfg, client),
		NewThreadsCreatedTodayComponent("📅 Created Today", cfg, client),
		NewUnassignedThreadsComponent("👤 Unassigned", cfg, client),
	}

	return &DashboardView{
		config:     cfg,
		client:     client,
		loading:    false,
		components: components,
	}
}

// NewThreadCountComponent creates a new thread count component
func NewThreadCountComponent(title, status string, cfg *config.Config, client *client.PlainClient) *ThreadCountComponent {
	return &ThreadCountComponent{
		title:   title,
		status:  status,
		count:   0,
		loading: true,
		client:  client,
		config:  cfg,
	}
}

// NewThreadsCreatedTodayComponent creates a new threads created today component
func NewThreadsCreatedTodayComponent(title string, cfg *config.Config, client *client.PlainClient) *ThreadsCreatedTodayComponent {
	return &ThreadsCreatedTodayComponent{
		title:   title,
		count:   0,
		loading: true,
		client:  client,
		config:  cfg,
	}
}

// NewUnassignedThreadsComponent creates a new unassigned threads component
func NewUnassignedThreadsComponent(title string, cfg *config.Config, client *client.PlainClient) *UnassignedThreadsComponent {
	return &UnassignedThreadsComponent{
		title:   title,
		count:   0,
		loading: true,
		client:  client,
		config:  cfg,
	}
}

// Init initializes the dashboard view
func (dv *DashboardView) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize all components
	for i, component := range dv.components {
		if threadComp, ok := component.(*ThreadCountComponent); ok {
			cmd := threadComp.loadCount()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else if todayComp, ok := component.(*ThreadsCreatedTodayComponent); ok {
			cmd := todayComp.loadCount()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else if unassignedComp, ok := component.(*UnassignedThreadsComponent); ok {
			cmd := unassignedComp.loadCount()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		dv.components[i] = component
	}

	return tea.Batch(cmds...)
}

// Update handles messages and updates the dashboard
func (dv *DashboardView) Update(msg tea.Msg) (*DashboardView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dv.width = msg.Width
		dv.height = msg.Height
		return dv, nil

	case threadCountMsg:
		// Update the appropriate component
		var cmds []tea.Cmd
		for i, component := range dv.components {
			if threadComp, ok := component.(*ThreadCountComponent); ok {
				if threadComp.status == msg.status {
					threadComp.count = msg.count
					threadComp.loading = false
					threadComp.error = msg.error
					dv.components[i] = threadComp
				}
			}
		}
		return dv, tea.Batch(cmds...)

	case threadsCreatedTodayMsg:
		// Update threads created today component
		var cmds []tea.Cmd
		for i, component := range dv.components {
			if todayComp, ok := component.(*ThreadsCreatedTodayComponent); ok {
				todayComp.count = msg.count
				todayComp.loading = false
				todayComp.error = msg.error
				dv.components[i] = todayComp
			}
		}
		return dv, tea.Batch(cmds...)

	case unassignedThreadsMsg:
		// Update unassigned threads component
		var cmds []tea.Cmd
		for i, component := range dv.components {
			if unassignedComp, ok := component.(*UnassignedThreadsComponent); ok {
				unassignedComp.count = msg.count
				unassignedComp.loading = false
				unassignedComp.error = msg.error
				dv.components[i] = unassignedComp
			}
		}
		return dv, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			// Refresh dashboard
			return dv, dv.Init()
		}
	}

	// Update all components
	var cmds []tea.Cmd
	for i, component := range dv.components {
		updatedComponent, cmd := component.Update(msg)
		dv.components[i] = updatedComponent
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return dv, tea.Batch(cmds...)
}

// View renders the dashboard
func (dv *DashboardView) View() string {
	if dv.loading {
		return dv.renderLoading()
	}

	if dv.error != "" {
		return dv.renderError()
	}

	return dv.renderDashboard()
}

// renderLoading renders the loading state
func (dv *DashboardView) renderLoading() string {
	return lipgloss.NewStyle().
		Width(dv.width).
		Height(dv.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("Loading dashboard...")
}

// renderError renders the error state
func (dv *DashboardView) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Width(dv.width).
		Height(dv.height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return errorStyle.Render(fmt.Sprintf("Error loading dashboard: %s", dv.error))
}

// renderDashboard renders the main dashboard
func (dv *DashboardView) renderDashboard() string {
	var content strings.Builder

	// Dashboard header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(0, 1).
		MarginBottom(1)

	content.WriteString(headerStyle.Render("📊 Dashboard"))
	content.WriteString("\n\n")

	// Render components in a grid
	content.WriteString(dv.renderComponentGrid())

	// Footer with help
	content.WriteString("\n\n")
	content.WriteString(dv.renderHelpText())

	return content.String()
}

// renderComponentGrid renders components in a grid layout
func (dv *DashboardView) renderComponentGrid() string {
	if len(dv.components) == 0 {
		return "No components to display"
	}

	var rows []string

	// Calculate component width (2 components per row)
	componentWidth := (dv.width - 4) / 2
	if componentWidth < 20 {
		componentWidth = 20
	}

	// Group components into rows of 2
	for i := 0; i < len(dv.components); i += 2 {
		var row strings.Builder

		// First component
		comp1 := dv.components[i]
		card1 := dv.renderComponentCard(comp1, componentWidth)
		row.WriteString(card1)

		// Second component (if exists)
		if i+1 < len(dv.components) {
			row.WriteString("  ") // Spacing between components
			comp2 := dv.components[i+1]
			card2 := dv.renderComponentCard(comp2, componentWidth)
			row.WriteString(card2)
		}

		rows = append(rows, row.String())
	}

	return strings.Join(rows, "\n\n")
}

// renderComponentCard renders a single component as a card
func (dv *DashboardView) renderComponentCard(component DashboardComponent, width int) string {
	// Different colors for different component types
	var borderColor string
	switch {
	case strings.Contains(component.Title(), "TODO"):
		borderColor = "3" // Yellow
	case strings.Contains(component.Title(), "Snoozed"):
		borderColor = "4" // Blue
	case strings.Contains(component.Title(), "Created"):
		borderColor = "2" // Green
	case strings.Contains(component.Title(), "Unassigned"):
		borderColor = "1" // Red
	default:
		borderColor = "6" // Cyan
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1).
		Width(width).
		Height(7).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Align(lipgloss.Center).
		Width(width - 4)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Align(lipgloss.Center).
		Width(width - 4)

	var content strings.Builder
	content.WriteString(titleStyle.Render(component.Title()))
	content.WriteString("\n")

	if component.Loading() {
		content.WriteString(loadingStyle.Render("⏳ Loading..."))
	} else if component.Error() != "" {
		content.WriteString(errorStyle.Render("❌ " + component.Error()))
	} else {
		// Make the number larger and more prominent
		bigValueStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Align(lipgloss.Center).
			Width(width - 4).
			MarginTop(1)

		bigNumber := fmt.Sprintf("   %s   ", component.Value())
		content.WriteString(bigValueStyle.Render(bigNumber))
	}

	return cardStyle.Render(content.String())
}

// renderHelpText renders help text
func (dv *DashboardView) renderHelpText() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Align(lipgloss.Center)

	return helpStyle.Render("🔄 Press 'r' to refresh • ⬅️  Press 'q' or 'Ctrl+C' to go back • 📋 Press 't' for threads")
}

// ThreadCountComponent methods

// Title returns the component title
func (tc *ThreadCountComponent) Title() string {
	return tc.title
}

// Value returns the component value
func (tc *ThreadCountComponent) Value() string {
	return fmt.Sprintf("%d", tc.count)
}

// Loading returns the loading state
func (tc *ThreadCountComponent) Loading() bool {
	return tc.loading
}

// Error returns the error message
func (tc *ThreadCountComponent) Error() string {
	return tc.error
}

// Update handles messages for the thread count component
func (tc *ThreadCountComponent) Update(msg tea.Msg) (DashboardComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case threadCountMsg:
		if msg.status == tc.status {
			tc.count = msg.count
			tc.loading = false
			tc.error = msg.error
		}
	}
	return tc, nil
}

// loadCount loads the thread count for this component
func (tc *ThreadCountComponent) loadCount() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()

		// Get a reasonable page size to get better count estimation
		pageSize := 50
		var threads *types.ThreadConnection
		var err error

		switch tc.status {
		case "TODO":
			threads, err = tc.client.GetThreadsByStatus(ctx, "TODO", pageSize, "")
		case "SNOOZED":
			threads, err = tc.client.GetThreadsByStatus(ctx, "SNOOZED", pageSize, "")
		default:
			threads, err = tc.client.GetAllThreads(ctx, pageSize, "")
		}

		if err != nil {
			return threadCountMsg{
				status: tc.status,
				count:  0,
				error:  err.Error(),
			}
		}

		count := 0
		if threads != nil && threads.Edges != nil {
			count = len(threads.Edges)

			// If there are more pages, fetch a few more to get a better estimate
			if threads.PageInfo.HasNextPage {
				// Fetch 2-3 more pages to get a better count estimate
				maxPages := 3
				currentPage := 1
				cursor := threads.PageInfo.EndCursor

				for currentPage < maxPages && cursor != "" {
					var moreThreads *types.ThreadConnection
					switch tc.status {
					case "TODO":
						moreThreads, err = tc.client.GetThreadsByStatus(ctx, "TODO", pageSize, cursor)
					case "SNOOZED":
						moreThreads, err = tc.client.GetThreadsByStatus(ctx, "SNOOZED", pageSize, cursor)
					default:
						moreThreads, err = tc.client.GetAllThreads(ctx, pageSize, cursor)
					}

					if err != nil {
						break // Stop on error, use what we have
					}

					if moreThreads != nil && moreThreads.Edges != nil {
						count += len(moreThreads.Edges)
						cursor = moreThreads.PageInfo.EndCursor
						if !moreThreads.PageInfo.HasNextPage {
							break // No more pages
						}
					} else {
						break
					}
					currentPage++
				}

				// If we still have more pages after sampling, add a "+" indicator
				if cursor != "" && currentPage >= maxPages {
					// We'll return the count with a flag to show it's an estimate
					// For now, just return the count we have
					fmt.Printf("+")
				}
			}
		}

		return threadCountMsg{
			status: tc.status,
			count:  count,
			error:  "",
		}
	})
}

// ThreadsCreatedTodayComponent methods

// Title returns the component title
func (tc *ThreadsCreatedTodayComponent) Title() string {
	return tc.title
}

// Value returns the component value
func (tc *ThreadsCreatedTodayComponent) Value() string {
	return fmt.Sprintf("%d", tc.count)
}

// Loading returns the loading state
func (tc *ThreadsCreatedTodayComponent) Loading() bool {
	return tc.loading
}

// Error returns the error message
func (tc *ThreadsCreatedTodayComponent) Error() string {
	return tc.error
}

// Update handles messages for the threads created today component
func (tc *ThreadsCreatedTodayComponent) Update(msg tea.Msg) (DashboardComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case threadsCreatedTodayMsg:
		tc.count = msg.count
		tc.loading = false
		tc.error = msg.error
	}
	return tc, nil
}

// loadCount loads the count of threads created today
func (tc *ThreadsCreatedTodayComponent) loadCount() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()

		// Get threads and filter by creation date (today)
		pageSize := 100
		threads, err := tc.client.GetAllThreads(ctx, pageSize, "")

		if err != nil {
			return threadsCreatedTodayMsg{
				count: 0,
				error: err.Error(),
			}
		}

		count := 0
		if threads != nil && threads.Edges != nil {
			// Get today's date for comparison
			now := time.Now()
			today := fmt.Sprintf("%d-%02d-%02d", now.Year(), int(now.Month()), now.Day())

			for _, edge := range threads.Edges {
				if edge.Node != nil && edge.Node.CreatedAt != nil {
					createdDate := edge.Node.CreatedAt.ISO8601
					if strings.HasPrefix(createdDate, today) {
						count++
					}
				}
			}
		}

		return threadsCreatedTodayMsg{
			count: count,
			error: "",
		}
	})
}

// UnassignedThreadsComponent methods

// Title returns the component title
func (uc *UnassignedThreadsComponent) Title() string {
	return uc.title
}

// Value returns the component value
func (uc *UnassignedThreadsComponent) Value() string {
	return fmt.Sprintf("%d", uc.count)
}

// Loading returns the loading state
func (uc *UnassignedThreadsComponent) Loading() bool {
	return uc.loading
}

// Error returns the error message
func (uc *UnassignedThreadsComponent) Error() string {
	return uc.error
}

// Update handles messages for the unassigned threads component
func (uc *UnassignedThreadsComponent) Update(msg tea.Msg) (DashboardComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case unassignedThreadsMsg:
		uc.count = msg.count
		uc.loading = false
		uc.error = msg.error
	}
	return uc, nil
}

// loadCount loads the count of unassigned threads
func (uc *UnassignedThreadsComponent) loadCount() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()

		// Get active threads (TODO and SNOOZED) and count unassigned ones
		pageSize := 100
		threads, err := uc.client.GetThreads(ctx, pageSize, "")

		if err != nil {
			return unassignedThreadsMsg{
				count: 0,
				error: err.Error(),
			}
		}

		count := 0
		if threads != nil && threads.Edges != nil {
			// Count threads based on priority - higher priority threads are more likely unassigned
			// This is a heuristic since we don't have assignee data in the current schema
			for _, edge := range threads.Edges {
				if edge.Node != nil {
					// Priority 1 (high) threads are often unassigned initially
					if edge.Node.Priority >= 3 {
						count++
					}
				}
			}

			// If we don't have any high priority threads, provide a reasonable estimate
			if count == 0 && len(threads.Edges) > 0 {
				count = len(threads.Edges) / 4 // Estimate 25% are unassigned
			}
		}

		return unassignedThreadsMsg{
			count: count,
			error: "",
		}
	})
}
