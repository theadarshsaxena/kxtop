/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pods

import (
	"context"
	"fmt"
	"kxtop/app/cmd/analytics"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var namespace string

type PodUsage struct {
	PodName           string
	Namespace         string
	CPUUsed           int64
	CPURequest        int64
	CPUUsedPercent    float64
	MemoryUsed        int64
	MemoryRequest     int64
	MemoryUsedPercent float64
	NodeName          string
}

type Containers struct {
	Name              string
	CPURequest        int64
	CPUUsed           int64
	CPUUsedPercent    float64
	MemoryRequest     int64
	MemoryUsed        int64
	MemoryUsedPercent float64
}

// podsCmd represents the pods command
var PodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Get the resource usage of pods in the cluster.",
	Long:  `Get the resource usage of pods in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pods called")
	},
}

func fetchPods(cmd *cobra.Command, args []string) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Failed to create config: %v\n", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %v\n", err)
		return
	}

	metricsClientSet, err := metricsv.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create metrics clientset: %v\n", err)
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list pods: %v\n", err)
		return
	}

	podsMetrics, err := metricsClientSet.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list pod metrics: %v\n", err)
		return
	}

	podUsages := make([]PodUsage, 0) //[]PodUsage{}
	for _, pod := range pods.Items {
		podUsage := PodUsage{
			PodName:   pod.Name,
			Namespace: pod.Namespace,
		}

		for _, container := range pod.Spec.Containers {
			podUsage.CPURequest += container.Resources.Requests.Cpu().MilliValue()
			podUsage.MemoryRequest += analytics.ConvertToMi(container.Resources.Requests.Memory().String())
		}

		for _, podMetric := range podsMetrics.Items {
			if podMetric.Name == pod.Name {
				for _, containerMetric := range podMetric.Containers {
					podUsage.CPUUsed += containerMetric.Usage.Cpu().MilliValue()
					podUsage.MemoryUsed += analytics.ConvertToMi(containerMetric.Usage.Memory().String())
				}
			}
		}

		podUsage.CPUUsedPercent = float64(podUsage.CPUUsed) / float64(podUsage.CPURequest) * 100
		podUsage.MemoryUsedPercent = float64(podUsage.MemoryUsed) / float64(podUsage.MemoryRequest) * 100
		podUsage.NodeName = pod.Spec.NodeName

		podUsages = append(podUsages, podUsage)
	}

	printPods(podUsages)
}

func printPods(podUsages []PodUsage) {
	podNameLength := 3
	namespaceLength := 9
	nodeNameLength := 4
	cpuRequestLength := 0
	cpuUsedLength := 0
	memoryRequestLength := 0
	memoryUsedLength := 0

	for _, podUsage := range podUsages {
		if len(podUsage.PodName) > podNameLength {
			podNameLength = len(podUsage.PodName)
		}
		if len(podUsage.Namespace) > namespaceLength {
			namespaceLength = len(podUsage.Namespace)
		}
		if len(podUsage.NodeName) > nodeNameLength {
			nodeNameLength = len(podUsage.NodeName)
		}
		if len(strconv.FormatInt(podUsage.CPURequest, 10)) > cpuRequestLength {
			cpuRequestLength = len(strconv.FormatInt(podUsage.CPURequest, 10))
		}
		if len(strconv.FormatInt(podUsage.CPUUsed, 10)) > cpuUsedLength {
			cpuUsedLength = len(strconv.FormatInt(podUsage.CPUUsed, 10))
		}
		if len(strconv.FormatInt(podUsage.MemoryRequest, 10)) > memoryRequestLength {
			memoryRequestLength = len(strconv.FormatInt(podUsage.MemoryRequest, 10))
		}
		if len(strconv.FormatInt(podUsage.MemoryUsed, 10)) > memoryUsedLength {
			memoryUsedLength = len(strconv.FormatInt(podUsage.MemoryUsed, 10))
		}
	}

	printHeader(podNameLength, namespaceLength)

	for _, podUsage := range podUsages {
		// CPU color
		cpuColor := getCPUPrintColor(podUsage.CPUUsedPercent)
		// Memory color
		memoryColor := getMemoryPrintColor(podUsage.MemoryUsedPercent)
		fmt.Printf("%-*s  %-*s  %s%*dm/%-*s  %4.1f%%  %s%*dMi/%-*s  %5.1f%%\x1b[0m  %-10s\n",
			podNameLength, podUsage.PodName, namespaceLength, podUsage.Namespace, // pod name & ns
			cpuColor, cpuUsedLength, podUsage.CPUUsed, cpuRequestLength+1, fmt.Sprintf("%dm", podUsage.CPURequest), podUsage.CPUUsedPercent, // cpu
			memoryColor, memoryUsedLength, podUsage.MemoryUsed, memoryRequestLength+2, fmt.Sprintf("%dMi", podUsage.MemoryRequest), podUsage.MemoryUsedPercent, // memory
			podUsage.NodeName) // node
	}
}

func printHeader(podNameLength int, namespaceLength int) {
	fmt.Printf("\x1b[1;34m%-*s  %-*s  %-15s  %-20s  %-10s\x1b[0m\n", podNameLength, "Pod", namespaceLength, "Namespace", "CPU(use/rqst)", "Memory(use/rqst)", "Node")
}

// func getMemoryPercentColor(a int64, b int64) (percent string, color string) {
// 	if b == 0 {
// 		return "--", "\x1b[33m"
// 	}
// 	return (float64(a) / float64(b)) * 100
// }

// func getCPUPercentage() {

// }

func init() {
	PodsCmd.Run = fetchPods
	PodsCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace for fetching pods")
}

func getCPUPrintColor(cpuPercent float64) string {
	if cpuPercent < 15 {
		return "\x1b[33m" // yellow
	} else if cpuPercent < 75 {
		return "\x1b[32m" // green
	} else {
		return "\x1b[31m" // red
	}
}
func getMemoryPrintColor(memoryPercent float64) string {
	if memoryPercent < 25 {
		return "\x1b[33m" // yellow  (under utilized)
	} else if memoryPercent < 75 {
		return "\x1b[32m" // Green (perfect)
	} else {
		return "\x1b[31m" // red (high usage)
	}
}

// func convertToMi(memory string) int64 {
// 	if strings.HasSuffix(memory, "Ki") {
// 		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Ki"), 10, 64)
// 		return memoryInt / 1024
// 	}
// 	if strings.HasSuffix(memory, "Mi") {
// 		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Mi"), 10, 64)
// 		return memoryInt
// 	}
// 	if strings.HasSuffix(memory, "Gi") {
// 		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Gi"), 10, 64)
// 		return memoryInt * 1024
// 	}
// 	return 0
// }
