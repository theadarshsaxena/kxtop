/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pvc

import (
	// "context"
	"fmt"
	// "path/filepath"

	"github.com/spf13/cobra"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/util/homedir"
	// metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
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

// REference: https://github.com/yashbhutwala/kubectl-df-pv/blob/master/pkg/df-pv/root.go

// func fetchPvc(cmd *cobra.Command, args []string) {
// kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// if err != nil {
// 	fmt.Printf("Failed to create config: %v\n", err)
// 	return
// }

// clientset, err := kubernetes.NewForConfig(config)
// if err != nil {
// 	fmt.Printf("Failed to create clientset: %v\n", err)
// 	return
// }

// metricsClientSet, err := metricsv.NewForConfig(config)
// if err != nil {
// 	fmt.Printf("Failed to create metrics clientset: %v\n", err)
// 	return
// }

// pvc, err := clientset.CoreV1().PersistentVolumeClaims("default").List(context.TODO(), metav1.ListOptions{})
// if err != nil {
// 	fmt.Printf("Failed to list PVCs: %v\n", err)
// 	return
// }

// pvc.Items[0].Status.

// if err != nil {
// 	fmt.Printf("Failed to list pods: %v\n", err)
// 	return
// }

// podsMetrics, err := metricsClientSet.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
// if err != nil {
// 	fmt.Printf("Failed to list pod metrics: %v\n", err)
// 	return
// }
// }

func init() {}
