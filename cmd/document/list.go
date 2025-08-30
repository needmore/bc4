package document

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project]",
		Short: "List documents",
		Long:  `List all documents in a project's document vault.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply project override if specified
			if len(args) > 0 {
				f = f.WithProject(args[0])
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			projectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get the vault for the project
			vault, err := client.GetVault(f.Context(), projectID)
			if err != nil {
				return err
			}

			// List documents
			documents, err := client.ListDocuments(f.Context(), projectID, vault.ID)
			if err != nil {
				return err
			}

			// Output format
			if viper.GetBool("json") {
				return json.NewEncoder(os.Stdout).Encode(documents)
			}

			// Terminal output
			if len(documents) == 0 {
				fmt.Println("No documents found in this project's vault.")
				return nil
			}

			// Display documents
			for _, doc := range documents {
				if ui.IsTerminal(os.Stdout) {
					fmt.Printf("📄 %s (#%d)\n", doc.Title, doc.ID)
					fmt.Printf("   Created: %s by %s\n", doc.CreatedAt.Format("2006-01-02 15:04"), doc.Creator.Name)
					if doc.UpdatedAt.After(doc.CreatedAt) {
						fmt.Printf("   Updated: %s\n", doc.UpdatedAt.Format("2006-01-02 15:04"))
					}
					if doc.CommentsCount > 0 {
						fmt.Printf("   Comments: %d\n", doc.CommentsCount)
					}
					fmt.Println()
				} else {
					fmt.Printf("%d\t%s\n", doc.ID, doc.Title)
				}
			}

			return nil
		},
	}

	return cmd
}
