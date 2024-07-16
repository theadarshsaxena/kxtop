package analytics

import (
	"fmt"
	"github.com/spf13/cobra"
)

// analyticsCmd represents the analytics command
var AnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Get the analytics of the resource usage in the cluster.",
	Long:  `Get the analytics of the resource usage in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Coming soon! will tell the analytics of the resource usage in the cluster.")
	},
}

func init() {}
