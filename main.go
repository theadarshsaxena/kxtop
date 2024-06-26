/*
Copyright © 2024 Adarsh adarshsaxena358@gmail.com
*/
package main

import (
	"kxtop/app/cmd/analytics"
	"kxtop/app/cmd/nodes"
	"kxtop/app/cmd/pods"
	"kxtop/app/cmd/pvc"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kxtop",
	Short: "An advanced kubectl top alternative for Kubernetes cluster.",
	Long: `Kxtop is an advanced kubectl top alternative for Kubernetes cluster.
It provides the resource usage of pods, nodes, and PVCs in the cluster.
It also provides the analytics of the resource usage in the cluster.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	rootCmd.AddCommand(pods.PodsCmd)
	rootCmd.AddCommand(nodes.NodesCmd)
	rootCmd.AddCommand(pvc.PvcCmd)
	rootCmd.AddCommand(analytics.AnalyticsCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
