package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"simple/client"
	"simple/config"
	"simple/types"
)

// priorityToString converts a priority number to a readable string
func priorityToString(priority int) string {
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
		return fmt.Sprintf("%d", priority)
	}
}

// ThreadsCmd represents the threads command
type ThreadsCmd struct {
	List ThreadsListCmd `cmd:"" help:"List threads"`
	All  ThreadsAllCmd  `cmd:"" help:"List all threads (including done)"`
	Get  ThreadsGetCmd  `cmd:"" help:"Get thread by ID"`
}

// ThreadsListCmd lists threads
type ThreadsListCmd struct {
	Limit  int    `help:"Number of threads to retrieve" default:"20"`
	Cursor string `help:"Cursor for pagination" optional:""`
	Status string `help:"Filter by status (TODO, SNOOZED, DONE)" optional:""`
}

// Run executes the threads list command
func (t *ThreadsListCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	var threads *types.ThreadConnection
	var err error

	if t.Status != "" {
		threads, err = client.GetThreadsByStatus(ctx, t.Status, t.Limit, t.Cursor)
	} else {
		threads, err = client.GetThreads(ctx, t.Limit, t.Cursor)
	}

	if err != nil {
		return fmt.Errorf("failed to get threads: %w", err)
	}

	if threads == nil || len(threads.Edges) == 0 {
		fmt.Println("No threads found")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY\tCUSTOMER\tCOMPANY\tCREATED")
	fmt.Fprintln(w, "---\t-----\t------\t--------\t--------\t-------\t-------")

	// Print threads
	for _, edge := range threads.Edges {
		thread := edge.Node
		createdAt := "N/A"
		if thread.CreatedAt != nil {
			if t, err := thread.CreatedAt.Time(); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}
		}

		customerName := "N/A"
		if thread.Customer != nil {
			customerName = thread.Customer.FullName
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\t%s\t%s\n",
			thread.ID,
			thread.Title,
			thread.Status,
			priorityToString(thread.Priority),
			customerName,
			thread.Customer.Company.Name,
			createdAt,
		)
	}

	// Print pagination info
	if threads.PageInfo.HasNextPage {
		fmt.Printf("\nNext page cursor: %s\n", threads.PageInfo.EndCursor)
	}

	return nil
}

// ThreadsAllCmd lists all threads including completed ones
type ThreadsAllCmd struct {
	Limit  int    `help:"Number of threads to retrieve" default:"20"`
	Cursor string `help:"Cursor for pagination" optional:""`
}

// Run executes the threads all command
func (t *ThreadsAllCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	threads, err := client.GetAllThreads(ctx, t.Limit, t.Cursor)
	if err != nil {
		return fmt.Errorf("failed to get threads: %w", err)
	}

	if threads == nil || len(threads.Edges) == 0 {
		fmt.Println("No threads found")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY\tCUSTOMER\tCOMPANY\tCREATED")
	fmt.Fprintln(w, "---\t-----\t------\t--------\t--------\t-------\t-------")

	// Print threads
	for _, edge := range threads.Edges {
		thread := edge.Node
		createdAt := "N/A"
		if thread.CreatedAt != nil {
			if t, err := thread.CreatedAt.Time(); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}
		}

		customerName := "N/A"
		if thread.Customer != nil {
			customerName = thread.Customer.FullName
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\t%s\t%s\n",
			thread.ID,
			thread.Title,
			thread.Status,
			priorityToString(thread.Priority),
			customerName,
			thread.Customer.Company.Name,
			createdAt,
		)
	}

	// Print pagination info
	if threads.PageInfo.HasNextPage {
		fmt.Printf("\nNext page cursor: %s\n", threads.PageInfo.EndCursor)
	}

	return nil
}

// ThreadsGetCmd gets a thread by ID
type ThreadsGetCmd struct {
	ID string `arg:"" help:"Thread ID"`
}

// Run executes the threads get command
func (t *ThreadsGetCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	thread, err := client.GetThreadById(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	if thread == nil {
		fmt.Printf("Thread with ID '%s' not found\n", t.ID)
		return nil
	}

	// Print thread details
	fmt.Printf("Thread Details:\n")
	fmt.Printf("  ID: %s\n", thread.ID)
	fmt.Printf("  Title: %s\n", thread.Title)
	fmt.Printf("  Status: %s\n", thread.Status)
	fmt.Printf("  Priority: %s\n", priorityToString(thread.Priority))

	if thread.Customer != nil {
		fmt.Printf("  Customer: %s (%s)\n", thread.Customer.FullName, thread.Customer.GetEmail())
	}

	if thread.CreatedAt != nil {
		if t, err := thread.CreatedAt.Time(); err == nil {
			fmt.Printf("  Created: %s\n", t.Format("2006-01-02 15:04:05"))
		}
	}

	if thread.UpdatedAt != nil {
		if t, err := thread.UpdatedAt.Time(); err == nil {
			fmt.Printf("  Updated: %s\n", t.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}
