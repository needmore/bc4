package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type notifyOptions struct {
	jsonOutput bool
	responding *bool
	subscribed *bool
}

func newNotifyCmd(f *factory.Factory) *cobra.Command {
	opts := &notifyOptions{}
	var respondingStr, subscribedStr string

	cmd := &cobra.Command{
		Use:   "notify <question-id|URL>",
		Short: "Manage notification settings for a check-in question",
		Long: `Update your notification settings for a specific check-in question.

Settings:
  --responding=true/false  - Notify when someone responds to this question
  --subscribed=true/false  - Receive question notifications (reminders)`,
		Example: `  # Subscribe to responses
  bc4 checkin notify 12345 --responding=true

  # Unsubscribe from reminders
  bc4 checkin notify 12345 --subscribed=false

  # Update both settings
  bc4 checkin notify 12345 --responding=true --subscribed=false`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")

			// Parse boolean flags
			if cmd.Flags().Changed("responding") {
				b, err := strconv.ParseBool(respondingStr)
				if err != nil {
					return fmt.Errorf("invalid value for --responding: must be true or false")
				}
				opts.responding = &b
			}
			if cmd.Flags().Changed("subscribed") {
				b, err := strconv.ParseBool(subscribedStr)
				if err != nil {
					return fmt.Errorf("invalid value for --subscribed: must be true or false")
				}
				opts.subscribed = &b
			}

			return runNotify(f, opts, args)
		},
	}

	cmd.Flags().StringVar(&respondingStr, "responding", "", "Notify when someone responds (true/false)")
	cmd.Flags().StringVar(&subscribedStr, "subscribed", "", "Receive question notifications (true/false)")

	return cmd
}

func runNotify(f *factory.Factory, opts *notifyOptions, args []string) error {
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	questionOps := client.Questions()

	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Parse question ID (could be numeric ID or URL)
	parsedID, parsedURL, err := parser.ParseArgument(args[0])
	if err != nil {
		return fmt.Errorf("invalid question ID or URL: %s", args[0])
	}

	if parsedURL != nil {
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
		}
	}

	// Check if any settings were provided
	if opts.responding == nil && opts.subscribed == nil {
		// No settings provided, show current settings
		question, err := questionOps.GetQuestion(f.Context(), projectID, parsedID)
		if err != nil {
			return fmt.Errorf("failed to get question: %w", err)
		}

		if opts.jsonOutput {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(question.NotificationSettings)
		}

		fmt.Printf("Notification settings for question %d:\n", parsedID)
		if question.NotificationSettings != nil {
			fmt.Printf("  Responding: %v\n", question.NotificationSettings.Responding)
			fmt.Printf("  Subscribed: %v\n", question.NotificationSettings.Subscribed)
		} else {
			fmt.Println("  (no settings available)")
		}
		return nil
	}

	// Build update request
	req := api.NotificationSettingsUpdateRequest{
		Responding: opts.responding,
		Subscribed: opts.subscribed,
	}

	// Update settings
	settings, err := questionOps.UpdateNotificationSettings(f.Context(), projectID, parsedID, req)
	if err != nil {
		return fmt.Errorf("failed to update notification settings: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(settings)
	}

	fmt.Printf("Notification settings updated for question %d:\n", parsedID)
	fmt.Printf("  Responding: %v\n", settings.Responding)
	fmt.Printf("  Subscribed: %v\n", settings.Subscribed)

	return nil
}
