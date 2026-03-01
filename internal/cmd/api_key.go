package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newAPIKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys",
	}
	cmd.AddCommand(newAPIKeyCreateCmd())
	cmd.AddCommand(newAPIKeyListCmd())
	cmd.AddCommand(newAPIKeyDeleteCmd())
	return cmd
}

func newAPIKeyCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			key, err := client.CreateAPIKey(context.Background(), &sdk.APIKeyCreate{Name: name})
			if err != nil {
				return fmt.Errorf("creating API key: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, apiKeyCreateOutput(key))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Key name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newAPIKeyListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			keys, err := client.ListAPIKeys(context.Background())
			if err != nil {
				return fmt.Errorf("listing API keys: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, apiKeyListOutput(keys))
		},
	}
	return cmd
}

func newAPIKeyDeleteCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteAPIKey(context.Background(), id); err != nil {
				return fmt.Errorf("deleting API key: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "API key %q deleted\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Key ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func apiKeyCreateOutput(k *sdk.APIKey) any {
	if flagOutput == "json" {
		return k
	}
	return &output.Table{
		Headers: []string{"ID", "NAME", "KEY", "PREFIX", "CREATED"},
		Rows:    [][]string{{k.ID, k.Name, k.Key, k.Prefix, k.CreatedAt}},
	}
}

func apiKeyListOutput(keys []sdk.APIKey) any {
	if flagOutput == "json" {
		return keys
	}
	rows := make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{k.ID, k.Name, k.Prefix, k.LastUsedAt, k.CreatedAt}
	}
	return &output.Table{
		Headers: []string{"ID", "NAME", "PREFIX", "LAST USED", "CREATED"},
		Rows:    rows,
	}
}
