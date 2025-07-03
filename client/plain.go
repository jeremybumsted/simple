package client

import (
	"context"
	"fmt"

	"simple/config"
	"simple/types"

	"github.com/machinebox/graphql"
)

// PlainClient wraps the GraphQL client for Plain API
type PlainClient struct {
	client *graphql.Client
	config *config.Config
}

// NewPlainClient creates a new Plain API client
func NewPlainClient(cfg *config.Config) *PlainClient {
	client := graphql.NewClient(cfg.Plain.Endpoint)

	return &PlainClient{
		client: client,
		config: cfg,
	}
}

// setHeaders sets the required headers for API requests
func (c *PlainClient) setHeaders(req *graphql.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Plain.APIKey))
}

// GetCustomerByEmail retrieves a customer by their email address
func (c *PlainClient) GetCustomerByEmail(ctx context.Context, email string) (*types.Customer, error) {
	req := graphql.NewRequest(`
		query customerByEmail($email: String!) {
			customerByEmail(email: $email) {
				id
				fullName
				email {
					email
				}
				status
				createdAt {
					iso8601
				}
				company {
					id
					name
				}
			}
		}
	`)

	req.Var("email", email)
	c.setHeaders(req)

	var resp struct {
		CustomerByEmail *types.Customer `json:"customerByEmail"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return resp.CustomerByEmail, nil
}

// GetCustomers retrieves a list of customers with pagination
func (c *PlainClient) GetCustomers(ctx context.Context, limit int, cursor string) (*types.CustomerConnection, error) {
	req := graphql.NewRequest(`
		query customers($first: Int!, $after: String) {
			customers(first: $first, after: $after) {
				edges {
					node {
						id
						fullName
						email {
							email
						}
						status
						company {
							id
							name
						}
						createdAt {
							iso8601
						}
					}
					cursor
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`)

	req.Var("first", limit)

	if cursor != "" {
		req.Var("after", cursor)
	}
	c.setHeaders(req)

	var resp struct {
		Customers *types.CustomerConnection `json:"threads"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get threads: %w", err)
	}

	return resp.Customers, nil
}

// GetThreadsByDateRange retrieves threads filtered by updated date range
// The statusDetails field is intentionally excluding threads that are IGNORED.
func (c *PlainClient) GetThreadsByDateRange(ctx context.Context, dateAfter string, limit int, cursor string) (*types.ThreadConnection, error) {

	req := graphql.NewRequest(`
		query GetThreadsByDateRange($first: Int!, $dateAfter: String, $cursor: String) {
			threads(first: $first, after: $cursor, filters: {
			statuses: [TODO,SNOOZED,DONE]
			updatedAt: {
				after: $dateAfter
			}
			statusDetails: [
		      CREATED,
		      IN_PROGRESS,
		      NEW_REPLY,
		      THREAD_LINK_UPDATED,
		      THREAD_DISCUSSION_RESOLVED,
		      WAITING_FOR_CUSTOMER,
		      WAITING_FOR_DURATION,
		      DONE_MANUALLY_SET,
		      DONE_AUTOMATICALLY_SET
		    ],
			isMarkedAsSpam: false
			}) {
    edges {
      node {
        id
        title
        status
        labels {
          labelType {
            name
            icon
          }
        }
        threadFields {
          key
          stringValue
          booleanValue
        }
        customer {
          fullName
          company {
            name
          }
        }
        assignedTo {
          ... on User {
            publicName
          }
        }
        updatedAt {
          iso8601
        }
        links {
          edges {
            node {
              ... on LinearIssueThreadLink {
                url
              }
            }
          }
        }
        createdAt {
          iso8601
        }
      }
    }
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
  }
		}
	`)
	req.Var("dateAfter", dateAfter)
	if cursor == "" {
		req.Var("cursor", nil)
	} else {
		req.Var("cursor", cursor)
	}
	req.Var("first", limit)

	c.setHeaders(req)

	var resp struct {
		Threads *types.ThreadConnection `json:"threads"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get threads: %w", err)
	}

	return resp.Threads, nil
}

// GetThreads retrieves a list of threads with pagination
func (c *PlainClient) GetThreads(ctx context.Context, limit int, cursor string) (*types.ThreadConnection, error) {
	req := graphql.NewRequest(`
		query threads($first: Int!, $after: String) {
			threads(first: $first, after: $after, filters: { statuses: [TODO, SNOOZED] }) {
				edges {
					node {
						id
						title
						status
						priority
						createdAt {
							iso8601
						}
						updatedAt {
							iso8601
						}
						customer {
							id
							fullName
							email {
								email
							}
							company {
								id
								name
							}
						}
						assignedTo {
							... on User {
								id
								fullName
								email
							}
						}
					}
					cursor
				}
				pageInfo {
					hasNextPage
					endCursor
				}
				totalCount
			}
		}
	`)

	req.Var("first", limit)
	if cursor != "" {
		req.Var("after", cursor)
	}
	c.setHeaders(req)

	var resp struct {
		Threads *types.ThreadConnection `json:"threads"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get threads: %w", err)
	}

	return resp.Threads, nil
}

// GetAllThreads retrieves all threads including completed ones
func (c *PlainClient) GetAllThreads(ctx context.Context, limit int, cursor string) (*types.ThreadConnection, error) {
	req := graphql.NewRequest(`
		query threads($first: Int!, $after: String) {
			threads(first: $first, after: $after) {
				edges {
					node {
						id
						title
						status
						priority
						createdAt {
							iso8601
						}
						updatedAt {
							iso8601
						}
						customer {
							id
							fullName
							email {
								email
							}
							company {
								id
								name
							}
						}
						assignedTo {
							... on User {
								id
								fullName
								email
							}
						}
					}
					cursor
				}
				pageInfo {
					hasNextPage
					endCursor
				}
				totalCount
			}
		}
	`)

	req.Var("first", limit)
	if cursor != "" {
		req.Var("after", cursor)
	}
	c.setHeaders(req)

	var resp struct {
		Threads *types.ThreadConnection `json:"threads"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get all threads: %w", err)
	}

	return resp.Threads, nil
}

// GetThreadsByStatus retrieves threads filtered by status
func (c *PlainClient) GetThreadsByStatus(ctx context.Context, status string, limit int, cursor string) (*types.ThreadConnection, error) {
	req := graphql.NewRequest(`
		query threads($first: Int!, $after: String, $status: ThreadStatus!) {
			threads(first: $first, after: $after, filters: { statuses: [$status] }) {
				edges {
					node {
						id
						title
						status
						priority
						createdAt {
							iso8601
						}
						updatedAt {
							iso8601
						}
						customer {
							id
							fullName
							email {
								email
							}
							company {
								id
								name
							}
						}
						assignedTo{
              ... on User {
                publicName
              }
            }
					}
					cursor
				}
				pageInfo {
					hasNextPage
					endCursor
				}
				totalCount
			}
		}
	`)

	req.Var("first", limit)
	req.Var("status", status)
	if cursor != "" {
		req.Var("after", cursor)
	}
	c.setHeaders(req)

	var resp struct {
		Threads *types.ThreadConnection `json:"threads"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get threads by status: %w", err)
	}

	return resp.Threads, nil
}

// GetThreadById retrieves a single thread by ID
func (c *PlainClient) GetThreadById(ctx context.Context, threadId string) (*types.Thread, error) {
	req := graphql.NewRequest(`
		query thread($threadId: ID!) {
			thread(threadId: $threadId) {
				id
				title
				status
				priority
				createdAt {
					iso8601
				}
				updatedAt {
					iso8601
				}
				customer {
					id
					fullName
					email {
						email
					}
					company {
						id
						name
					}
				}
				assignedTo {
					... on User {
						id
						fullName
						email
					}
				}
			}
		}
	`)

	req.Var("threadId", threadId)
	c.setHeaders(req)

	var resp struct {
		Thread *types.Thread `json:"thread"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	return resp.Thread, nil
}

// GetLabels retrieves all labels
func (c *PlainClient) GetLabels(ctx context.Context) ([]*types.LabelType, error) {
	req := graphql.NewRequest(`
		query labels {
			labels {
				id
				name
				color
				createdAt {
					iso8601
				}
			}
		}
	`)

	c.setHeaders(req)

	var resp struct {
		Labels []*types.LabelType `json:"labels"`
	}

	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	return resp.Labels, nil
}

// CreateLabel creates a new label
func (c *PlainClient) CreateLabel(ctx context.Context, name, color string) (*types.LabelType, error) {
	req := graphql.NewRequest(`
		mutation createLabel($input: CreateLabelInput!) {
			createLabel(input: $input) {
				labelType {
					id
					name
					color
					createdAt {
						iso8601
					}
				}
				error {
					message
					type
				}
			}
		}
	`)

	input := map[string]interface{}{
		"name":  name,
		"color": color,
	}
	req.Var("input", input)
	c.setHeaders(req)

	var resp struct {
		CreateLabel struct {
			LabelType *types.LabelType `json:"labelType"`
			Error     *types.APIError  `json:"error"`
		} `json:"createLabel"`
	}

	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	if resp.CreateLabel.Error != nil {
		return nil, resp.CreateLabel.Error
	}

	return resp.CreateLabel.LabelType, nil
}

// GetThreadWithMessages retrieves a thread with its messages
func (c *PlainClient) GetThreadWithMessages(ctx context.Context, threadId string) (*types.Thread, error) {
	req := graphql.NewRequest(`
		query thread($threadId: ID!) {
			thread(threadId: $threadId) {
				id
				title
				status
				priority
				createdAt {
					iso8601
				}
				updatedAt {
					iso8601
				}
				customer {
					id
					fullName
					email {
						email
					}
					company {
						id
						name
					}
				}
				assignedTo {
					... on User {
						id
						fullName
						email
					}
				}
				timelineEntries {
					edges {
						node {
							id
							timestamp {
								iso8601
							}
							actor {
								... on UserActor {
									user {
										id
										fullName
										email
									}
								}
								... on CustomerActor {
									customer {
										id
										fullName
										email {
											email
										}
									}
								}
								... on SystemActor {
									systemId
								}
							}
							entry {
								... on EmailEntry {
									__typename
									emailId
									textContent
									from {
										name
										email
									}
									to {
										name
										email
									}
								}
								... on ChatEntry {
									__typename
									chatId
									chatText: text
								}
								... on NoteEntry {
									__typename
									noteId
									noteText: text
									markdown
									attachments {
										id
										fileName
										fileSize {
											bytes
											kiloBytes
											megaBytes
										}
										fileExtension
										fileMimeType
										type
									}
								}
								... on ThreadAssignmentTransitionedEntry {
									__typename
									previousAssignee {
										... on User {
											id
											fullName
											email
										}
									}
									nextAssignee {
										... on User {
											id
											fullName
											email
										}
									}
								}
								... on ThreadStatusTransitionedEntry {
									__typename
									previousStatus
									nextStatus
								}
								... on ThreadPriorityChangedEntry {
									__typename
									previousPriority
									nextPriority
								}
							}
						}
						cursor
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`)

	req.Var("threadId", threadId)
	c.setHeaders(req)

	var resp struct {
		Thread *types.Thread `json:"thread"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get thread with messages: %w", err)
	}

	return resp.Thread, nil
}

// SearchCustomers searches for customers by name
func (c *PlainClient) SearchCustomers(ctx context.Context, query string, limit int) ([]*types.Customer, error) {
	req := graphql.NewRequest(`
		query searchCustomers($query: String!, $first: Int!) {
			customers(first: $first, filters: { fullName: { contains: $query } }) {
				edges {
					node {
						id
						fullName
						email {
							email
						}
						status
						company {
							id
							name
						}
						createdAt {
							iso8601
						}
					}
					cursor
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`)

	req.Var("query", query)
	req.Var("first", limit)
	c.setHeaders(req)

	var resp struct {
		Customers *types.CustomerConnection `json:"customers"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	var customers []*types.Customer
	for _, edge := range resp.Customers.Edges {
		if edge.Node != nil {
			customers = append(customers, edge.Node)
		}
	}

	return customers, nil
}
