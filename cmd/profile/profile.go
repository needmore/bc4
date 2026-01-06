package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

// NewProfileCmd creates the profile command
func NewProfileCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "profile",
		Short:   "Show current user profile",
		Long:    `Display your Basecamp profile information including name, email, and account details.`,
		Aliases: []string{"me"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get account ID for display
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			// Fetch profile from API
			profile, err := client.GetMyProfile(f.Context())
			if err != nil {
				return fmt.Errorf("failed to fetch profile: %w", err)
			}

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(profile)
			}

			// Display profile
			fmt.Println()
			fmt.Printf("👤 %s\n", ui.TitleStyle.Render(profile.Name))
			fmt.Println()

			if profile.EmailAddress != "" {
				fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Email:"), ui.ValueStyle.Render(profile.EmailAddress))
			}

			if profile.Title != "" {
				fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Title:"), ui.ValueStyle.Render(profile.Title))
			}

			if profile.Company != nil && profile.Company.Name != "" {
				fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Company:"), ui.ValueStyle.Render(profile.Company.Name))
			}

			fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Account:"), ui.ValueStyle.Render(accountID))

			// Show role if admin or owner
			if profile.Owner {
				fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Role:"), ui.ValueStyle.Render("Owner"))
			} else if profile.Admin {
				fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Role:"), ui.ValueStyle.Render("Admin"))
			}

			// Parse and display member since date
			if profile.CreatedAt != "" {
				if t, err := time.Parse(time.RFC3339, profile.CreatedAt); err == nil {
					fmt.Printf("   %s %s\n", ui.LabelStyle.Render("Member since:"), ui.ValueStyle.Render(t.Format("January 2006")))
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
