package utils

import (
	"sort"
	"strings"

	"github.com/needmore/bc4/internal/api"
)

// SortProjectsByName sorts a slice of projects alphabetically by name (case-insensitive)
func SortProjectsByName(projects []api.Project) {
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
}

