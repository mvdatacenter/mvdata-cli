package cmd

import (
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/config"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

var (
	flagAPIURL   string
	flagAPIToken string
	flagOutput   string
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mvdata",
		Short: "CLI for managing MV Data cloud resources",
		Long:  "mvdata is a command-line interface for managing VPCs, subnets, instances, SSH keys, and Kubernetes clusters on MV Data.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table or json")
	root.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "MV Data API URL")
	root.PersistentFlags().StringVar(&flagAPIToken, "api-token", "", "MV Data API token")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newVPCCmd())
	root.AddCommand(newSubnetCmd())
	root.AddCommand(newInstanceCmd())
	root.AddCommand(newKeyCmd())
	root.AddCommand(newKubernetesCmd())
	root.AddCommand(newInstanceTypesCmd())

	return root
}

// newClient resolves config and returns an SDK client.
func newClient() (*sdk.Client, error) {
	cfg, err := config.Resolve(flagAPIURL, flagAPIToken)
	if err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}
	return sdk.New(cfg.APIURL, cfg.APIToken), nil
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
