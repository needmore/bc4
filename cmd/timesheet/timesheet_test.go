package timesheet

import (
	"testing"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimesheetCmd_Properties(t *testing.T) {
	f := factory.New()
	cmd := NewTimesheetCmd(f)

	assert.Equal(t, "timesheet", cmd.Use)
	assert.Equal(t, "View timesheet entries", cmd.Short)
	assert.Contains(t, cmd.Long, "timesheet entries")
	assert.Equal(t, []string{"time", "ts"}, cmd.Aliases)
}

func TestNewTimesheetCmd_Subcommands(t *testing.T) {
	f := factory.New()
	cmd := NewTimesheetCmd(f)

	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}
	assert.True(t, subcommands["list [project]"], "list subcommand should exist")
	assert.True(t, subcommands["report"], "report subcommand should exist")
}

func TestListCmd_Properties(t *testing.T) {
	f := factory.New()
	cmd := newListCmd(f)

	assert.Equal(t, "list [project]", cmd.Use)
	assert.Equal(t, "List timesheet entries", cmd.Short)
	assert.Contains(t, cmd.Long, "List timesheet entries")
	assert.Equal(t, []string{"ls"}, cmd.Aliases)
}

func TestListCmd_Flags(t *testing.T) {
	f := factory.New()
	cmd := newListCmd(f)

	flags := []struct {
		name      string
		shorthand string
	}{
		{"account", "a"},
		{"project", "p"},
		{"person", ""},
		{"since", ""},
		{"format", "f"},
		{"recording", ""},
	}

	for _, fl := range flags {
		flag := cmd.Flags().Lookup(fl.name)
		assert.NotNil(t, flag, "%s flag should exist", fl.name)
		if fl.shorthand != "" {
			assert.Equal(t, fl.shorthand, flag.Shorthand, "%s shorthand", fl.name)
		}
	}
}

func TestReportCmd_Properties(t *testing.T) {
	f := factory.New()
	cmd := newReportCmd(f)

	assert.Equal(t, "report", cmd.Use)
	assert.Equal(t, "Generate account-wide timesheet report", cmd.Short)
	assert.Contains(t, cmd.Long, "timesheet report across all projects")
}

func TestReportCmd_Flags(t *testing.T) {
	f := factory.New()
	cmd := newReportCmd(f)

	flags := []struct {
		name      string
		shorthand string
	}{
		{"account", "a"},
		{"project", "p"},
		{"person", ""},
		{"start", ""},
		{"end", ""},
		{"format", "f"},
		{"group-by", ""},
	}

	for _, fl := range flags {
		flag := cmd.Flags().Lookup(fl.name)
		assert.NotNil(t, flag, "%s flag should exist", fl.name)
		if fl.shorthand != "" {
			assert.Equal(t, fl.shorthand, flag.Shorthand, "%s shorthand", fl.name)
		}
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, result time.Time)
	}{
		{
			name:  "days duration",
			input: "7d",
			checkFunc: func(t *testing.T, result time.Time) {
				expected := time.Now().AddDate(0, 0, -7)
				assert.WithinDuration(t, expected, result, 2*time.Second)
			},
		},
		{
			name:  "hours duration",
			input: "24h",
			checkFunc: func(t *testing.T, result time.Time) {
				expected := time.Now().Add(-24 * time.Hour)
				assert.WithinDuration(t, expected, result, 2*time.Second)
			},
		},
		{
			name:  "ISO date",
			input: "2024-01-01",
			checkFunc: func(t *testing.T, result time.Time) {
				expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				assert.Equal(t, expected, result)
			},
		},
		{
			name:    "invalid input",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid day format",
			input:   "xd",
			wantErr: true,
		},
		{
			name:    "invalid hour format",
			input:   "xh",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSince(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.checkFunc(t, result)
		})
	}
}

func TestFilterEntries(t *testing.T) {
	entries := []api.TimesheetEntry{
		{
			Date:    "2024-01-15",
			Hours:   2.0,
			Creator: api.Person{ID: 1, Name: "Alice Johnson"},
			Bucket:  api.Bucket{ID: 100, Name: "Project A"},
		},
		{
			Date:    "2024-01-10",
			Hours:   3.5,
			Creator: api.Person{ID: 2, Name: "Bob Smith"},
			Bucket:  api.Bucket{ID: 200, Name: "Project B"},
		},
		{
			Date:    "2024-02-01",
			Hours:   1.0,
			Creator: api.Person{ID: 1, Name: "Alice Johnson"},
			Bucket:  api.Bucket{ID: 100, Name: "Project A"},
		},
	}

	t.Run("no filters", func(t *testing.T) {
		result := filterEntries(entries, "", time.Time{})
		assert.Len(t, result, 3)
	})

	t.Run("filter by person", func(t *testing.T) {
		result := filterEntries(entries, "alice", time.Time{})
		assert.Len(t, result, 2)
		for _, e := range result {
			assert.Contains(t, e.Creator.Name, "Alice")
		}
	})

	t.Run("filter by person case insensitive", func(t *testing.T) {
		result := filterEntries(entries, "BOB", time.Time{})
		assert.Len(t, result, 1)
		assert.Equal(t, "Bob Smith", result[0].Creator.Name)
	})

	t.Run("filter by date", func(t *testing.T) {
		since := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
		result := filterEntries(entries, "", since)
		assert.Len(t, result, 2)
	})

	t.Run("combined filters", func(t *testing.T) {
		since := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
		result := filterEntries(entries, "alice", since)
		assert.Len(t, result, 2)
		for _, e := range result {
			assert.Contains(t, e.Creator.Name, "Alice")
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := filterEntries(entries, "charlie", time.Time{})
		assert.Len(t, result, 0)
	})

	t.Run("empty input", func(t *testing.T) {
		result := filterEntries(nil, "", time.Time{})
		assert.Len(t, result, 0)
	})
}
