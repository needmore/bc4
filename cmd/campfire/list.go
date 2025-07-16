package campfire

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all campfires in the project",
		Long: `Display all campfires (chat rooms) in the current project with their status and activity.
		
Use --all to show campfires across all projects you have access to.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get required dependencies
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			projectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			campfireOps := client.Campfires()

			// Get campfires
			campfires, err := campfireOps.ListCampfires(f.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to list campfires: %w", err)
			}

			// Filter campfires to only show those in the current project unless --all is specified
			if !showAll {
				var projectCampfires []api.Campfire
				projectIDInt, _ := strconv.ParseInt(projectID, 10, 64)
				for _, cf := range campfires {
					if cf.Bucket.ID == projectIDInt {
						projectCampfires = append(projectCampfires, cf)
					}
				}
				campfires = projectCampfires
			}

			if len(campfires) == 0 {
				fmt.Println("No campfires found in this project.")
				return nil
			}

			// Get default campfire ID from config
			defaultCampfireID := ""
			if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
				if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
					defaultCampfireID = projDefaults.DefaultCampfire
				}
			}

			// Create table
			table := tableprinter.New(os.Stdout)

			// Add headers
			if showAll {
				if table.IsTTY() {
					table.AddHeader("ID", "NAME", "PROJECT", "STATUS", "LAST ACTIVITY")
				} else {
					table.AddHeader("ID", "NAME", "PROJECT", "STATUS", "STATE", "LAST_ACTIVITY")
				}
			} else {
				if table.IsTTY() {
					table.AddHeader("ID", "NAME", "STATUS", "LAST ACTIVITY")
				} else {
					table.AddHeader("ID", "NAME", "STATUS", "STATE", "LAST_ACTIVITY")
				}
			}

			// Sort campfires by updated_at (most recent first)
			for i := len(campfires) - 1; i >= 0; i-- {
				cf := campfires[i]
				idStr := strconv.FormatInt(cf.ID, 10)

				// Add ID field with default indicator
				if idStr == defaultCampfireID {
					if table.IsTTY() {
						table.AddIDField(idStr+"*", cf.Status) // Add asterisk for default
					} else {
						table.AddField(idStr)
					}
				} else {
					table.AddIDField(idStr, cf.Status)
				}

				// Add name field
				name := cf.Name
				if name == "" {
					name = "(untitled)"
				}
				table.AddProjectField(name, cf.Status)

				// Add project field if showing all
				if showAll {
					table.AddField(cf.Bucket.Name)
				}

				// Add status field
				if table.IsTTY() {
					table.AddColorField(cf.Status, cf.Status)
				} else {
					table.AddField(cf.Status)
				}

				// Add state column for non-TTY
				if !table.IsTTY() {
					table.AddField(cf.Status)
				}

				// Add last activity
				now := time.Now()
				table.AddTimeField(now, cf.UpdatedAt)

				table.EndRow()
			}

			// Render table
			return table.Render()
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show campfires from all projects")

	return cmd
}
