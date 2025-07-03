package types

import (
	"encoding/json"
	"testing"
)

func TestTimelineEntryActorUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantType string
		wantID   string
		wantName string
	}{
		{
			name: "UserActor",
			jsonData: `{
				"id": "te_123",
				"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
				"actor": {
					"user": {
						"id": "user_123",
						"fullName": "John Doe",
						"email": "john@example.com"
					}
				},
				"entry": null
			}`,
			wantType: "*types.UserActor",
			wantID:   "user_123",
			wantName: "John Doe",
		},
		{
			name: "CustomerActor",
			jsonData: `{
				"id": "te_124",
				"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
				"actor": {
					"customer": {
						"id": "customer_123",
						"fullName": "Jane Smith",
						"email": {"email": "jane@example.com"}
					}
				},
				"entry": null
			}`,
			wantType: "*types.CustomerActor",
			wantID:   "customer_123",
			wantName: "Jane Smith",
		},
		{
			name: "DeletedCustomerActor",
			jsonData: `{
				"id": "te_125",
				"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
				"actor": {
					"customerId": "deleted_customer_123"
				},
				"entry": null
			}`,
			wantType: "*types.DeletedCustomerActor",
			wantID:   "deleted_customer_123",
			wantName: "[Deleted Customer]",
		},
		{
			name: "SystemActor",
			jsonData: `{
				"id": "te_126",
				"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
				"actor": {
					"systemId": "system_123"
				},
				"entry": null
			}`,
			wantType: "*types.SystemActor",
			wantID:   "system_123",
			wantName: "System",
		},
		{
			name: "MachineUserActor",
			jsonData: `{
				"id": "te_127",
				"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
				"actor": {
					"machineUser": {
						"id": "machine_123",
						"fullName": "AI Assistant",
						"email": "ai@example.com"
					}
				},
				"entry": null
			}`,
			wantType: "*types.MachineUserActor",
			wantID:   "machine_123",
			wantName: "AI Assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry TimelineEntry
			err := json.Unmarshal([]byte(tt.jsonData), &entry)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if entry.Actor == nil {
				t.Fatal("Actor is nil")
			}

			// Check the type
			actualType := getTypeName(entry.Actor)
			if actualType != tt.wantType {
				t.Errorf("Expected actor type %s, got %s", tt.wantType, actualType)
			}

			// Check the ID
			if entry.Actor.GetID() != tt.wantID {
				t.Errorf("Expected actor ID %s, got %s", tt.wantID, entry.Actor.GetID())
			}

			// Check the name
			if entry.Actor.GetFullName() != tt.wantName {
				t.Errorf("Expected actor name %s, got %s", tt.wantName, entry.Actor.GetFullName())
			}
		})
	}
}

func TestMessageActorUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantType string
		wantID   string
		wantName string
	}{
		{
			name: "UserActor",
			jsonData: `{
				"id": "msg_123",
				"content": "Hello world",
				"sentBy": {
					"user": {
						"id": "user_123",
						"fullName": "John Doe",
						"email": "john@example.com"
					}
				},
				"createdAt": {"iso8601": "2023-01-01T00:00:00Z"}
			}`,
			wantType: "*types.UserActor",
			wantID:   "user_123",
			wantName: "John Doe",
		},
		{
			name: "MachineUserActor",
			jsonData: `{
				"id": "msg_124",
				"content": "Automated response",
				"sentBy": {
					"machineUser": {
						"id": "machine_123",
						"fullName": "AI Assistant",
						"email": "ai@example.com"
					}
				},
				"createdAt": {"iso8601": "2023-01-01T00:00:00Z"}
			}`,
			wantType: "*types.MachineUserActor",
			wantID:   "machine_123",
			wantName: "AI Assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message Message
			err := json.Unmarshal([]byte(tt.jsonData), &message)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if message.SentBy == nil {
				t.Fatal("SentBy is nil")
			}

			// Check the type
			actualType := getTypeName(message.SentBy)
			if actualType != tt.wantType {
				t.Errorf("Expected sentBy type %s, got %s", tt.wantType, actualType)
			}

			// Check the ID
			if message.SentBy.GetID() != tt.wantID {
				t.Errorf("Expected sentBy ID %s, got %s", tt.wantID, message.SentBy.GetID())
			}

			// Check the name
			if message.SentBy.GetFullName() != tt.wantName {
				t.Errorf("Expected sentBy name %s, got %s", tt.wantName, message.SentBy.GetFullName())
			}
		})
	}
}

func TestUnknownActorType(t *testing.T) {
	jsonData := `{
		"id": "te_128",
		"timestamp": {"iso8601": "2023-01-01T00:00:00Z"},
		"actor": {
			"unknownField": "some_value"
		},
		"entry": null
	}`

	var entry TimelineEntry
	err := json.Unmarshal([]byte(jsonData), &entry)
	if err == nil {
		t.Fatal("Expected error for unknown actor type, but got nil")
	}

	expectedError := "unknown actor type"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// Helper function to get the type name of an interface
func getTypeName(v interface{}) string {
	switch v.(type) {
	case *UserActor:
		return "*types.UserActor"
	case *CustomerActor:
		return "*types.CustomerActor"
	case *DeletedCustomerActor:
		return "*types.DeletedCustomerActor"
	case *SystemActor:
		return "*types.SystemActor"
	case *MachineUserActor:
		return "*types.MachineUserActor"
	default:
		return "unknown"
	}
}
