package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateTime represents a Plain API datetime with ISO8601 format
type DateTime struct {
	ISO8601 string `json:"iso8601"`
}

// Time returns the parsed time
func (dt *DateTime) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, dt.ISO8601)
}

// String returns the ISO8601 string representation
func (dt *DateTime) String() string {
	return dt.ISO8601
}

// Email represents a Plain email object
type Email struct {
	Email string `json:"email"`
}

// Customer represents a Plain customer
type Customer struct {
	ID        string    `json:"id"`
	FullName  string    `json:"fullName"`
	Email     *Email    `json:"email"`
	Status    string    `json:"status"`
	Company   *Company  `json:"company"`
	CreatedAt *DateTime `json:"createdAt"`
	UpdatedAt *DateTime `json:"updatedAt"`
}

// CustomerEdge represents a customer edge in a connection
type CustomerEdge struct {
	Node   *Customer `json:"node"`
	Cursor string    `json:"cursor"`
}

// CustomerConnection represents a paginated connection of customers
type CustomerConnection struct {
	Edges    []*CustomerEdge `json:"edges"`
	PageInfo *PageInfo       `json:"pageInfo"`
}

// Thread represents a Plain thread
type Thread struct {
	ID              string                   `json:"id"`
	Title           string                   `json:"title"`
	Status          string                   `json:"status"`
	Priority        int                      `json:"priority"`
	Customer        *Customer                `json:"customer"`
	Assignee        *User                    `json:"assignee"`
	Messages        []*Message               `json:"messages"`
	TimelineEntries *TimelineEntryConnection `json:"timelineEntries"`
	Labels          []Label                  `json:"labels"`
	CreatedAt       *DateTime                `json:"createdAt"`
	UpdatedAt       *DateTime                `json:"updatedAt"`
}

// ThreadEdge represents a thread edge in a connection
type ThreadEdge struct {
	Node   *Thread `json:"node"`
	Cursor string  `json:"cursor"`
}

// ThreadConnection represents a paginated connection of threads
type ThreadConnection struct {
	Edges    []*ThreadEdge `json:"edges"`
	PageInfo *PageInfo     `json:"pageInfo"`
}

// Message represents a message in a thread
type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	SentBy    Actor     `json:"-"`
	CreatedAt *DateTime `json:"createdAt"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Message
func (m *Message) UnmarshalJSON(data []byte) error {
	// Create an auxiliary struct to handle the basic fields
	type Alias Message
	aux := &struct {
		SentBy json.RawMessage `json:"sentBy"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the sentBy field to determine its type
	var actorType struct {
		User        *User        `json:"user"`
		Customer    *Customer    `json:"customer"`
		CustomerID  *string      `json:"customerId"`
		SystemID    *string      `json:"systemId"`
		MachineUser *MachineUser `json:"machineUser"`
	}

	if err := json.Unmarshal(aux.SentBy, &actorType); err != nil {
		return err
	}

	// Assign the appropriate actor type
	if actorType.User != nil {
		m.SentBy = &UserActor{User: actorType.User}
	} else if actorType.Customer != nil {
		m.SentBy = &CustomerActor{Customer: actorType.Customer}
	} else if actorType.CustomerID != nil {
		m.SentBy = &DeletedCustomerActor{CustomerID: *actorType.CustomerID}
	} else if actorType.SystemID != nil {
		m.SentBy = &SystemActor{SystemID: *actorType.SystemID}
	} else if actorType.MachineUser != nil {
		m.SentBy = &MachineUserActor{MachineUser: actorType.MachineUser}
	} else {
		return fmt.Errorf("unknown actor type for message sentBy")
	}

	return nil
}

// TimelineEntry represents an entry in a thread's timeline
type TimelineEntry struct {
	ID        string    `json:"id"`
	Timestamp *DateTime `json:"timestamp"`
	Actor     Actor     `json:"-"`
	Entry     Entry     `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling for TimelineEntry
func (te *TimelineEntry) UnmarshalJSON(data []byte) error {
	// Create an auxiliary struct to handle the basic fields
	type Alias TimelineEntry
	aux := &struct {
		Actor json.RawMessage `json:"actor"`
		Entry json.RawMessage `json:"entry"`
		*Alias
	}{
		Alias: (*Alias)(te),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the actor field to determine its type
	var actorType struct {
		User        *User        `json:"user"`
		Customer    *Customer    `json:"customer"`
		CustomerID  *string      `json:"customerId"`
		SystemID    *string      `json:"systemId"`
		MachineUser *MachineUser `json:"machineUser"`
	}

	if err := json.Unmarshal(aux.Actor, &actorType); err != nil {
		return err
	}

	// Assign the appropriate actor type
	if actorType.User != nil {
		te.Actor = &UserActor{User: actorType.User}
	} else if actorType.Customer != nil {
		te.Actor = &CustomerActor{Customer: actorType.Customer}
	} else if actorType.CustomerID != nil {
		te.Actor = &DeletedCustomerActor{CustomerID: *actorType.CustomerID}
	} else if actorType.SystemID != nil {
		te.Actor = &SystemActor{SystemID: *actorType.SystemID}
	} else if actorType.MachineUser != nil {
		te.Actor = &MachineUserActor{MachineUser: actorType.MachineUser}
	} else {
		return fmt.Errorf("unknown actor type")
	}

	// Handle null or empty entry
	if len(aux.Entry) == 0 || string(aux.Entry) == "null" {
		te.Entry = nil
		return nil
	}

	// Parse the entry field to determine its type
	var entryType struct {
		EmailID             *string           `json:"emailId"`
		TextContent         *string           `json:"textContent"`
		From                *EmailParticipant `json:"from"`
		To                  *EmailParticipant `json:"to"`
		ChatID              *string           `json:"chatId"`
		Text                *string           `json:"text"`
		NoteID              *string           `json:"noteId"`
		Title               *string           `json:"title"`
		Components          []interface{}     `json:"components"`
		SlackMessageLink    *string           `json:"slackMessageLink"`
		SlackWebMessageLink *string           `json:"slackWebMessageLink"`
	}

	if err := json.Unmarshal(aux.Entry, &entryType); err != nil {
		return fmt.Errorf("failed to parse entry type: %w, raw data: %s", err, string(aux.Entry))
	}

	// Assign the appropriate entry type based on discriminator fields
	if entryType.EmailID != nil {
		emailEntry := &EmailEntry{}
		if err := json.Unmarshal(aux.Entry, emailEntry); err != nil {
			return fmt.Errorf("failed to unmarshal EmailEntry: %w", err)
		}
		te.Entry = emailEntry
	} else if entryType.ChatID != nil {
		chatEntry := &ChatEntry{}
		if err := json.Unmarshal(aux.Entry, chatEntry); err != nil {
			return fmt.Errorf("failed to unmarshal ChatEntry: %w", err)
		}
		te.Entry = chatEntry
	} else if entryType.NoteID != nil {
		noteEntry := &NoteEntry{}
		if err := json.Unmarshal(aux.Entry, noteEntry); err != nil {
			return fmt.Errorf("failed to unmarshal NoteEntry: %w", err)
		}
		te.Entry = noteEntry
	} else if entryType.Title != nil && entryType.Components != nil {
		customEntry := &CustomEntry{}
		if err := json.Unmarshal(aux.Entry, customEntry); err != nil {
			return fmt.Errorf("failed to unmarshal CustomEntry: %w", err)
		}
		te.Entry = customEntry
	} else if entryType.SlackMessageLink != nil || entryType.SlackWebMessageLink != nil {
		// Check for reply vs message based on additional context in raw JSON
		var rawSlack map[string]interface{}
		if err := json.Unmarshal(aux.Entry, &rawSlack); err != nil {
			return fmt.Errorf("failed to unmarshal Slack entry: %w", err)
		}

		// If it has thread/parent message indicators, treat as reply
		if _, hasThreadTs := rawSlack["threadTs"]; hasThreadTs {
			slackReplyEntry := &SlackReplyEntry{}
			if err := json.Unmarshal(aux.Entry, slackReplyEntry); err != nil {
				return fmt.Errorf("failed to unmarshal SlackReplyEntry: %w", err)
			}
			te.Entry = slackReplyEntry
		} else {
			slackMessageEntry := &SlackMessageEntry{}
			if err := json.Unmarshal(aux.Entry, slackMessageEntry); err != nil {
				return fmt.Errorf("failed to unmarshal SlackMessageEntry: %w", err)
			}
			te.Entry = slackMessageEntry
		}
	} else {
		// Handle all other entry types by storing the raw JSON as a map
		// This includes: ThreadAssignmentTransitionedEntry, ThreadStatusTransitionedEntry,
		// ThreadPriorityChangedEntry, SlackMessageEntry, ThreadLabelsChangedEntry,
		// ServiceLevelAgreementStatusTransitionedEntry, etc.
		var rawEntry map[string]interface{}
		if err := json.Unmarshal(aux.Entry, &rawEntry); err != nil {
			return fmt.Errorf("failed to unmarshal entry type: %w, raw data: %s", err, string(aux.Entry))
		}
		te.Entry = rawEntry
	}

	return nil
}

// TimelineEntryEdge represents a timeline entry edge in a connection
type TimelineEntryEdge struct {
	Node   *TimelineEntry `json:"node"`
	Cursor string         `json:"cursor"`
}

// TimelineEntryConnection represents a paginated connection of timeline entries
type TimelineEntryConnection struct {
	Edges    []*TimelineEntryEdge `json:"edges"`
	PageInfo *PageInfo            `json:"pageInfo"`
}

// Entry represents different types of timeline entries
type Entry interface{}

// EmailEntry represents an email timeline entry
type EmailEntry struct {
	EmailID     string            `json:"emailId"`
	TextContent string            `json:"textContent"`
	From        *EmailParticipant `json:"from"`
	To          *EmailParticipant `json:"to"`
}

// ChatEntry represents a chat timeline entry
type ChatEntry struct {
	ChatID string `json:"chatId"`
	Text   string `json:"text"`
}

// NoteEntry represents a note timeline entry
type NoteEntry struct {
	NoteID      string        `json:"noteId"`
	Text        string        `json:"text"`
	Markdown    string        `json:"markdown"`
	Attachments []*Attachment `json:"attachments"`
}

// CustomEntry represents a custom timeline entry
type CustomEntry struct {
	ExternalID  string                   `json:"externalId"`
	Title       string                   `json:"title"`
	Type        string                   `json:"type"`
	Components  []map[string]interface{} `json:"components"`
	Attachments []*Attachment            `json:"attachments"`
}

// ThreadAssignmentTransitionedEntry represents a thread assignment change
type ThreadAssignmentTransitionedEntry struct {
	PreviousAssignee interface{} `json:"previousAssignee"`
	NextAssignee     interface{} `json:"nextAssignee"`
}

// ThreadStatusTransitionedEntry represents a thread status change
type ThreadStatusTransitionedEntry struct {
	PreviousStatus       string      `json:"previousStatus"`
	PreviousStatusDetail interface{} `json:"previousStatusDetail"`
	NextStatus           string      `json:"nextStatus"`
	NextStatusDetail     interface{} `json:"nextStatusDetail"`
}

// ThreadPriorityChangedEntry represents a thread priority change
type ThreadPriorityChangedEntry struct {
	PreviousPriority int `json:"previousPriority"`
	NextPriority     int `json:"nextPriority"`
}

// SlackMessageEntry represents a Slack message timeline entry
type SlackMessageEntry struct {
	SlackMessageLink    string                   `json:"slackMessageLink"`
	SlackWebMessageLink string                   `json:"slackWebMessageLink"`
	Text                string                   `json:"text"`
	CustomerID          string                   `json:"customerId"`
	RelatedThread       interface{}              `json:"relatedThread"`
	Attachments         []*Attachment            `json:"attachments"`
	LastEditedOnSlackAt *DateTime                `json:"lastEditedOnSlackAt"`
	DeletedOnSlackAt    *DateTime                `json:"deletedOnSlackAt"`
	Reactions           []map[string]interface{} `json:"reactions"`
}

// SlackReplyEntry represents a Slack reply timeline entry
type SlackReplyEntry struct {
	SlackMessageLink    string                   `json:"slackMessageLink"`
	SlackWebMessageLink string                   `json:"slackWebMessageLink"`
	Text                string                   `json:"text"`
	CustomerID          string                   `json:"customerId"`
	RelatedThread       interface{}              `json:"relatedThread"`
	Attachments         []*Attachment            `json:"attachments"`
	LastEditedOnSlackAt *DateTime                `json:"lastEditedOnSlackAt"`
	DeletedOnSlackAt    *DateTime                `json:"deletedOnSlackAt"`
	Reactions           []map[string]interface{} `json:"reactions"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID            string    `json:"id"`
	FileName      string    `json:"fileName"`
	FileSize      *FileSize `json:"fileSize"`
	FileExtension string    `json:"fileExtension"`
	FileMimeType  string    `json:"fileMimeType"`
	Type          string    `json:"type"`
	CreatedAt     *DateTime `json:"createdAt"`
	UpdatedAt     *DateTime `json:"updatedAt"`
}

// FileSize represents file size information
type FileSize struct {
	Bytes     int     `json:"bytes"`
	KiloBytes float64 `json:"kiloBytes"`
	MegaBytes float64 `json:"megaBytes"`
}

// EmailParticipant represents an email participant
type EmailParticipant struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Actor represents either a User or Customer who sent a message
type Actor interface {
	GetID() string
	GetFullName() string
	GetEmail() string
}

// User represents a Plain user
type User struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// UserActor represents a user actor in timeline entries
type UserActor struct {
	User *User `json:"user"`
}

// CustomerActor represents a customer actor in timeline entries
type CustomerActor struct {
	Customer *Customer `json:"customer"`
}

// DeletedCustomerActor represents a deleted customer actor in timeline entries
type DeletedCustomerActor struct {
	CustomerID string `json:"customerId"`
}

// SystemActor represents a system actor in timeline entries
type SystemActor struct {
	SystemID string `json:"systemId"`
}

// MachineUserActor represents a machine user actor in timeline entries
type MachineUserActor struct {
	MachineUser *MachineUser `json:"machineUser"`
}

// MachineUser represents a Plain machine user
type MachineUser struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// GetID implements Actor interface for UserActor
func (ua *UserActor) GetID() string {
	if ua.User != nil {
		return ua.User.ID
	}
	return ""
}

// GetFullName implements Actor interface for UserActor
func (ua *UserActor) GetFullName() string {
	if ua.User != nil {
		return ua.User.FullName
	}
	return ""
}

// GetEmail implements Actor interface for UserActor
func (ua *UserActor) GetEmail() string {
	if ua.User != nil {
		return ua.User.Email
	}
	return ""
}

// GetID implements Actor interface for CustomerActor
func (ca *CustomerActor) GetID() string {
	if ca.Customer != nil {
		return ca.Customer.ID
	}
	return ""
}

// GetFullName implements Actor interface for CustomerActor
func (ca *CustomerActor) GetFullName() string {
	if ca.Customer != nil {
		return ca.Customer.FullName
	}
	return ""
}

// GetEmail implements Actor interface for CustomerActor
func (ca *CustomerActor) GetEmail() string {
	if ca.Customer != nil {
		return ca.Customer.GetEmail()
	}
	return ""
}

// GetID implements Actor interface for DeletedCustomerActor
func (dca *DeletedCustomerActor) GetID() string {
	return dca.CustomerID
}

// GetFullName implements Actor interface for DeletedCustomerActor
func (dca *DeletedCustomerActor) GetFullName() string {
	return "[Deleted Customer]"
}

// GetEmail implements Actor interface for DeletedCustomerActor
func (dca *DeletedCustomerActor) GetEmail() string {
	return ""
}

// GetID implements Actor interface for SystemActor
func (sa *SystemActor) GetID() string {
	return sa.SystemID
}

// GetFullName implements Actor interface for SystemActor
func (sa *SystemActor) GetFullName() string {
	return "System"
}

// GetEmail implements Actor interface for SystemActor
func (sa *SystemActor) GetEmail() string {
	return ""
}

// GetID implements Actor interface for MachineUserActor
func (mua *MachineUserActor) GetID() string {
	if mua.MachineUser != nil {
		return mua.MachineUser.ID
	}
	return ""
}

// GetFullName implements Actor interface for MachineUserActor
func (mua *MachineUserActor) GetFullName() string {
	if mua.MachineUser != nil {
		return mua.MachineUser.FullName
	}
	return ""
}

// GetEmail implements Actor interface for MachineUserActor
func (mua *MachineUserActor) GetEmail() string {
	if mua.MachineUser != nil {
		return mua.MachineUser.Email
	}
	return ""
}

// GetID implements Actor interface
func (u *User) GetID() string {
	return u.ID
}

// GetFullName implements Actor interface
func (u *User) GetFullName() string {
	return u.FullName
}

// GetEmail implements Actor interface
func (u *User) GetEmail() string {
	return u.Email
}

// GetID implements Actor interface
func (c *Customer) GetID() string {
	return c.ID
}

// GetFullName implements Actor interface
func (c *Customer) GetFullName() string {
	return c.FullName
}

// GetEmail implements Actor interface
func (c *Customer) GetEmail() string {
	if c.Email != nil {
		return c.Email.Email
	}
	return ""
}

type Label struct {
	LabelType LabelType `json:"labelType"`
}

// LabelType represents a Plain label
type LabelType struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      *string   `json:"icon"`
	CreatedAt *DateTime `json:"createdAt"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// APIError represents a Plain API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Company represents a Plain company
type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	CreatedAt *DateTime `json:"createdAt"`
	UpdatedAt *DateTime `json:"updatedAt"`
}

// Tenant represents a Plain tenant
type Tenant struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PublicName string    `json:"publicName"`
	CreatedAt  *DateTime `json:"createdAt"`
	UpdatedAt  *DateTime `json:"updatedAt"`
}

// Event represents a Plain event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt *DateTime              `json:"createdAt"`
}

// Tier represents a Plain tier
type Tier struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt *DateTime `json:"createdAt"`
	UpdatedAt *DateTime `json:"updatedAt"`
}
