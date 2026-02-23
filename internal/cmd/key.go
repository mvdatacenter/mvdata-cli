package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage SSH keys",
	}
	cmd.AddCommand(newKeyCreateCmd())
	cmd.AddCommand(newKeyGetCmd())
	cmd.AddCommand(newKeyDeleteCmd())
	return cmd
}

func newKeyCreateCmd() *cobra.Command {
	var name, publicKey string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register an SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			key, err := client.CreateKey(context.Background(), &sdk.Key{
				Name: name,
				Key:  publicKey,
			})
			if err != nil {
				return fmt.Errorf("creating key: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, keyOutput(key))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Key name (required)")
	cmd.Flags().StringVar(&publicKey, "key", "", "Public key (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newKeyGetCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			key, err := client.GetKey(context.Background(), name)
			if err != nil {
				return fmt.Errorf("getting key: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, keyOutput(key))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Key name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newKeyDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteKey(context.Background(), name); err != nil {
				return fmt.Errorf("deleting key: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Key %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Key name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func keyOutput(k *sdk.Key) any {
	if flagOutput == "json" {
		return k
	}
	return &output.Table{
		Headers: []string{"NAME", "KEY", "CREATED"},
		Rows:    [][]string{{k.Name, k.Key, k.CreatedAt}},
	}
}
