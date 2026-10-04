package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mvdatacenter/mvdata-cli/internal/config"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

var (
	flagAPIURL   string
	flagAPIToken string
	flagOutput   string
	flagProfile  string
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "mvdata",
		Short:         "CLI for managing MV Data cloud resources",
		Long:          "mvdata is a command-line interface for managing VPCs, subnets, instances, SSH keys, Kubernetes clusters, and secrets on MV Data.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table or json")
	root.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "MV Data API URL")
	root.PersistentFlags().StringVar(&flagAPIToken, "api-token", "", "MV Data API token")
	root.PersistentFlags().StringVar(&flagProfile, "profile", "", "Config profile name")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newVPCCmd())
	root.AddCommand(newSubnetCmd())
	root.AddCommand(newInstanceCmd())
	root.AddCommand(newKeyCmd())
	root.AddCommand(newKubernetesCmd())
	root.AddCommand(newInstanceTypesCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newAPIKeyCmd())
	root.AddCommand(newConfigureCmd())
	root.AddCommand(newSecretStoreCmd())
	root.AddCommand(newSecretCmd())

	return root
}

// newClient resolves config and returns an SDK client.
func newClient() (*sdk.Client, error) {
	cfg, err := config.Resolve(flagAPIURL, flagAPIToken, flagProfile)
	if err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}
	return sdk.New(cfg.APIURL, cfg.APIToken), nil
}

func Run(args []string, stdout, stderr io.Writer) int {
	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", strings.TrimSpace(errorMessage(err)))
		return exitStatus(err)
	}
	return 0
}

var exitStatuses = []struct {
	sentinel error
	status   int
}{
	{sdk.ErrNotFound, 3},
	{sdk.ErrForbidden, 4},
	{sdk.ErrLimitExceeded, 5},
	{sdk.ErrConflict, 6},
	{sdk.ErrInvalid, 7},
	{sdk.ErrUnavailable, 8},
}

func exitStatus(err error) int {
	for _, s := range exitStatuses {
		if errors.Is(err, s.sentinel) {
			return s.status
		}
	}
	return 1
}

func errorMessage(err error) string {
	var limit *sdk.LimitExceededError
	if errors.As(err, &limit) {
		return fmt.Sprintf("%s (%s)", err.Error(), limit.Error())
	}
	return err.Error()
}

func Execute() int {
	return Run(os.Args[1:], os.Stdout, os.Stderr)
}
