package ui

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"simple/client"
	"simple/config"
	"simple/types"
)

// ThreadFilter represents the current thread filter
type ThreadFilter int

const (
	FilterTODO ThreadFilter = iota
	FilterSNOOZED
	FilterAll
)

// ThreadViewState represents the current view mode
type ThreadViewState int

const (
	ViewList ThreadViewState = iota
	ViewDetail
)

// ThreadItem represents a thread item for the list
type ThreadItem struct {
	Thread *types.Thread
}

// FilterValue returns the filter value for the list
func (t ThreadItem) FilterValue() string {
	return fmt.Sprintf("%s %s", t.Thread.Title, t.Thread.Customer.FullName)
}

// ThreadsView represents the threads view
type ThreadsView struct {
	config         *config.Config
	client         *client.PlainClient
	list           list.Model
	filter         ThreadFilter
	viewState      ThreadViewState
	selectedThread *types.Thread
	loading        bool
	error          string
	cursor         string
	hasNextPage    bool
	viewport       viewport.Model
	viewportReady  bool
	width          int
	height         int
}

// threadsLoadedMsg is sent when threads are loaded
type threadsLoadedMsg struct {
	threads     []*types.Thread
	cursor      string
	hasNextPage bool
	error       string
}

// threadDetailLoadedMsg is sent when thread details with messages are loaded
type threadDetailLoadedMsg struct {
	thread *types.Thread
	error  string
}

// NewThreadsView creates a new threads view
func NewThreadsView(cfg *config.Config, client *client.PlainClient) *ThreadsView {
	// Create list model
	l := list.New([]list.Item{}, threadDelegate{}, 0, 0)
	l.Title = "Threads"
	l.SetShowStatusBar(true)
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	// Custom keybindings
	l.KeyMap.Quit = key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	)

	return &ThreadsView{
		config:    cfg,
		client:    client,
		list:      l,
		filter:    FilterTODO,
		viewState: ViewList,
	}
}

// Init initializes the threads view
func (tv *ThreadsView) Init() tea.Cmd {
	return tv.loadThreads("")
}

// loadThreads loads threads from the API
func (tv *ThreadsView) loadThreads(cursor string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()
		tv.loading = true

		var threads *types.ThreadConnection
		var err error

		switch tv.filter {
		case FilterTODO:
			threads, err = tv.client.GetThreadsByStatus(ctx, "TODO", tv.config.UI.PageSize, cursor)
		case FilterSNOOZED:
			threads, err = tv.client.GetThreadsByStatus(ctx, "SNOOZED", tv.config.UI.PageSize, cursor)
		case FilterAll:
			threads, err = tv.client.GetAllThreads(ctx, tv.config.UI.PageSize, cursor)
		default:
			threads, err = tv.client.GetThreadsByStatus(ctx, "TODO", tv.config.UI.PageSize, cursor)
		}

		if err != nil {
			return threadsLoadedMsg{error: err.Error()}
		}

		if threads == nil || threads.Edges == nil {
			return threadsLoadedMsg{error: "No threads returned from API"}
		}

		threadList := make([]*types.Thread, len(threads.Edges))
		for i, edge := range threads.Edges {
			threadList[i] = edge.Node
		}

		return threadsLoadedMsg{
			threads:     threadList,
			cursor:      threads.PageInfo.EndCursor,
			hasNextPage: threads.PageInfo.HasNextPage,
		}
	})
}

// loadThreadDetail loads detailed thread information with messages from the API
func (tv *ThreadsView) loadThreadDetail(threadID string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()

		thread, err := tv.client.GetThreadWithMessages(ctx, threadID)
		if err != nil {
			return threadDetailLoadedMsg{error: err.Error()}
		}

		return threadDetailLoadedMsg{
			thread: thread,
		}
	})
}

// Update handles messages and updates the model
func (tv *ThreadsView) Update(msg tea.Msg) (*ThreadsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tv.width = msg.Width
		tv.height = msg.Height
		tv.list.SetWidth(msg.Width)
		tv.list.SetHeight(msg.Height - 4) // Account for padding

		// Update viewport if in detail view
		if tv.viewState == ViewDetail {
			tv.updateViewport()
		}
		return tv, nil

	case threadsLoadedMsg:
		tv.loading = false
		if msg.error != "" {
			tv.error = msg.error
		} else {
			tv.error = ""
			tv.cursor = msg.cursor
			tv.hasNextPage = msg.hasNextPage

			// Convert threads to list items
			items := make([]list.Item, len(msg.threads))
			for i, thread := range msg.threads {
				items[i] = ThreadItem{Thread: thread}
			}

			tv.list.SetItems(items)
			tv.updateTitle()
		}
		return tv, nil

	case threadDetailLoadedMsg:
		tv.loading = false
		if msg.error != "" {
			tv.error = msg.error
		} else {
			tv.error = ""
			tv.selectedThread = msg.thread
			tv.updateViewport()
		}
		return tv, nil

	case tea.KeyMsg:
		if tv.viewState == ViewDetail {
			return tv.handleDetailKeys(msg)
		}
		return tv.handleListKeys(msg)
	}

	// Update the list
	var cmd tea.Cmd
	tv.list, cmd = tv.list.Update(msg)
	return tv, cmd
}

// handleListKeys handles key events in list view
func (tv *ThreadsView) handleListKeys(msg tea.KeyMsg) (*ThreadsView, tea.Cmd) {
	switch msg.String() {
	case "1":
		if tv.filter != FilterTODO {
			tv.filter = FilterTODO
			tv.updateTitle()
			return tv, tv.loadThreads("")
		}
	case "2":
		if tv.filter != FilterSNOOZED {
			tv.filter = FilterSNOOZED
			tv.updateTitle()
			return tv, tv.loadThreads("")
		}
	case "3":
		if tv.filter != FilterAll {
			tv.filter = FilterAll
			tv.updateTitle()
			return tv, tv.loadThreads("")
		}
	case "r":
		return tv, tv.loadThreads("")
	case "n":
		if tv.hasNextPage {
			return tv, tv.loadThreads(tv.cursor)
		}
	case "enter":
		if item, ok := tv.list.SelectedItem().(ThreadItem); ok {
			tv.selectedThread = item.Thread
			tv.viewState = ViewDetail
			tv.loading = true
			return tv, tv.loadThreadDetail(item.Thread.ID)
		}
	}

	// Update the list
	var cmd tea.Cmd
	tv.list, cmd = tv.list.Update(msg)
	return tv, cmd
}

// handleDetailKeys handles key events in detail view
func (tv *ThreadsView) handleDetailKeys(msg tea.KeyMsg) (*ThreadsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "esc":
		tv.viewState = ViewList
		tv.selectedThread = nil
		tv.viewportReady = false
	case "b":
		if tv.selectedThread != nil {
			return tv, tv.openInBrowser(tv.selectedThread.ID)
		}
	default:
		// Handle viewport scrolling
		if tv.viewportReady {
			tv.viewport, cmd = tv.viewport.Update(msg)
		}
	}
	return tv, cmd
}

// updateTitle updates the list title based on current filter
func (tv *ThreadsView) updateTitle() {
	var title string
	switch tv.filter {
	case FilterTODO:
		title = "Threads (TODO)"
	case FilterSNOOZED:
		title = "Threads (SNOOZED)"
	case FilterAll:
		title = "Threads (All)"
	}
	tv.list.Title = title
}

// IsInDetailView returns true if in detail view
func (tv *ThreadsView) IsInDetailView() bool {
	return tv.viewState == ViewDetail
}

// View renders the threads view
func (tv *ThreadsView) View() string {
	if tv.loading {
		return tv.renderLoading()
	}

	if tv.error != "" {
		return tv.renderError()
	}

	if tv.viewState == ViewDetail {
		return tv.renderDetail()
	}

	return tv.renderList()
}

// renderLoading renders the loading state
func (tv *ThreadsView) renderLoading() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("69")).
		Padding(2)
	return style.Render("Loading threads...")
}

// renderError renders the error state
func (tv *ThreadsView) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2)

	content := fmt.Sprintf("Error: %s\n\nPress 'r' to retry or 'q' to quit.", tv.error)
	return errorStyle.Render(content)
}

// renderList renders the list view
func (tv *ThreadsView) renderList() string {
	helpText := tv.renderHelpText()
	return tv.list.View() + "\n" + helpText
}

// renderDetail renders the detail view
func (tv *ThreadsView) renderDetail() string {
	if tv.selectedThread == nil {
		return "No thread selected"
	}

	if !tv.viewportReady {
		return "Loading thread details..."
	}

	header := tv.renderDetailHeader()
	footer := tv.renderDetailFooter()

	return fmt.Sprintf("%s\n%s\n%s", header, tv.viewport.View(), footer)
}

// renderDetailHeader renders the header for detail view
func (tv *ThreadsView) renderDetailHeader() string {
	thread := tv.selectedThread

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	// Status styling
	statusColor := getStatusColor(thread.Status)
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(statusColor)

	// Priority styling
	priorityColor := getPriorityColor(thread.Priority)
	priorityStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(priorityColor)

	// Thread details styles
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var content strings.Builder

	// Title and status line
	title := titleStyle.Render(fmt.Sprintf("Thread: %s", thread.Title))
	status := statusStyle.Render(thread.Status)
	priority := priorityStyle.Render(getPriorityString(thread.Priority))
	headerInfo := fmt.Sprintf("%s | %s", status, priority)

	content.WriteString(title)
	content.WriteString("\n")
	content.WriteString(headerInfo)
	content.WriteString("\n\n")

	// Thread details
	content.WriteString(labelStyle.Render("ID: "))
	content.WriteString(valueStyle.Render(thread.ID))
	content.WriteString("\n")

	if thread.Customer != nil {
		content.WriteString(labelStyle.Render("Customer: "))
		content.WriteString(valueStyle.Render(thread.Customer.FullName))
		if thread.Customer.GetEmail() != "" {
			content.WriteString(" (")
			content.WriteString(valueStyle.Render(thread.Customer.GetEmail()))
			content.WriteString(")")
		}
		content.WriteString("\n")

		if thread.Customer.Company != nil && thread.Customer.Company.Name != "" {
			content.WriteString(labelStyle.Render("Company: "))
			content.WriteString(valueStyle.Render(thread.Customer.Company.Name))
			content.WriteString("\n")
		}
	}

	if thread.CreatedAt != nil {
		if t, err := thread.CreatedAt.Time(); err == nil {
			content.WriteString(labelStyle.Render("Created: "))
			content.WriteString(valueStyle.Render(t.Format("2006-01-02 15:04:05")))
			content.WriteString("\n")
		}
	}

	if thread.UpdatedAt != nil {
		if t, err := thread.UpdatedAt.Time(); err == nil {
			content.WriteString(labelStyle.Render("Updated: "))
			content.WriteString(valueStyle.Render(t.Format("2006-01-02 15:04:05")))
			content.WriteString("\n")
		}
	}

	// Create a border line
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	line := strings.Repeat("─", tv.width-2)
	content.WriteString("\n")
	content.WriteString(borderStyle.Render(line))

	return content.String()
}

// renderDetailFooter renders the footer for detail view
func (tv *ThreadsView) renderDetailFooter() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("118"))

	help := "b: Open in browser • q/esc: Back to list • ↑/↓: Scroll"

	scrollInfo := ""
	if tv.viewportReady {
		scrollInfo = fmt.Sprintf("%3.f%%", tv.viewport.ScrollPercent()*100)
	}

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	line := strings.Repeat("─", max(0, tv.width-lipgloss.Width(scrollInfo)-2))

	if scrollInfo != "" {
		return fmt.Sprintf("%s\n%s %s", borderStyle.Render(line), helpStyle.Render(help), scrollStyle.Render(scrollInfo))
	}

	return fmt.Sprintf("%s\n%s", borderStyle.Render(line), helpStyle.Render(help))
}

// updateViewport updates the viewport with thread messages
func (tv *ThreadsView) updateViewport() {
	if tv.selectedThread == nil || tv.width == 0 || tv.height == 0 {
		return
	}

	// Calculate header height dynamically
	headerContent := tv.renderDetailHeader()
	headerHeight := lipgloss.Height(headerContent) + 1 // +1 for spacing

	// Footer height
	footerHeight := 2 // Border + help text

	// Calculate viewport dimensions
	viewportHeight := tv.height - headerHeight - footerHeight

	if viewportHeight < 3 {
		viewportHeight = 3
	}

	// Initialize viewport if not ready
	if !tv.viewportReady {
		tv.viewport = viewport.New(tv.width-2, viewportHeight)
		tv.viewport.YPosition = headerHeight
		tv.viewportReady = true
	} else {
		tv.viewport.Width = tv.width - 2
		tv.viewport.Height = viewportHeight
	}

	// Generate content for viewport
	content := tv.renderThreadContent()
	tv.viewport.SetContent(content)
}

// renderThreadContent renders only the messages for the scrollable viewport
func (tv *ThreadsView) renderThreadContent() string {
	if tv.selectedThread == nil {
		return "No thread selected"
	}

	thread := tv.selectedThread
	var content strings.Builder

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Messages section from timeline entries
	if thread.TimelineEntries != nil && len(thread.TimelineEntries.Edges) > 0 {
		content.WriteString(labelStyle.Render("Messages"))
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", 50))
		content.WriteString("\n\n")

		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 0, 1, 0)

		senderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("118"))

		timestampStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

		entryTypeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

		for i := len(thread.TimelineEntries.Edges) - 1; i >= 0; i-- {
			edge := thread.TimelineEntries.Edges[i]
			entry := edge.Node

			// Message header with sender and timestamp
			sender := "Unknown"
			if entry.Actor != nil {
				sender = entry.Actor.GetFullName()
				if sender == "" {
					sender = entry.Actor.GetID()
				}
			}

			timestamp := ""
			if entry.Timestamp != nil {
				if t, err := entry.Timestamp.Time(); err == nil {
					timestamp = t.Format("2006-01-02 15:04:05")
				}
			}

			content.WriteString(senderStyle.Render(sender))
			if timestamp != "" {
				content.WriteString(" ")
				content.WriteString(timestampStyle.Render(timestamp))
			}
			content.WriteString("\n")

			// Message content based on entry type
			messageContent := tv.getEntryContent(entry)
			entryType := tv.getEntryType(entry)

			if entryType != "" {
				content.WriteString(entryTypeStyle.Render(fmt.Sprintf("[%s] ", entryType)))
			}

			if messageContent != "" {
				content.WriteString(messageStyle.Render(messageContent))
			} else {
				content.WriteString(messageStyle.Render("(no content)"))
			}

			// Add separator between messages (except for the last one)
			if i > 0 {
				content.WriteString("\n")
				content.WriteString(strings.Repeat("─", 30))
				content.WriteString("\n\n")
			}
		}
	} else {
		content.WriteString(labelStyle.Render("Messages"))
		content.WriteString("\n")
		content.WriteString(strings.Repeat("─", 50))
		content.WriteString("\n\n")
		content.WriteString(valueStyle.Render("No messages in this thread"))
	}

	return content.String()
}

// renderHelpText renders the help text for list view
func (tv *ThreadsView) renderHelpText() string {
	helpItems := []string{
		"1: TODO only",
		"2: SNOOZED only",
		"3: All threads",
		"r: Refresh",
		"enter: View details",
		"d: Dashboard",
	}

	if tv.hasNextPage {
		helpItems = append(helpItems, "n: Next page")
	}

	helpItems = append(helpItems, "q: Quit")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0, 0, 0)

	return helpStyle.Render(strings.Join(helpItems, " • "))
}

// openInBrowser opens the thread in the browser
func (tv *ThreadsView) openInBrowser(threadID string) tea.Cmd {
	url := fmt.Sprintf("https://app.plain.com/workspace/%s/thread/%s", tv.config.Plain.WorkspaceID, threadID)

	return tea.Cmd(func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			// Unsupported platform, do nothing
			return nil
		}

		if cmd != nil {
			_ = cmd.Start()
		}
		return nil
	})
}

// threadDelegate implements list.ItemDelegate for ThreadItem
type threadDelegate struct{}

func (d threadDelegate) Height() int                             { return 2 }
func (d threadDelegate) Spacing() int                            { return 1 }
func (d threadDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d threadDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ThreadItem)
	if !ok {
		return
	}

	thread := i.Thread
	var str strings.Builder

	// First line: Title and Status
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	statusColor := getStatusColor(thread.Status)
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	str.WriteString(titleStyle.Render(fmt.Sprintf("%-50s", truncateString(thread.Title, 50))))
	str.WriteString(" ")
	str.WriteString(statusStyle.Render(thread.Status))

	// Second line: Customer and Company info
	customerInfo := "No customer"
	if thread.Customer != nil {
		customerInfo = thread.Customer.FullName
		if thread.Customer.Company != nil && thread.Customer.Company.Name != "" {
			customerInfo += " @ " + thread.Customer.Company.Name
		}
	}

	// Priority info
	priorityColor := getPriorityColor(thread.Priority)
	priorityStyle := lipgloss.NewStyle().
		Foreground(priorityColor)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	str.WriteString("\n")
	str.WriteString(infoStyle.Render(fmt.Sprintf("%-50s", truncateString(customerInfo, 50))))
	str.WriteString(" ")
	str.WriteString(priorityStyle.Render(getPriorityString(thread.Priority)))

	// Apply selection styling
	if index == m.Index() {
		selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
		fmt.Fprint(w, selectedStyle.Render(str.String()))
	} else {
		normalStyle := lipgloss.NewStyle().
			Padding(0, 1)
		fmt.Fprint(w, normalStyle.Render(str.String()))
	}
}

// getStatusColor returns the color for a thread status
func getStatusColor(status string) lipgloss.Color {
	switch status {
	case "TODO":
		return lipgloss.Color("3") // Yellow
	case "SNOOZED":
		return lipgloss.Color("4") // Blue
	case "OPEN":
		return lipgloss.Color("2") // Green
	case "PENDING":
		return lipgloss.Color("3") // Yellow
	case "DONE":
		return lipgloss.Color("8") // Gray
	default:
		return lipgloss.Color("15") // White
	}
}

// getPriorityColor returns the color for a thread priority
func getPriorityColor(priority int) lipgloss.Color {
	switch priority {
	case 0:
		return lipgloss.Color("9") // Bright red - Urgent
	case 1:
		return lipgloss.Color("1") // Red - High
	case 2:
		return lipgloss.Color("3") // Yellow - Medium
	case 3:
		return lipgloss.Color("2") // Green - Low
	default:
		return lipgloss.Color("15") // White
	}
}

// getPriorityString converts a priority number to a readable string
func getPriorityString(priority int) string {
	switch priority {
	case 0:
		return "Urgent"
	case 1:
		return "High"
	case 2:
		return "Medium"
	case 3:
		return "Low"
	default:
		return fmt.Sprintf("P%d", priority)
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getEntryContent extracts content from a timeline entry based on its type
func (tv *ThreadsView) getEntryContent(entry *types.TimelineEntry) string {
	if entry.Entry == nil {
		return ""
	}

	switch e := entry.Entry.(type) {
	case *types.EmailEntry:
		return e.TextContent
	case *types.ChatEntry:
		return e.Text
	case *types.NoteEntry:
		if e.Markdown != "" {
			return e.Markdown
		}
		return e.Text
	case *types.SlackMessageEntry:
		return e.Text
	case *types.SlackReplyEntry:
		return e.Text
	case *types.CustomEntry:
		return e.Title

	case *types.ThreadStatusTransitionedEntry:
		return fmt.Sprintf("Status changed from %s to %s", e.PreviousStatus, e.NextStatus)
	case *types.ThreadPriorityChangedEntry:
		return fmt.Sprintf("Priority changed from %s to %s", getPriorityString(e.PreviousPriority), getPriorityString(e.NextPriority))
	case map[string]interface{}:
		// Handle raw JSON entries that weren't specifically typed
		// Check for aliased text fields first (from GraphQL fragments)
		if slackText, ok := e["slackText"].(string); ok && slackText != "" {
			return slackText
		}
		if chatText, ok := e["chatText"].(string); ok && chatText != "" {
			return chatText
		}
		if noteText, ok := e["noteText"].(string); ok && noteText != "" {
			return noteText
		}

		// Check for standard text fields
		if text, ok := e["text"].(string); ok && text != "" {
			return text
		}
		if textContent, ok := e["textContent"].(string); ok && textContent != "" {
			return textContent
		}
		if markdown, ok := e["markdown"].(string); ok && markdown != "" {
			return markdown
		}
		if title, ok := e["title"].(string); ok && title != "" {
			return title
		}

		// Handle system event entries
		if prevStatus, ok := e["previousStatus"].(string); ok {
			if nextStatus, ok := e["nextStatus"].(string); ok {
				return fmt.Sprintf("Status changed from %s to %s", prevStatus, nextStatus)
			}
		}
		if prevPriority, ok := e["previousPriority"].(float64); ok {
			if nextPriority, ok := e["nextPriority"].(float64); ok {
				return fmt.Sprintf("Priority changed from %s to %s", getPriorityString(int(prevPriority)), getPriorityString(int(nextPriority)))
			}
		}

		// Handle assignment changes in raw format
		if prevAssignee, ok := e["previousAssignee"]; ok || e["nextAssignee"] != nil {
			prevName := "None"
			nextName := "None"

			if prevAssignee != nil {
				if assigneeMap, ok := prevAssignee.(map[string]interface{}); ok {
					if user, ok := assigneeMap["user"].(map[string]interface{}); ok {
						if fullName, ok := user["fullName"].(string); ok {
							prevName = fullName
						}
					} else if team, ok := assigneeMap["team"].(map[string]interface{}); ok {
						if teamName, ok := team["name"].(string); ok {
							prevName = teamName + " (Team)"
						}
					}
				}
			}

			if nextAssignee, ok := e["nextAssignee"]; ok && nextAssignee != nil {
				if assigneeMap, ok := nextAssignee.(map[string]interface{}); ok {
					if user, ok := assigneeMap["user"].(map[string]interface{}); ok {
						if fullName, ok := user["fullName"].(string); ok {
							nextName = fullName
						}
					} else if team, ok := assigneeMap["team"].(map[string]interface{}); ok {
						if teamName, ok := team["name"].(string); ok {
							nextName = teamName + " (Team)"
						}
					}
				}
			}

			return fmt.Sprintf("Assignment changed from %s to %s", prevName, nextName)
		}

		// For any other entry type, try to extract some meaningful content
		if len(e) > 0 {
			// Look for common content fields in any order
			contentFields := []string{"content", "message", "description", "body", "value"}
			for _, field := range contentFields {
				if value, ok := e[field].(string); ok && value != "" {
					return value
				}
			}
		}

		return ""
	default:
		return ""
	}
}

// getEntryType returns a readable type name for a timeline entry
func (tv *ThreadsView) getEntryType(entry *types.TimelineEntry) string {
	if entry.Entry == nil {
		return "Event"
	}

	switch e := entry.Entry.(type) {
	case *types.EmailEntry:
		return "Email"
	case *types.ChatEntry:
		return "Chat"
	case *types.NoteEntry:
		return "Note"
	case *types.SlackMessageEntry:
		return "Slack"
	case *types.SlackReplyEntry:
		return "Slack Reply"
	case *types.CustomEntry:
		return "Custom"

	case *types.ThreadStatusTransitionedEntry:
		return "Status Change"
	case *types.ThreadPriorityChangedEntry:
		return "Priority Change"
	case map[string]interface{}:
		// Handle raw JSON entries by looking at discriminator fields
		// Check for specific ID fields first
		if _, ok := e["emailId"]; ok {
			return "Email"
		}
		if _, ok := e["chatId"]; ok {
			return "Chat"
		}
		if _, ok := e["noteId"]; ok {
			return "Note"
		}

		// Check for aliased text fields from GraphQL fragments
		if _, ok := e["slackText"]; ok {
			return "Slack"
		}
		if _, ok := e["chatText"]; ok {
			return "Chat"
		}
		if _, ok := e["noteText"]; ok {
			return "Note"
		}

		// Check for Slack-specific fields
		if _, ok := e["slackMessageLink"]; ok {
			return "Slack"
		}
		if _, ok := e["slackWebMessageLink"]; ok {
			return "Slack"
		}

		// Check for system event fields
		if _, ok := e["previousStatus"]; ok {
			return "Status Change"
		}
		if _, ok := e["previousPriority"]; ok {
			return "Priority Change"
		}
		if _, ok := e["previousAssignee"]; ok {
			return "Assignment"
		}

		// Check for custom entry fields
		if _, ok := e["externalId"]; ok {
			return "Custom"
		}
		if entryType, ok := e["type"].(string); ok && entryType != "" {
			return "Custom"
		}
		if _, ok := e["title"]; ok {
			return "Custom"
		}

		// Check for generic text content
		if _, ok := e["text"]; ok {
			return "Message"
		}
		if _, ok := e["textContent"]; ok {
			return "Message"
		}
		if _, ok := e["markdown"]; ok {
			return "Note"
		}

		return "Event"
	default:
		return "Event"
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
