package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"simple/client"
	"simple/config"
)

// CustomersCmd represents the customers command
type CustomersCmd struct {
	List   CustomersListCmd   `cmd:"" help:"List customers"`
	Get    CustomersGetCmd    `cmd:"" help:"Get customer by email"`
	Search CustomersSearchCmd `cmd:"" help:"Search customers"`
}

// CustomersListCmd lists customers
type CustomersListCmd struct {
	Limit  int    `help:"Number of customers to retrieve" default:"20"`
	Cursor string `help:"Cursor for pagination" optional:""`
}

// Run executes the customers list command
func (c *CustomersListCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	customers, err := client.GetCustomers(ctx, c.Limit, c.Cursor)
	if err != nil {
		return fmt.Errorf("failed to get customers: %w", err)
	}

	if customers == nil || len(customers.Edges) == 0 {
		fmt.Println("No customers found")
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tSTATUS\tCREATED")
	fmt.Fprintln(w, "---\t----\t-----\t------\t-------")

	// Print customers
	for _, edge := range customers.Edges {
		customer := edge.Node
		createdAt := "N/A"
		if customer.CreatedAt != nil {
			if t, err := customer.CreatedAt.Time(); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			customer.ID,
			customer.FullName,
			customer.GetEmail(),
			customer.Status,
			createdAt,
		)
	}

	// Print pagination info
	if customers.PageInfo.HasNextPage {
		fmt.Printf("\nNext page cursor: %s\n", customers.PageInfo.EndCursor)
	}

	return nil
}

// CustomersGetCmd gets a customer by email
type CustomersGetCmd struct {
	Email string `arg:"" help:"Customer email address"`
}

// Run executes the customers get command
func (c *CustomersGetCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	customer, err := client.GetCustomerByEmail(ctx, c.Email)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		fmt.Printf("Customer with email '%s' not found\n", c.Email)
		return nil
	}

	// Print customer details
	fmt.Printf("Customer Details:\n")
	fmt.Printf("  ID: %s\n", customer.ID)
	fmt.Printf("  Name: %s\n", customer.FullName)
	fmt.Printf("  Email: %s\n", customer.GetEmail())
	fmt.Printf("  Status: %s\n", customer.Status)

	if customer.CreatedAt != nil {
		if t, err := customer.CreatedAt.Time(); err == nil {
			fmt.Printf("  Created: %s\n", t.Format("2006-01-02 15:04:05"))
		}
	}

	if customer.UpdatedAt != nil {
		if t, err := customer.UpdatedAt.Time(); err == nil {
			fmt.Printf("  Updated: %s\n", t.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}

// CustomersSearchCmd searches customers
type CustomersSearchCmd struct {
	Query string `arg:"" help:"Search query"`
	Limit int    `help:"Number of results to return" default:"10"`
}

// Run executes the customers search command
func (c *CustomersSearchCmd) Run(cfg *config.Config) error {
	ctx := context.Background()
	client := client.NewPlainClient(cfg)

	customers, err := client.SearchCustomers(ctx, c.Query, c.Limit)
	if err != nil {
		return fmt.Errorf("failed to search customers: %w", err)
	}

	if len(customers) == 0 {
		fmt.Printf("No customers found matching '%s'\n", c.Query)
		return nil
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tSTATUS\tCREATED")
	fmt.Fprintln(w, "---\t----\t-----\t------\t-------")

	// Print customers
	for _, customer := range customers {
		createdAt := "N/A"
		if customer.CreatedAt != nil {
			if t, err := customer.CreatedAt.Time(); err == nil {
				createdAt = t.Format("2006-01-02 15:04")
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			customer.ID,
			customer.FullName,
			customer.GetEmail(),
			customer.Status,
			createdAt,
		)
	}

	return nil
}
