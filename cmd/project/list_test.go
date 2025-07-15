package project

import (
	"reflect"
	"testing"

	"github.com/needmore/bc4/internal/api"
)

func TestSortProjectsByName(t *testing.T) {
	tests := []struct {
		name     string
		input    []api.Project
		expected []api.Project
	}{
		{
			name: "Sort projects alphabetically",
			input: []api.Project{
				{Name: "Zebra"},
				{Name: "Apple"},
				{Name: "Banana"},
			},
			expected: []api.Project{
				{Name: "Apple"},
				{Name: "Banana"},
				{Name: "Zebra"},
			},
		},
		{
			name: "Case-insensitive sorting",
			input: []api.Project{
				{Name: "zebra"},
				{Name: "APPLE"},
				{Name: "Banana"},
			},
			expected: []api.Project{
				{Name: "APPLE"},
				{Name: "Banana"},
				{Name: "zebra"},
			},
		},
		{
			name:     "Empty slice",
			input:    []api.Project{},
			expected: []api.Project{},
		},
		{
			name: "Single project",
			input: []api.Project{
				{Name: "Project A"},
			},
			expected: []api.Project{
				{Name: "Project A"},
			},
		},
		{
			name: "Already sorted",
			input: []api.Project{
				{Name: "Alpha"},
				{Name: "Beta"},
				{Name: "Gamma"},
			},
			expected: []api.Project{
				{Name: "Alpha"},
				{Name: "Beta"},
				{Name: "Gamma"},
			},
		},
		{
			name: "Reverse sorted",
			input: []api.Project{
				{Name: "Gamma"},
				{Name: "Beta"},
				{Name: "Alpha"},
			},
			expected: []api.Project{
				{Name: "Alpha"},
				{Name: "Beta"},
				{Name: "Gamma"},
			},
		},
		{
			name: "Projects with spaces and special characters",
			input: []api.Project{
				{Name: "Project Z"},
				{Name: "Project A"},
				{Name: "!Important"},
				{Name: "_Underscore"},
			},
			expected: []api.Project{
				{Name: "!Important"},
				{Name: "_Underscore"},
				{Name: "Project A"},
				{Name: "Project Z"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test data
			projects := make([]api.Project, len(tt.input))
			copy(projects, tt.input)

			sortProjectsByName(projects)

			if !reflect.DeepEqual(projects, tt.expected) {
				t.Errorf("sortProjectsByName() = %v, want %v", projects, tt.expected)
			}
		})
	}
}

func BenchmarkSortProjectsByName(b *testing.B) {
	// Create a slice of projects for benchmarking
	projects := []api.Project{
		{Name: "Zulu"},
		{Name: "Yankee"},
		{Name: "X-ray"},
		{Name: "Whiskey"},
		{Name: "Victor"},
		{Name: "Uniform"},
		{Name: "Tango"},
		{Name: "Sierra"},
		{Name: "Romeo"},
		{Name: "Quebec"},
		{Name: "Papa"},
		{Name: "Oscar"},
		{Name: "November"},
		{Name: "Mike"},
		{Name: "Lima"},
		{Name: "Kilo"},
		{Name: "Juliet"},
		{Name: "India"},
		{Name: "Hotel"},
		{Name: "Golf"},
		{Name: "Foxtrot"},
		{Name: "Echo"},
		{Name: "Delta"},
		{Name: "Charlie"},
		{Name: "Bravo"},
		{Name: "Alpha"},
	}

	for i := 0; i < b.N; i++ {
		// Make a copy for each iteration
		benchProjects := make([]api.Project, len(projects))
		copy(benchProjects, projects)
		sortProjectsByName(benchProjects)
	}
}
