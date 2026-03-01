package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage compute instances",
	}
	cmd.AddCommand(newInstanceCreateCmd())
	cmd.AddCommand(newInstanceGetCmd())
	cmd.AddCommand(newInstanceDeleteCmd())
	return cmd
}

func newInstanceCreateCmd() *cobra.Command {
	var name, vpc, instanceType, key string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			instance, err := client.CreateInstance(context.Background(), &sdk.Instance{
				Name:              name,
				VPCName:           vpc,
				InstanceType:      instanceType,
				AuthorizedKeyName: key,
			})
			if err != nil {
				return fmt.Errorf("creating instance: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, instanceOutput(instance))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Instance name (required)")
	cmd.Flags().StringVar(&vpc, "vpc", "", "VPC name (required)")
	cmd.Flags().StringVar(&instanceType, "type", "", "Instance type (required)")
	cmd.Flags().StringVar(&key, "key", "", "SSH key name (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("vpc")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newInstanceGetCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			instance, err := client.GetInstance(context.Background(), name)
			if err != nil {
				return fmt.Errorf("getting instance: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, instanceOutput(instance))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Instance name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newInstanceDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteInstance(context.Background(), name); err != nil {
				return fmt.Errorf("deleting instance: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Instance %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Instance name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func instanceOutput(i *sdk.Instance) any {
	if flagOutput == "json" {
		return i
	}
	return &output.Table{
		Headers: []string{"NAME", "TYPE", "KEY", "PRIVATE IP", "STATUS", "HOURLY PRICE", "CREATED"},
		Rows: [][]string{{
			i.Name,
			i.InstanceType,
			i.AuthorizedKeyName,
			i.PrivateIP,
			i.Status,
			fmt.Sprintf("%.2f", i.HourlyPrice),
			i.CreatedAt,
		}},
	}
}
