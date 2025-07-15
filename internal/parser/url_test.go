package parser

import (
	"testing"
)

func TestParseBasecampURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantAccount int64
		wantProject int64
		wantType    ResourceType
		wantID      int64
		wantParent  int64
		wantErr     bool
	}{
		// Project URLs
		{
			name:        "project URL",
			url:         "https://3.basecamp.com/1234567/projects/89012345",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeProject,
			wantID:      89012345,
		},
		{
			name:        "project URL with trailing slash",
			url:         "https://3.basecamp.com/1234567/projects/89012345/",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeProject,
			wantID:      89012345,
		},
		// Todo URLs
		{
			name:        "todo URL",
			url:         "https://3.basecamp.com/5624304/buckets/36656602/todos/8840769454",
			wantAccount: 5624304,
			wantProject: 36656602,
			wantType:    ResourceTypeTodo,
			wantID:      8840769454,
		},
		{
			name:        "todo API URL with .json",
			url:         "https://3.basecampapi.com/5624304/buckets/36656602/todos/8840769454.json",
			wantAccount: 5624304,
			wantProject: 36656602,
			wantType:    ResourceTypeTodo,
			wantID:      8840769454,
		},
		// Todo set URLs
		{
			name:        "todoset URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/todosets/34567890",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeTodoSet,
			wantID:      34567890,
		},
		// Card URLs
		{
			name:        "card URL",
			url:         "https://3.basecamp.com/5624304/buckets/36656602/card_tables/cards/8840769454",
			wantAccount: 5624304,
			wantProject: 36656602,
			wantType:    ResourceTypeCard,
			wantID:      8840769454,
		},
		{
			name:        "card table URL",
			url:         "https://3.basecamp.com/5624304/buckets/36656602/card_tables/8168120731",
			wantAccount: 5624304,
			wantProject: 36656602,
			wantType:    ResourceTypeCardTable,
			wantID:      8168120731,
		},
		// Column URLs
		{
			name:        "column URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/card_tables/34567890/columns/45678901",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeColumn,
			wantID:      45678901,
			wantParent:  34567890,
		},
		// Card step URLs
		{
			name:        "card step URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/34567890/steps/45678901",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeStep,
			wantID:      45678901,
			wantParent:  34567890,
		},
		// Campfire URLs
		{
			name:        "campfire URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/chats/34567890",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeCampfire,
			wantID:      34567890,
		},
		// Message URLs
		{
			name:        "message URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/messages/34567890",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeMessage,
			wantID:      34567890,
		},
		// Document URLs
		{
			name:        "document URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/documents/34567890",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeDocument,
			wantID:      34567890,
		},
		// Vault URLs
		{
			name:        "vault URL",
			url:         "https://3.basecamp.com/1234567/buckets/89012345/vaults/34567890",
			wantAccount: 1234567,
			wantProject: 89012345,
			wantType:    ResourceTypeVault,
			wantID:      34567890,
		},
		// Error cases
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "non-Basecamp URL",
			url:     "https://github.com/user/repo",
			wantErr: true,
		},
		{
			name:    "unrecognized pattern",
			url:     "https://3.basecamp.com/1234567/unknown/89012345",
			wantErr: true,
		},
		{
			name:    "missing IDs",
			url:     "https://3.basecamp.com/buckets/todos",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBasecampURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBasecampURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.AccountID != tt.wantAccount {
				t.Errorf("ParseBasecampURL() AccountID = %v, want %v", got.AccountID, tt.wantAccount)
			}
			if got.ProjectID != tt.wantProject {
				t.Errorf("ParseBasecampURL() ProjectID = %v, want %v", got.ProjectID, tt.wantProject)
			}
			if got.ResourceType != tt.wantType {
				t.Errorf("ParseBasecampURL() ResourceType = %v, want %v", got.ResourceType, tt.wantType)
			}
			if got.ResourceID != tt.wantID {
				t.Errorf("ParseBasecampURL() ResourceID = %v, want %v", got.ResourceID, tt.wantID)
			}
			if got.ParentID != tt.wantParent {
				t.Errorf("ParseBasecampURL() ParentID = %v, want %v", got.ParentID, tt.wantParent)
			}
		})
	}
}

func TestIsBasecampURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"basecamp.com URL", "https://3.basecamp.com/1234567/projects/89012345", true},
		{"basecampapi.com URL", "https://3.basecampapi.com/1234567/buckets/89012345/todos/34567890.json", true},
		{"partial URL", "basecamp.com/something", true},
		{"GitHub URL", "https://github.com/user/repo", false},
		{"numeric ID", "12345", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBasecampURL(tt.url); got != tt.want {
				t.Errorf("IsBasecampURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseArgument(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantID      int64
		wantParsed  bool
		wantAccount int64
		wantProject int64
		wantErr     bool
	}{
		{
			name:   "numeric ID",
			arg:    "12345",
			wantID: 12345,
		},
		{
			name:        "Basecamp URL",
			arg:         "https://3.basecamp.com/5624304/buckets/36656602/todos/8840769454",
			wantID:      8840769454,
			wantParsed:  true,
			wantAccount: 5624304,
			wantProject: 36656602,
		},
		{
			name:    "invalid numeric",
			arg:     "abc123",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			arg:     "https://github.com/user/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotParsed, err := ParseArgument(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if gotID != tt.wantID {
				t.Errorf("ParseArgument() ID = %v, want %v", gotID, tt.wantID)
			}
			if tt.wantParsed && gotParsed == nil {
				t.Errorf("ParseArgument() expected parsed URL but got nil")
			} else if tt.wantParsed {
				if gotParsed.AccountID != tt.wantAccount {
					t.Errorf("ParseArgument() AccountID = %v, want %v", gotParsed.AccountID, tt.wantAccount)
				}
				if gotParsed.ProjectID != tt.wantProject {
					t.Errorf("ParseArgument() ProjectID = %v, want %v", gotParsed.ProjectID, tt.wantProject)
				}
			}
		})
	}
}
