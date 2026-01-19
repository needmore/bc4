package people

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
)

type inviteOptions struct {
	name        string
	email       string
	title       string
	companyName string
	projectID   string
	accountID   string
	jsonOutput  bool
}

func newInviteCmd(f *factory.Factory) *cobra.Command {
	opts := &inviteOptions{}

	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a new person to a project",
		Long: `Invite a new person to a Basecamp project.

This command creates a NEW user account and grants them access to the specified
project. The person will receive an email invitation to join Basecamp and the project.

Note: Use this command to invite people who don't yet have a Basecamp account.
If the person already has an account, use 'bc4 people update --grant' instead
to grant them access to the project.

You must specify both the name and email of the person to invite.
A project must be specified either via --project flag or default project.`,
		Example: `  # Invite a person to a project
  bc4 people invite --name "John Doe" --email john@example.com --project 12345

  # Invite with title and company
  bc4 people invite --name "Jane Smith" --email jane@example.com \
    --title "Developer" --company "Acme Inc" --project 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvite(f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Name of the person to invite (required)")
	cmd.Flags().StringVarP(&opts.email, "email", "e", "", "Email address of the person to invite (required)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Job title of the person")
	cmd.Flags().StringVar(&opts.companyName, "company", "", "Company name of the person")
	cmd.Flags().StringVarP(&opts.projectID, "project", "p", "", "Project ID to invite the person to (defaults to selected project)")
	cmd.Flags().StringVarP(&opts.accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output as JSON")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func runInvite(f *factory.Factory, opts *inviteOptions) error {
	// Validate required fields
	if opts.name == "" {
		return fmt.Errorf("name is required (use --name)")
	}
	if opts.email == "" {
		return fmt.Errorf("email is required (use --email)")
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

	// Create the invite request
	newPerson := api.ProjectAccessNewPerson{
		Name:         opts.name,
		EmailAddress: opts.email,
		Title:        opts.title,
		CompanyName:  opts.companyName,
	}

	req := api.ProjectAccessUpdateRequest{
		Create: []api.ProjectAccessNewPerson{newPerson},
	}

	// Send the invite
	response, err := peopleOps.UpdateProjectAccess(f.Context(), resolvedProjectID, req)
	if err != nil {
		return fmt.Errorf("failed to invite person: %w", err)
	}

	// Handle JSON output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	}

	// Display result
	if len(response.Granted) > 0 {
		fmt.Printf("Invited %s (%s) to project %s\n", opts.name, opts.email, resolvedProjectID)
		for _, person := range response.Granted {
			fmt.Printf("  Person ID: %d\n", person.ID)
		}
	} else {
		fmt.Println("Invitation sent, but no new access was granted (user may already have access)")
	}

	return nil
}
