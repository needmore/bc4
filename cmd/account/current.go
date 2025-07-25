package account

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

func newCurrentCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "current",
		Short:   "Show current account",
		Long:    `Display information about the current default account.`,
		Aliases: []string{"whoami"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config from factory
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get auth client from factory
			authClient, err := f.AuthClient()
			if err != nil {
				return err
			}

			// Get default account
			defaultAccountID := authClient.GetDefaultAccount()
			if defaultAccountID == "" {
				fmt.Println("No default account set.")
				return nil
			}

			// Get account details
			accounts := authClient.GetAccounts()
			account, exists := accounts[defaultAccountID]
			if !exists {
				return fmt.Errorf("default account %s not found", defaultAccountID)
			}

			// Prepare output data
			type currentAccount struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Default bool   `json:"default"`
			}

			current := currentAccount{
				ID:      defaultAccountID,
				Name:    account.AccountName,
				Default: true,
			}

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(current)
			}

			// Display account details
			fmt.Println()
			fmt.Println(ui.TitleStyle.Render("Current Account"))
			fmt.Println()
			fmt.Printf("%s %s\n", ui.LabelStyle.Render("Name:"), ui.ValueStyle.Render(account.AccountName))
			fmt.Printf("%s %s\n", ui.LabelStyle.Render("ID:"), ui.ValueStyle.Render(defaultAccountID))

			// Show default project if set
			if cfg.DefaultProject != "" {
				fmt.Printf("%s %s\n", ui.LabelStyle.Render("Default Project:"), ui.ValueStyle.Render(cfg.DefaultProject))
			} else if cfg.Accounts != nil {
				if acc, ok := cfg.Accounts[defaultAccountID]; ok && acc.DefaultProject != "" {
					fmt.Printf("%s %s\n", ui.LabelStyle.Render("Default Project:"), ui.ValueStyle.Render(acc.DefaultProject))
				}
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
