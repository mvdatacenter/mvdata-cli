package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mvdatacenter/mvdata-cli/internal/config"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var apiURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via browser-based SSO",
		Long:  "Initiates a device authorization flow. Opens a URL in your browser, authenticates via SSO, and saves the API key to your config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if apiURL == "" {
				apiURL = "https://console-api.mvdatacenter.com"
			}

			// Create an unauthenticated client for the device flow.
			client := sdk.New(apiURL, "")

			ctx := context.Background()
			auth, err := client.DeviceAuthorize(ctx)
			if err != nil {
				return fmt.Errorf("requesting device authorization: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Open this URL in your browser: %s\n", auth.VerificationURI)
			fmt.Fprintf(cmd.OutOrStdout(), "Enter code: %s\n\n", auth.UserCode)
			fmt.Fprintf(cmd.OutOrStdout(), "Waiting for authentication...")

			interval := time.Duration(auth.Interval) * time.Second
			if interval < time.Second {
				interval = 5 * time.Second
			}
			deadline := time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)

			for {
				if time.Now().After(deadline) {
					fmt.Fprintln(cmd.OutOrStdout())
					return fmt.Errorf("authentication timed out — please try again")
				}

				time.Sleep(interval)

				resp, err := client.DeviceToken(ctx, auth.DeviceCode)
				if err != nil {
					return fmt.Errorf("polling for token: %w", err)
				}

				switch resp.Status {
				case "pending":
					continue
				case "expired":
					fmt.Fprintln(cmd.OutOrStdout())
					return fmt.Errorf("device code expired — please try again")
				case "complete":
					profileName := flagProfile
					if profileName == "" {
						profileName = "default"
					}

					profile := &config.Profile{
						APIURL:   apiURL,
						APIToken: resp.APIToken,
					}
					if err := config.Save(profileName, profile); err != nil {
						return fmt.Errorf("saving config: %w", err)
					}

					fmt.Fprintf(cmd.OutOrStdout(), " done\n\n")
					fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s (account: %s, #%s)\n", resp.Email, resp.AccountName, resp.AccountNumber)
					fmt.Fprintf(cmd.OutOrStdout(), "API key saved to ~/.mvdata/config.json (profile: %s)\n", profileName)
					return nil
				default:
					fmt.Fprintln(cmd.OutOrStdout())
					return fmt.Errorf("unexpected status: %s", resp.Status)
				}
			}
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "MV Data API URL (default: https://console-api.mvdatacenter.com)")

	return cmd
}
