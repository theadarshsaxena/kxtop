package pvc

import (
	"fmt"
	"github.com/spf13/cobra"
)

// pvcCmd represents the pvc command
var PvcCmd = &cobra.Command{
	Use:   "pvc",
	Short: "Get the resource usage of PVCs in the cluster.",
	Long:  `Get the resource usage of PVCs in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pvc called")
	},
}

func init() {}
