package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/mvdatacenter/mvdata-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure a profile interactively",
		Long:  "Set the API URL and token for a profile. Uses the --profile flag to select which profile to configure (defaults to \"default\").",
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := flagProfile
			if profileName == "" {
				profileName = "default"
			}

			reader := bufio.NewReader(cmd.InOrStdin())

			fmt.Fprintf(cmd.OutOrStdout(), "API URL [https://console-api.mvdatacenter.com]: ")
			apiURL, _ := reader.ReadString('\n')
			apiURL = strings.TrimSpace(apiURL)
			if apiURL == "" {
				apiURL = "https://console-api.mvdatacenter.com"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "API Token: ")
			apiToken, _ := reader.ReadString('\n')
			apiToken = strings.TrimSpace(apiToken)

			if apiToken == "" {
				return fmt.Errorf("API token is required")
			}

			profile := &config.Profile{
				APIURL:   apiURL,
				APIToken: apiToken,
			}
			if err := config.Save(profileName, profile); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Profile %q saved to ~/.mvdata/config.json\n", profileName)
			return nil
		},
	}
	return cmd
}
