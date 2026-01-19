package people

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
)

type updateOptions struct {
	grant      []string
	revoke     []string
	projectID  string
	accountID  string
	jsonOutput bool
}

func newUpdateCmd(f *factory.Factory) *cobra.Command {
	opts := &updateOptions{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project access for people",
		Long: `Update who has access to a Basecamp project.

Use --grant to give existing account members access to a project.
Use --revoke to remove access from people currently in the project.

Multiple person IDs can be specified by using the flag multiple times
or by comma-separating the IDs.`,
		Aliases: []string{"access"},
		Example: `  # Grant access to a project
  bc4 people update --grant 12345 --project 67890

  # Grant access to multiple people
  bc4 people update --grant 12345 --grant 12346 --project 67890

  # Grant and revoke in one command
  bc4 people update --grant 12345 --revoke 12346 --project 67890

  # Using comma-separated IDs
  bc4 people update --grant 12345,12346,12347 --project 67890`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(f, opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.grant, "grant", nil, "Person IDs to grant access (can be used multiple times)")
	cmd.Flags().StringSliceVar(&opts.revoke, "revoke", nil, "Person IDs to revoke access (can be used multiple times)")
	cmd.Flags().StringVarP(&opts.projectID, "project", "p", "", "Project ID to update access for (required)")
	cmd.Flags().StringVarP(&opts.accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func runUpdate(f *factory.Factory, opts *updateOptions) error {
	// Validate that at least one operation is specified
	if len(opts.grant) == 0 && len(opts.revoke) == 0 {
		return fmt.Errorf("at least one of --grant or --revoke must be specified")
	}

	// Parse person IDs
	grantIDs, err := parsePersonIDs(opts.grant)
	if err != nil {
		return fmt.Errorf("invalid grant person ID: %w", err)
	}

	revokeIDs, err := parsePersonIDs(opts.revoke)
	if err != nil {
		return fmt.Errorf("invalid revoke person ID: %w", err)
	}

	// Apply overrides if specified
	f = f.ApplyOverrides(opts.accountID, opts.projectID)

	// Get resolved project ID
	resolvedProjectID, err := f.ProjectID()
	if err != nil {
		return fmt.Errorf("project is required: use --project flag or set a default project")
	}

	// Get API client from factory
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	peopleOps := client.People()

	// Create the update request
	req := api.ProjectAccessUpdateRequest{
		Grant:  grantIDs,
		Revoke: revokeIDs,
	}

	// Send the update request
	response, err := peopleOps.UpdateProjectAccess(f.Context(), resolvedProjectID, req)
	if err != nil {
		return fmt.Errorf("failed to update project access: %w", err)
	}

	// Handle JSON output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	}

	// Display results
	if len(response.Granted) > 0 {
		fmt.Println("Granted access:")
		for _, person := range response.Granted {
			fmt.Printf("  + %s (%d)\n", person.Name, person.ID)
		}
	}

	if len(response.Revoked) > 0 {
		fmt.Println("Revoked access:")
		for _, person := range response.Revoked {
			fmt.Printf("  - %s (%d)\n", person.Name, person.ID)
		}
	}

	if len(response.Granted) == 0 && len(response.Revoked) == 0 {
		fmt.Println("No changes made (users may already have the specified access)")
	}

	return nil
}

// parsePersonIDs parses a slice of strings that may contain comma-separated IDs
func parsePersonIDs(input []string) ([]int64, error) {
	var ids []int64
	for _, s := range input {
		// Handle comma-separated values
		parts := strings.Split(s, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("'%s' is not a valid person ID", part)
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}
