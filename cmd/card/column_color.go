package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newColumnColorCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "color [COLUMN_ID or URL] COLOR",
		Short: "Set the color of a column",
		Long: `Set the color of a column.

Available colors:
  white, red, orange, yellow, green, blue, aqua, purple, gray, pink, brown

You can specify the column using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345")

Examples:
  bc4 card column color 123 blue
  bc4 card column color 123 green
  bc4 card column color https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345 red`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse column ID (could be numeric ID or URL)
			columnID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid column ID or URL: %s", args[0])
			}

			// Validate color
			color := strings.ToLower(args[1])
			validColors := []string{"white", "red", "orange", "yellow", "green", "blue", "aqua", "purple", "gray", "pink", "brown"}
			isValid := false
			for _, vc := range validColors {
				if color == vc {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("invalid color: %s. Available colors: %s", args[1], strings.Join(validColors, ", "))
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeColumn {
					return fmt.Errorf("URL is not for a column: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
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

			// Set the column color
			err = client.Columns().SetColumnColor(f.Context(), resolvedProjectID, columnID, color)
			if err != nil {
				return fmt.Errorf("failed to set column color: %w", err)
			}

			fmt.Printf("Column #%d color set to %s\n", columnID, color)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
