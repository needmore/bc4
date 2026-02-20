package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/needmore/bc4/internal/api"
)

// UserResolver helps resolve user identifiers to Person objects
type UserResolver struct {
	client    api.APIClient
	projectID string
	people    []api.Person
	cached    bool
}

// NewUserResolver creates a new user resolver for a project
func NewUserResolver(client api.APIClient, projectID string) *UserResolver {
	return &UserResolver{
		client:    client,
		projectID: projectID,
	}
}

// ResolveUsers resolves a list of user identifiers to person IDs
// Supports:
// - Email addresses: john@example.com
// - @mentions: @john (matches by name, case-insensitive)
// - Mixed: @john,jane@example.com
func (ur *UserResolver) ResolveUsers(ctx context.Context, identifiers []string) ([]int64, error) {
	// Ensure we have the people list cached
	if err := ur.ensurePeopleCached(ctx); err != nil {
		return nil, err
	}

	var personIDs []int64
	var notFound []string

	for _, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}

		personID, found := ur.resolveIdentifier(identifier)
		if found {
			// Avoid duplicates
			duplicate := false
			for _, id := range personIDs {
				if id == personID {
					duplicate = true
					break
				}
			}
			if !duplicate {
				personIDs = append(personIDs, personID)
			}
		} else {
			notFound = append(notFound, identifier)
		}
	}

	if len(notFound) > 0 {
		return nil, fmt.Errorf("could not find users: %s", strings.Join(notFound, ", "))
	}

	return personIDs, nil
}

// GetPeople returns the cached list of people in the project
func (ur *UserResolver) GetPeople(ctx context.Context) ([]api.Person, error) {
	if err := ur.ensurePeopleCached(ctx); err != nil {
		return nil, err
	}
	return ur.people, nil
}

// ResolvePeople resolves a list of user identifiers to full Person objects
func (ur *UserResolver) ResolvePeople(ctx context.Context, identifiers []string) ([]api.Person, error) {
	if err := ur.ensurePeopleCached(ctx); err != nil {
		return nil, err
	}

	var people []api.Person
	var notFound []string

	for _, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}

		personID, found := ur.resolveIdentifier(identifier)
		if !found {
			notFound = append(notFound, identifier)
			continue
		}

		for _, p := range ur.people {
			if p.ID == personID {
				people = append(people, p)
				break
			}
		}
	}

	if len(notFound) > 0 {
		return nil, fmt.Errorf("could not find users: %s", strings.Join(notFound, ", "))
	}

	return people, nil
}

// ensurePeopleCached loads the project people if not already cached
func (ur *UserResolver) ensurePeopleCached(ctx context.Context) error {
	if ur.cached {
		return nil
	}

	people, err := ur.client.GetProjectPeople(ctx, ur.projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project people: %w", err)
	}

	ur.people = people
	ur.cached = true
	return nil
}

// resolveIdentifier resolves a single identifier to a person ID
func (ur *UserResolver) resolveIdentifier(identifier string) (int64, bool) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return 0, false
	}

	// Check if it's an @mention
	if strings.HasPrefix(identifier, "@") {
		name := strings.TrimPrefix(identifier, "@")
		return ur.findByName(name)
	}

	// Check if it's an email
	if strings.Contains(identifier, "@") {
		return ur.findByEmail(identifier)
	}

	// Try both name and email
	if id, found := ur.findByName(identifier); found {
		return id, true
	}
	return ur.findByEmail(identifier)
}

// findByEmail finds a person by email address (case-insensitive)
func (ur *UserResolver) findByEmail(email string) (int64, bool) {
	email = strings.ToLower(strings.TrimSpace(email))
	for _, person := range ur.people {
		if strings.ToLower(person.EmailAddress) == email {
			return person.ID, true
		}
	}
	return 0, false
}

// findByName finds a person by name (case-insensitive, partial match)
func (ur *UserResolver) findByName(name string) (int64, bool) {
	name = strings.ToLower(strings.TrimSpace(name))

	// First try exact match
	for _, person := range ur.people {
		if strings.ToLower(person.Name) == name {
			return person.ID, true
		}
	}

	// Then try partial match (first name or last name)
	for _, person := range ur.people {
		personNameLower := strings.ToLower(person.Name)
		// Check if the search term matches any part of the name
		nameParts := strings.Fields(personNameLower)
		for _, part := range nameParts {
			if part == name {
				return person.ID, true
			}
		}
	}

	// Finally try contains match
	for _, person := range ur.people {
		if strings.Contains(strings.ToLower(person.Name), name) {
			return person.ID, true
		}
	}

	return 0, false
}
