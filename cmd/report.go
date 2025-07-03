package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"simple/client"
	"simple/config"
	"simple/types"
)

// ReportCmd represents the report command.
type ReportCmd struct {
	Range   string `arg:"" enum:"1d,7d,30d,60d" placeholder:"7d" default:"1d" help:"Generate a report of threads for a time range, accepts [1d, 7d,30d]"`
	Summary bool   `help:"Display only the summary of the report"`
}

// Run executes the report command.
func (r *ReportCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	// Calculate date range based on the specified range.
	now := time.Now()
	var startTime time.Time

	switch r.Range {
	case "1d":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	case "60d":
		startTime = now.Add(-60 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
	}

	// Create DateTime objects for the API call.
	after := types.DateTime{ISO8601: startTime.UTC().Format(time.RFC3339)}
	// before := types.DateTime{ISO8601: now.Format(time.RFC3339)}

	fmt.Printf("Generating report for threads from %s to %s\n",
		startTime.Format("2006-01-02 15:04"),
		now.Format("2006-01-02 15:04"))

	// Fetch threads from the API.
	threads, err := client.GetThreadsByDateRange(ctx, after.ISO8601, 100, "")
	if err != nil {
		return fmt.Errorf("failed to get threads for date range: %w", err)
	}

	// If we are paginating, we should fetch all of the threads until we reach the end.
	if threads.PageInfo.HasNextPage {
		nextThreads, err := client.GetThreadsByDateRange(ctx, after.ISO8601, 100, threads.PageInfo.EndCursor)
		if err != nil {
			return fmt.Errorf("failed to get next page of threads: %w", err)
		}
		threads.Edges = append(threads.Edges, nextThreads.Edges...)
		// while there are more pages, continue fetching the threads.
		for nextThreads.PageInfo.HasNextPage {
			fmt.Printf("Found another page -- continuing\n")
			nextThreads, err = client.GetThreadsByDateRange(ctx, after.ISO8601, 100, nextThreads.PageInfo.EndCursor)
			if err != nil {
				return fmt.Errorf("failed to get next page of threads: %w", err)
			}
			threads.Edges = append(threads.Edges, nextThreads.Edges...)
		}
	}

	if threads == nil || len(threads.Edges) == 0 {
		fmt.Println("No threads found for the specified date range")
		return nil
	}

	if r.Summary {
		fmt.Printf("\n=== Summary ===\n")
		r.displaySummary(threads)
		return nil
	}

	// Display the report.
	err = r.displayReport(threads, r.Range)
	if err != nil {
		return fmt.Errorf("failed to display report: %w", err)
	}

	return nil
}

// displayReport formats and displays the thread report.
// This report is intended to get details on the threads UPDATED, not CREATED, after the timestamp passed.
func (r *ReportCmd) displayReport(threads *types.ThreadConnection, timeRange string) error {
	fmt.Printf("\n=== Thread Report (%s) ===\n", timeRange)
	fmt.Printf("Total threads found: %d\n\n", len(threads.Edges))

	// Create table writer for detailed thread list.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header with LABELS column.
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tLABELS\tCUSTOMER\tCOMPANY\tCREATED\tUPDATED")
	fmt.Fprintln(w, "---\t-----\t------\t------\t--------\t-------\t-------\t-------")

	// Print threads.
	for _, edge := range threads.Edges {
		thread := edge.Node
		if thread == nil {
			continue
		}

		// Format created date.
		createdAt := "N/A"
		if thread.CreatedAt != nil {
			if t, err := thread.CreatedAt.Time(); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}
		}

		// Format updated date.
		updatedAt := "N/A"
		if thread.UpdatedAt != nil {
			if t, err := thread.UpdatedAt.Time(); err == nil {
				updatedAt = t.Format("2006-01-02 15:04")
			}
		}

		// Get customer info.
		customerName := "N/A"
		companyName := "N/A"
		if thread.Customer != nil {
			customerName = thread.Customer.FullName
			if thread.Customer.Company != nil {
				companyName = thread.Customer.Company.Name
			}
		}

		// Truncate title if too long.
		title := thread.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		// Process labels: join multiple labels with a comma separator.
		labelsList := "N/A"
		if thread.Labels != nil && len(thread.Labels) > 0 {
			var labelNames []string
			for _, label := range thread.Labels {
				labelNames = append(labelNames, label.LabelType.Name)
			}
			if len(labelNames) > 0 {
				labelsList = strings.Join(labelNames, ", ")
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			thread.ID,
			title,
			thread.Status,
			labelsList,
			customerName,
			companyName,
			createdAt,
			updatedAt,
		)
	}

	fmt.Printf("\n=== Summary ===\n")
	r.displaySummary(threads)

	return nil
}

// displaySummary shows aggregate statistics for the thread report.
func (r *ReportCmd) displaySummary(threads *types.ThreadConnection) {
	statusCounts := make(map[string]int)

	for _, edge := range threads.Edges {
		if edge.Node != nil {
			statusCounts[edge.Node.Status]++
		}
	}

	fmt.Printf("Thread counts by status:\n")
	for status, count := range statusCounts {
		fmt.Printf("  %s: %d\n", status, count)
	}
}
