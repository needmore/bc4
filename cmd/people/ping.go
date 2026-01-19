package people

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

func newPingCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var formatStr string
	var accountID string

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "List people who can be pinged",
		Long: `List all people in the account who can be pinged.

Pings are private messages in Basecamp. This command shows all account
members who are available to receive pings from you.`,
		Aliases: []string{"pingable"},
		Example: `  # List pingable people
  bc4 people ping

  # Output as JSON
  bc4 people ping --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			f = f.ApplyOverrides(accountID, "")

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			peopleOps := client.People()

			// Fetch pingable people
			people, err := peopleOps.GetPingablePeople(f.Context())
			if err != nil {
				return fmt.Errorf("failed to fetch pingable people: %w", err)
			}

			// Sort people alphabetically by name
			sort.Slice(people, func(i, j int) bool {
				return strings.ToLower(people[i].Name) < strings.ToLower(people[j].Name)
			})

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle legacy JSON flag
			if jsonOutput {
				format = ui.OutputFormatJSON
			}

			// Handle JSON output
			if format == ui.OutputFormatJSON {
				return outputPingablePeopleJSON(people)
			}

			// Check if there are any people
			if len(people) == 0 {
				fmt.Println("No pingable people found.")
				return nil
			}

			// Render the people table without role column
			return renderPeopleTable(people, false)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or csv")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}

func outputPingablePeopleJSON(people []api.Person) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(people)
}
