package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newInstanceTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance-types",
		Short: "List available instance types",
	}
	cmd.AddCommand(newInstanceTypesListCmd())
	return cmd
}

func newInstanceTypesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all instance types and pricing",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			types, err := client.ListInstanceTypes(context.Background())
			if err != nil {
				return fmt.Errorf("listing instance types: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, instanceTypesOutput(types))
		},
	}
}

func instanceTypesOutput(types []sdk.InstanceType) any {
	if flagOutput == "json" {
		return types
	}
	rows := make([][]string, len(types))
	for i, t := range types {
		rows[i] = []string{t.InstanceType, fmt.Sprintf("$%.2f", t.HourlyPrice)}
	}
	return &output.Table{
		Headers: []string{"INSTANCE TYPE", "HOURLY PRICE"},
		Rows:    rows,
	}
}
