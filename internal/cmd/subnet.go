package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newSubnetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "Manage subnets",
	}
	cmd.AddCommand(newSubnetCreateCmd())
	cmd.AddCommand(newSubnetGetCmd())
	cmd.AddCommand(newSubnetDeleteCmd())
	return cmd
}

func newSubnetCreateCmd() *cobra.Command {
	var name, vpc, cidr string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a subnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			subnet, err := client.CreateSubnet(context.Background(), &sdk.Subnet{
				Name:      name,
				VPCName:   vpc,
				CIDRBlock: cidr,
			})
			if err != nil {
				return fmt.Errorf("creating subnet: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, subnetOutput(subnet))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Subnet name (required)")
	cmd.Flags().StringVar(&vpc, "vpc", "", "VPC name (required)")
	cmd.Flags().StringVar(&cidr, "cidr", "", "CIDR block (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("vpc")
	cmd.MarkFlagRequired("cidr")
	return cmd
}

func newSubnetGetCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a subnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			subnet, err := client.GetSubnet(context.Background(), name)
			if err != nil {
				return fmt.Errorf("getting subnet: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, subnetOutput(subnet))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Subnet name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newSubnetDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a subnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteSubnet(context.Background(), name); err != nil {
				return fmt.Errorf("deleting subnet: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Subnet %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Subnet name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func subnetOutput(s *sdk.Subnet) any {
	if flagOutput == "json" {
		return s
	}
	return &output.Table{
		Headers: []string{"NAME", "VPC", "CIDR", "CREATED"},
		Rows:    [][]string{{s.Name, s.VPCName, s.CIDRBlock, s.CreatedAt}},
	}
}
