package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ResourceType represents the type of Basecamp resource
type ResourceType string

const (
	ResourceTypeProject       ResourceType = "project"
	ResourceTypeTodo          ResourceType = "todo"
	ResourceTypeTodoSet       ResourceType = "todoset"
	ResourceTypeTodoList      ResourceType = "todolist"
	ResourceTypeCard          ResourceType = "card"
	ResourceTypeCardTable     ResourceType = "card_table"
	ResourceTypeColumn        ResourceType = "column"
	ResourceTypeStep          ResourceType = "step"
	ResourceTypeCampfire      ResourceType = "campfire"
	ResourceTypeMessage       ResourceType = "message"
	ResourceTypeDocument      ResourceType = "document"
	ResourceTypeVault         ResourceType = "vault"
	ResourceTypeSchedule      ResourceType = "schedule"
	ResourceTypeQuestionnaire ResourceType = "questionnaire"
	ResourceTypeUnknown       ResourceType = "unknown"
)

// ParsedURL represents the extracted information from a Basecamp URL
type ParsedURL struct {
	AccountID    int64
	ProjectID    int64
	ResourceType ResourceType
	ResourceID   int64
	ParentID     int64 // For nested resources (e.g., card table ID for cards)
}

// urlPattern defines a pattern for matching Basecamp URLs
type urlPattern struct {
	regex        *regexp.Regexp
	resourceType ResourceType
	extractor    func(matches []string) (*ParsedURL, error)
}

var urlPatterns = []urlPattern{
	// Project pattern: /1234567/projects/89012345
	{
		regex:        regexp.MustCompile(`^/(\d+)/projects/(\d+)`),
		resourceType: ResourceTypeProject,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeProject,
				ResourceID:   projectID,
			}, nil
		},
	},
	// Todo pattern: /1234567/buckets/89012345/todos/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/todos/(\d+)`),
		resourceType: ResourceTypeTodo,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			todoID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeTodo,
				ResourceID:   todoID,
			}, nil
		},
	},
	// Todo set pattern: /1234567/buckets/89012345/todosets/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/todosets/(\d+)$`),
		resourceType: ResourceTypeTodoSet,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			todoSetID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeTodoSet,
				ResourceID:   todoSetID,
			}, nil
		},
	},
	// Todo list pattern: /1234567/buckets/89012345/todosets/34567890/todolists/45678901
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/todosets/(\d+)/todolists/(\d+)`),
		resourceType: ResourceTypeTodoList,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			todoSetID, _ := strconv.ParseInt(matches[3], 10, 64)
			todoListID, _ := strconv.ParseInt(matches[4], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeTodoList,
				ResourceID:   todoListID,
				ParentID:     todoSetID,
			}, nil
		},
	},
	// Card step pattern: /1234567/buckets/89012345/card_tables/cards/34567890/steps/45678901
	// NOTE: This must come before the general card pattern to match correctly
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/card_tables/cards/(\d+)/steps/(\d+)`),
		resourceType: ResourceTypeStep,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			cardID, _ := strconv.ParseInt(matches[3], 10, 64)
			stepID, _ := strconv.ParseInt(matches[4], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeStep,
				ResourceID:   stepID,
				ParentID:     cardID,
			}, nil
		},
	},
	// Card pattern: /1234567/buckets/89012345/card_tables/cards/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/card_tables/cards/(\d+)$`),
		resourceType: ResourceTypeCard,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			cardID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeCard,
				ResourceID:   cardID,
			}, nil
		},
	},
	// Card table pattern: /1234567/buckets/89012345/card_tables/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/card_tables/(\d+)$`),
		resourceType: ResourceTypeCardTable,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			tableID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeCardTable,
				ResourceID:   tableID,
			}, nil
		},
	},
	// Column pattern: /1234567/buckets/89012345/card_tables/34567890/columns/45678901
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/card_tables/(\d+)/columns/(\d+)`),
		resourceType: ResourceTypeColumn,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			tableID, _ := strconv.ParseInt(matches[3], 10, 64)
			columnID, _ := strconv.ParseInt(matches[4], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeColumn,
				ResourceID:   columnID,
				ParentID:     tableID,
			}, nil
		},
	},
	// Campfire pattern: /1234567/buckets/89012345/chats/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/chats/(\d+)`),
		resourceType: ResourceTypeCampfire,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			campfireID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeCampfire,
				ResourceID:   campfireID,
			}, nil
		},
	},
	// Message pattern: /1234567/buckets/89012345/messages/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/messages/(\d+)`),
		resourceType: ResourceTypeMessage,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			messageID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeMessage,
				ResourceID:   messageID,
			}, nil
		},
	},
	// Document pattern: /1234567/buckets/89012345/documents/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/documents/(\d+)`),
		resourceType: ResourceTypeDocument,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			documentID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeDocument,
				ResourceID:   documentID,
			}, nil
		},
	},
	// Vault pattern: /1234567/buckets/89012345/vaults/34567890
	{
		regex:        regexp.MustCompile(`^/(\d+)/buckets/(\d+)/vaults/(\d+)`),
		resourceType: ResourceTypeVault,
		extractor: func(matches []string) (*ParsedURL, error) {
			accountID, _ := strconv.ParseInt(matches[1], 10, 64)
			projectID, _ := strconv.ParseInt(matches[2], 10, 64)
			vaultID, _ := strconv.ParseInt(matches[3], 10, 64)
			return &ParsedURL{
				AccountID:    accountID,
				ProjectID:    projectID,
				ResourceType: ResourceTypeVault,
				ResourceID:   vaultID,
			}, nil
		},
	},
}

// ParseBasecampURL parses a Basecamp URL and extracts relevant IDs
func ParseBasecampURL(inputURL string) (*ParsedURL, error) {
	// Parse the URL
	u, err := url.Parse(inputURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check if it's a Basecamp URL
	if !strings.Contains(u.Host, "basecamp.com") && !strings.Contains(u.Host, "basecampapi.com") {
		return nil, fmt.Errorf("not a Basecamp URL: %s", u.Host)
	}

	// Remove .json extension if present (API URLs)
	path := strings.TrimSuffix(u.Path, ".json")

	// Try to match against each pattern
	for _, pattern := range urlPatterns {
		if matches := pattern.regex.FindStringSubmatch(path); matches != nil {
			return pattern.extractor(matches)
		}
	}

	return nil, fmt.Errorf("unrecognized Basecamp URL pattern: %s", path)
}

// IsBasecampURL checks if a string looks like a Basecamp URL
func IsBasecampURL(s string) bool {
	return strings.Contains(s, "basecamp.com") || strings.Contains(s, "basecampapi.com")
}

// ParseArgument parses a command argument that could be either a numeric ID or a Basecamp URL
func ParseArgument(arg string) (int64, *ParsedURL, error) {
	// Check if it's a URL
	if IsBasecampURL(arg) {
		parsed, err := ParseBasecampURL(arg)
		if err != nil {
			return 0, nil, err
		}
		return parsed.ResourceID, parsed, nil
	}

	// Try to parse as numeric ID
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("argument must be a numeric ID or Basecamp URL: %s", arg)
	}

	return id, nil, nil
}

