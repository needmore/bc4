package activity

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newWatchCmd(f *factory.Factory) *cobra.Command {
	var (
		accountID     string
		projectID     string
		recordingType string
		personStr     string
		interval      int
	)

	cmd := &cobra.Command{
		Use:   "watch [project]",
		Short: "Watch for real-time project activity",
		Long: `Watch for real-time activity and changes across a Basecamp project.

This command polls the activity feed at regular intervals and displays new items
as they appear. Press Ctrl+C to stop watching.`,
		Aliases: []string{"w"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse project argument if provided (could be URL or ID)
			if len(args) > 0 {
				if parser.IsBasecampURL(args[0]) {
					parsed, err := parser.ParseBasecampURL(args[0])
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.ResourceType != parser.ResourceTypeProject {
						return fmt.Errorf("URL is not for a project: %s", args[0])
					}
					// Override factory with URL-provided values
					if parsed.AccountID > 0 {
						f = f.WithAccount(strconv.FormatInt(parsed.AccountID, 10))
					}
					if parsed.ProjectID > 0 {
						f = f.WithProject(strconv.FormatInt(parsed.ProjectID, 10))
					}
				} else {
					// Treat as project ID
					f = f.WithProject(args[0])
				}
			}

			// Apply flag overrides (flags take precedence over URL)
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get project name for display
			project, err := client.GetProject(cmd.Context(), resolvedProjectID)
			if err != nil {
				return err
			}

			// Build options
			opts := &api.ActivityListOptions{
				Limit: 10, // Only show recent items in watch mode
			}

			// Parse type filter
			if recordingType != "" {
				opts.RecordingTypes = parseTypes(recordingType)
			}

			// Parse person filter
			if personStr != "" {
				personID, err := parsePersonIdentifier(client, cmd.Context(), personStr)
				if err != nil {
					return fmt.Errorf("invalid --person value: %w", err)
				}
				opts.PersonID = personID
			}

			// Start watching
			return watchActivity(cmd.Context(), client, resolvedProjectID, project.Name, opts, time.Duration(interval)*time.Second)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVarP(&recordingType, "type", "t", "", "Filter by type: todo, message, document, comment, upload")
	cmd.Flags().StringVar(&personStr, "person", "", "Filter by person (ID, name, or email)")
	cmd.Flags().IntVarP(&interval, "interval", "i", 30, "Polling interval in seconds")

	return cmd
}

// watchActivity continuously polls for new activity and displays it
func watchActivity(ctx context.Context, client *api.ModularClient, projectID string, projectName string, opts *api.ActivityListOptions, interval time.Duration) error {
	fmt.Printf("Watching activity in project: %s\n", projectName)
	fmt.Printf("Polling every %v (Press Ctrl+C to stop)\n\n", interval)

	// Track the last seen timestamp to avoid showing duplicates
	var lastSeen time.Time

	// Initial fetch to establish baseline
	recordings, err := client.ListRecordings(ctx, projectID, opts)
	if err != nil {
		return err
	}

	if len(recordings) > 0 {
		// Show initial state
		fmt.Println("Recent activity:")
		for i := len(recordings) - 1; i >= 0; i-- {
			displayActivityItem(recordings[i])
		}
		lastSeen = recordings[0].UpdatedAt
		fmt.Println()
	}

	// Poll for updates
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nStopping watch...")
			return nil
		case <-ticker.C:
			// Update the since time to only get new items
			since := lastSeen
			opts.Since = &since

			recordings, err := client.ListRecordings(ctx, projectID, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching activity: %v\n", err)
				continue
			}

			// Display new items (in reverse chronological order, oldest first)
			if len(recordings) > 0 {
				for i := len(recordings) - 1; i >= 0; i-- {
					if recordings[i].UpdatedAt.After(lastSeen) {
						displayActivityItem(recordings[i])
						lastSeen = recordings[i].UpdatedAt
					}
				}
			}
		}
	}
}

// displayActivityItem displays a single activity item in a compact format
func displayActivityItem(r api.Recording) {
	timestamp := r.UpdatedAt.Format("15:04:05")
	typeLabel := formatRecordingType(r.Type)

	// Build the display line
	line := fmt.Sprintf("[%s] %s", timestamp, typeLabel)

	// Add creator
	line = fmt.Sprintf("%s by %s:", line, r.Creator.Name)

	// Add title
	title := r.Title
	if len(title) > 80 {
		title = title[:77] + "..."
	}
	line = fmt.Sprintf("%s %s", line, title)

	// Add parent context if available
	if r.Parent != nil {
		parentTitle := r.Parent.Title
		if len(parentTitle) > 40 {
			parentTitle = parentTitle[:37] + "..."
		}
		line = fmt.Sprintf("%s (in %s)", line, parentTitle)
	}

	fmt.Println(line)
}
