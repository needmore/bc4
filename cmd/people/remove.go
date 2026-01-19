package people

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
)

type removeOptions struct {
	projectID  string
	accountID  string
	jsonOutput bool
}

func newRemoveCmd(f *factory.Factory) *cobra.Command {
	opts := &removeOptions{}

	cmd := &cobra.Command{
		Use:   "remove <person-id>",
		Short: "Remove a person from a project",
		Long: `Remove a person's access from a Basecamp project.

This revokes the person's access to the specified project. They will no longer
be able to view or interact with the project content.

Note: This does not delete the person from the account, only removes their
access to the specific project.`,
		Aliases: []string{"rm", "revoke"},
		Example: `  # Remove a person from a project
  bc4 people remove 12345 --project 67890

  # Remove with default project
  bc4 people remove 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.projectID, "project", "p", "", "Project ID to remove the person from (required)")
	cmd.Flags().StringVarP(&opts.accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func runRemove(f *factory.Factory, opts *removeOptions, args []string) error {
	// Parse person ID
	personID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid person ID: %s", args[0])
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

	// Create the revoke request
	req := api.ProjectAccessUpdateRequest{
		Revoke: []int64{personID},
	}

	// Send the revoke request
	response, err := peopleOps.UpdateProjectAccess(f.Context(), resolvedProjectID, req)
	if err != nil {
		return fmt.Errorf("failed to remove person: %w", err)
	}

	// Handle JSON output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	}

	// Display result
	if len(response.Revoked) > 0 {
		for _, person := range response.Revoked {
			fmt.Printf("Removed %s (%d) from project %s\n", person.Name, person.ID, resolvedProjectID)
		}
	} else {
		fmt.Println("No access was revoked (person may not have had access)")
	}

	return nil
}
