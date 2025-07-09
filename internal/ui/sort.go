package ui

import (
	"sort"
	"strings"
)

// SortableByName is an interface for items that can be sorted by name
type SortableByName interface {
	GetName() string
}

// SortByName sorts a slice of items alphabetically by name (case-insensitive)
func SortByName[T SortableByName](items []T) {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].GetName()) < strings.ToLower(items[j].GetName())
	})
}

// SortStrings sorts a slice of strings alphabetically (case-insensitive)
func SortStrings(items []string) {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i]) < strings.ToLower(items[j])
	})
}