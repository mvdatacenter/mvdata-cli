package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newSecretStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-store",
		Aliases: []string{"secret-stores"},
		Short:   "Manage secret stores",
	}
	cmd.AddCommand(newSecretStoreListCmd())
	cmd.AddCommand(newSecretStoreCreateCmd())
	cmd.AddCommand(newSecretStoreGetCmd())
	cmd.AddCommand(newSecretStoreUpdateCmd())
	cmd.AddCommand(newSecretStoreDeleteCmd())
	return cmd
}

func newSecretStoreListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List secret stores",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			stores, err := client.ListSecretStores(context.Background())
			if err != nil {
				return fmt.Errorf("listing secret stores: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretStoresOutput(stores))
		},
	}
}

func newSecretStoreCreateCmd() *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "create <store>",
		Short: "Create a secret store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			store, err := client.CreateSecretStore(context.Background(), &sdk.SecretStore{Name: args[0], Description: description})
			if err != nil {
				return fmt.Errorf("creating secret store: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretStoreOutput(store))
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Store description")
	return cmd
}

func newSecretStoreGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <store>",
		Short: "Get a secret store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			store, err := client.GetSecretStore(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("getting secret store: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretStoreOutput(store))
		},
	}
}

func newSecretStoreUpdateCmd() *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "update <store>",
		Short: "Update a secret store's description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			store, err := client.UpdateSecretStore(context.Background(), args[0], &sdk.SecretStoreUpdate{Description: description})
			if err != nil {
				return fmt.Errorf("updating secret store: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretStoreOutput(store))
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Store description (required)")
	cmd.MarkFlagRequired("description")
	return cmd
}

func newSecretStoreDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <store>",
		Short: "Delete a secret store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteSecretStore(context.Background(), args[0]); err != nil {
				return fmt.Errorf("deleting secret store: %w", err)
			}
			return nil
		},
	}
}

func newSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"secrets"},
		Short:   "Manage secrets",
	}
	cmd.AddCommand(newSecretListCmd())
	cmd.AddCommand(newSecretGetCmd())
	cmd.AddCommand(newSecretPutCmd())
	cmd.AddCommand(newSecretDeleteCmd())
	return cmd
}

func newSecretListCmd() *cobra.Command {
	var store string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the secrets in a store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			secrets, err := client.ListSecrets(context.Background(), store)
			if err != nil {
				return fmt.Errorf("listing secrets: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretsOutput(secrets))
		},
	}
	cmd.Flags().StringVar(&store, "store", "", "Store name (required)")
	cmd.MarkFlagRequired("store")
	return cmd
}

func newSecretGetCmd() *cobra.Command {
	var reveal bool
	cmd := &cobra.Command{
		Use:   "get <store>/<name>",
		Short: "Get a secret's metadata, or its value with --reveal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, name, err := splitSecretRef(args[0])
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			if !reveal {
				secret, err := client.GetSecret(context.Background(), store, name)
				if err != nil {
					return fmt.Errorf("getting secret: %w", err)
				}
				return output.Print(cmd.OutOrStdout(), flagOutput, secretOutput(secret))
			}
			secret, err := client.RevealSecret(context.Background(), store, name)
			if err != nil {
				return fmt.Errorf("revealing secret: %w", err)
			}
			if flagOutput == "json" {
				return output.Print(cmd.OutOrStdout(), flagOutput, revealedSecret{
					StoreName: secret.StoreName,
					Name:      secret.Name,
					Version:   secret.Version,
					Value:     secret.Value.Reveal(),
				})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), secret.Value.Reveal())
			return err
		},
	}
	cmd.Flags().BoolVar(&reveal, "reveal", false, "Print the secret's value; the console records the read in the account's audit trail")
	return cmd
}

// sdk.SecretValue marshals its value as "[redacted]", so the one place that prints a value builds
// its own object.
type revealedSecret struct {
	StoreName string `json:"storeName"`
	Name      string `json:"name"`
	Version   int    `json:"version"`
	Value     string `json:"value"`
}

func newSecretPutCmd() *cobra.Command {
	var fromFile string
	var keepNewline bool
	cmd := &cobra.Command{
		Use:   "put <store>/<name>",
		Short: "Create a secret or replace its value, read from stdin, a prompt or --from-file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, name, err := splitSecretRef(args[0])
			if err != nil {
				return err
			}
			raw, err := readSecretValue(cmd, fromFile)
			if err != nil {
				return err
			}
			if !keepNewline {
				raw = dropOneNewline(raw)
			}
			value := sdk.NewValue(raw)

			client, err := newClient()
			if err != nil {
				return err
			}
			secret, err := putSecret(context.Background(), client, store, name, value)
			if err != nil {
				return err
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, secretOutput(secret))
		},
	}
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read the value from this file")
	cmd.Flags().BoolVar(&keepNewline, "keep-newline", false, "Keep one trailing newline instead of dropping it")
	return cmd
}

func putSecret(ctx context.Context, client *sdk.Client, store, name string, value sdk.Value) (*sdk.SecretMetadata, error) {
	_, err := client.GetSecret(ctx, store, name)
	if err == nil {
		return updateSecret(ctx, client, store, name, value)
	}
	if !errors.Is(err, sdk.ErrNotFound) {
		return nil, fmt.Errorf("getting secret: %w", err)
	}
	secret, err := client.CreateSecret(ctx, store, name, value)
	if errors.Is(err, sdk.ErrConflict) {
		// Another caller created it between the read and the create.
		return updateSecret(ctx, client, store, name, value)
	}
	if err != nil {
		return nil, fmt.Errorf("creating secret: %w", err)
	}
	return secret, nil
}

func updateSecret(ctx context.Context, client *sdk.Client, store, name string, value sdk.Value) (*sdk.SecretMetadata, error) {
	secret, err := client.UpdateSecretValue(ctx, store, name, value)
	if err != nil {
		return nil, fmt.Errorf("updating secret: %w", err)
	}
	return secret, nil
}

func readSecretValue(cmd *cobra.Command, fromFile string) (string, error) {
	if fromFile != "" {
		b, err := os.ReadFile(fromFile)
		if err != nil {
			return "", fmt.Errorf("reading value: %w", err)
		}
		return string(b), nil
	}
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(cmd.ErrOrStderr(), "Value: ")
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("reading value: %w", err)
		}
		return string(b), nil
	}
	b, err := io.ReadAll(in)
	if err != nil {
		return "", fmt.Errorf("reading value: %w", err)
	}
	return string(b), nil
}

func dropOneNewline(s string) string {
	if strings.HasSuffix(s, "\r\n") {
		return strings.TrimSuffix(s, "\r\n")
	}
	return strings.TrimSuffix(s, "\n")
}

func newSecretDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <store>/<name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, name, err := splitSecretRef(args[0])
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteSecret(context.Background(), store, name); err != nil {
				return fmt.Errorf("deleting secret: %w", err)
			}
			return nil
		},
	}
}

func splitSecretRef(ref string) (string, string, error) {
	store, name, ok := strings.Cut(ref, "/")
	if !ok || store == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("secret %q must be written <store>/<name>", ref)
	}
	return store, name, nil
}

func secretStoresOutput(stores []sdk.SecretStore) any {
	if flagOutput == "json" {
		return stores
	}
	return secretStoreTable(stores)
}

func secretStoreOutput(store *sdk.SecretStore) any {
	if flagOutput == "json" {
		return store
	}
	return secretStoreTable([]sdk.SecretStore{*store})
}

func secretStoreTable(stores []sdk.SecretStore) *output.Table {
	rows := make([][]string, 0, len(stores))
	for _, s := range stores {
		rows = append(rows, []string{s.Name, s.Description, strconv.Itoa(s.SecretCount), authorOf(s.LastModifiedBy), s.UpdatedAt})
	}
	return &output.Table{Headers: []string{"NAME", "DESCRIPTION", "SECRETS", "LAST MODIFIED BY", "UPDATED"}, Rows: rows}
}

func secretsOutput(secrets []sdk.SecretMetadata) any {
	if flagOutput == "json" {
		return secrets
	}
	return secretTable(secrets)
}

func secretOutput(secret *sdk.SecretMetadata) any {
	if flagOutput == "json" {
		return secret
	}
	return secretTable([]sdk.SecretMetadata{*secret})
}

func secretTable(secrets []sdk.SecretMetadata) *output.Table {
	rows := make([][]string, 0, len(secrets))
	for _, s := range secrets {
		rows = append(rows, []string{s.StoreName, s.Name, strconv.Itoa(s.Version), s.SyncState, authorOf(s.LastModifiedBy), s.UpdatedAt})
	}
	return &output.Table{Headers: []string{"STORE", "NAME", "VERSION", "SYNC", "LAST MODIFIED BY", "UPDATED"}, Rows: rows}
}

func authorOf(id *sdk.Identity) string {
	if id == nil {
		return ""
	}
	return id.Email
}
