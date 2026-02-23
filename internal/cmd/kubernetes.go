package cmd

import (
	"context"
	"fmt"

	"github.com/mvdatacenter/mvdata-cli/internal/output"
	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

func newKubernetesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "Manage Kubernetes clusters",
	}
	cmd.AddCommand(newKubernetesCreateCmd())
	cmd.AddCommand(newKubernetesGetCmd())
	cmd.AddCommand(newKubernetesUpdateCmd())
	cmd.AddCommand(newKubernetesDeleteCmd())
	return cmd
}

func newKubernetesCreateCmd() *cobra.Command {
	var name, k8sVersion, nodeType string
	var nodeCount int
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			cluster, err := client.CreateKubernetesCluster(context.Background(), &sdk.KubernetesCluster{
				Name:             name,
				Version:          k8sVersion,
				NodeInstanceType: nodeType,
				NodeCount:        nodeCount,
			})
			if err != nil {
				return fmt.Errorf("creating kubernetes cluster: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, kubernetesOutput(cluster))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.Flags().StringVar(&k8sVersion, "version", "", "Kubernetes version (required)")
	cmd.Flags().StringVar(&nodeType, "node-type", "", "Node instance type (required)")
	cmd.Flags().IntVar(&nodeCount, "node-count", 0, "Number of nodes (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("node-type")
	cmd.MarkFlagRequired("node-count")
	return cmd
}

func newKubernetesGetCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			cluster, err := client.GetKubernetesCluster(context.Background(), name)
			if err != nil {
				return fmt.Errorf("getting kubernetes cluster: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, kubernetesOutput(cluster))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newKubernetesUpdateCmd() *cobra.Command {
	var name string
	var nodeCount int
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			cluster, err := client.UpdateKubernetesCluster(context.Background(), name, &sdk.KubernetesClusterUpdate{
				NodeCount: nodeCount,
			})
			if err != nil {
				return fmt.Errorf("updating kubernetes cluster: %w", err)
			}
			return output.Print(cmd.OutOrStdout(), flagOutput, kubernetesOutput(cluster))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.Flags().IntVar(&nodeCount, "node-count", 0, "Number of nodes (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("node-count")
	return cmd
}

func newKubernetesDeleteCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}
			if err := client.DeleteKubernetesCluster(context.Background(), name); err != nil {
				return fmt.Errorf("deleting kubernetes cluster: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Kubernetes cluster %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func kubernetesOutput(c *sdk.KubernetesCluster) any {
	if flagOutput == "json" {
		return c
	}
	return &output.Table{
		Headers: []string{"NAME", "VERSION", "NODE TYPE", "NODES", "ENDPOINT", "STATUS", "CREATED"},
		Rows: [][]string{{
			c.Name,
			c.Version,
			c.NodeInstanceType,
			fmt.Sprintf("%d", c.NodeCount),
			c.Endpoint,
			c.Status,
			c.CreatedAt,
		}},
	}
}
