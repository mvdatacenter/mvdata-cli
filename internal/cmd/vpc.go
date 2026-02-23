package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newVPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpc",
		Short: "Manage VPCs",
	}
	cmd.AddCommand(newVPCCreateCmd())
	cmd.AddCommand(newVPCGetCmd())
	cmd.AddCommand(newVPCDeleteCmd())
	return cmd
}

func newVPCCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a VPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			vpc, err := client.CreateVPC(context.Background(), &sdk.VPC{Name: name})
			if err != nil {
				return fmt.Errorf("creating VPC: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, vpcOutput(vpc))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "VPC name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newVPCGetCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a VPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			vpc, err := client.GetVPC(context.Background(), name)
			if err != nil {
				return fmt.Errorf("getting VPC: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, vpcOutput(vpc))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "VPC name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newVPCDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a VPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteVPC(context.Background(), name); err != nil {
				return fmt.Errorf("deleting VPC: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "VPC %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "VPC name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func vpcOutput(vpc *sdk.VPC) any {
	if flagOutput == "json" {
		return vpc
	}
	return &output.Table{
		Headers: []string{"NAME", "CREATED"},
		Rows:    [][]string{{vpc.Name, vpc.CreatedAt}},
	}
}
