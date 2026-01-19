package people

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var accountID string

	cmd := &cobra.Command{
		Use:   "view <person-id>",
		Short: "View person details",
		Long: `View detailed information about a specific person.

Displays the person's name, email, title, company, and role information.`,
		Aliases: []string{"show", "get"},
		Example: `  # View person details
  bc4 people view 12345

  # View as JSON
  bc4 people view 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse person ID
			personID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid person ID: %s", args[0])
			}

			// Apply overrides if specified
			f = f.ApplyOverrides(accountID, "")

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			peopleOps := client.People()

			// Fetch person
			person, err := peopleOps.GetPerson(f.Context(), personID)
			if err != nil {
				return fmt.Errorf("failed to fetch person: %w", err)
			}

			// Handle JSON output
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(person)
			}

			// Determine role
			role := "member"
			if person.Owner {
				role = "owner"
			} else if person.Admin {
				role = "admin"
			}

			// Create output table for key-value display
			table := tableprinter.New(os.Stdout)
			cs := table.GetColorScheme()

			// Display person details in a formatted way
			fmt.Printf("\n%s\n", cs.Bold(person.Name))
			fmt.Printf("%s\n\n", cs.Muted(fmt.Sprintf("#%d", person.ID)))

			// Email
			if person.EmailAddress != "" {
				fmt.Printf("Email:   %s\n", person.EmailAddress)
			}

			// Title
			if person.Title != "" {
				fmt.Printf("Title:   %s\n", person.Title)
			}

			// Company
			if person.Company != nil && person.Company.Name != "" {
				fmt.Printf("Company: %s\n", person.Company.Name)
			}

			// Role
			fmt.Printf("Role:    %s\n", role)

			// Created at
			if person.CreatedAt != "" {
				fmt.Printf("Joined:  %s\n", person.CreatedAt)
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}
