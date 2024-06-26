/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package nodes

import (
	"context"
	"fmt"
	"kxtop/app/cmd/analytics"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeUsage struct {
	NodeGroup                string
	NodeName                 string
	CPUUsed                  int64
	CPUAllocatable           int64
	CPUAllocatablePercent    float64
	CPUPercent               float64
	CPUAllocated             int64
	MemoryUsed               int64
	MemoryAllocatable        int64
	MemoryPercent            float64
	MemoryAllocated          int64
	MemoryAllocatablePercent float64
	NumPods                  string
	Disk                     string
	InstanceType             string
	CostPerHour              float64
}

// nodesCmd represents the nodes command
var NodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Get the resource usage of nodes in the cluster.",
	Long:  `Get the resource usage of nodes in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nodes called")
	},
}

func init() {
	NodesCmd.Run = fetchNodes
}

func fetchNodes(cmd *cobra.Command, args []string) {
	// Create a Kubernetes clientset
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

	// Fetch all nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to fetch nodes: %v\n", err)
		return
	}

	// Fetch node metrics
	nodeMetrics, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to fetch node metrics: %v\n", err)
		return
	}

	// Fetch all pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to fetch pods: %v\n", err)
		return
	}

	// Combine pods and nodes
	podCount := make(map[string]int)
	nodeCPUAllocated := make(map[string]int64)
	nodeMemoryAllocated := make(map[string]int64)
	for _, pod := range pods.Items {
		nodeName := pod.Spec.NodeName
		podCount[nodeName]++
		var podCPURequest int64 = 0
		var podMemoryRequest int64 = 0
		for _, container := range pod.Spec.Containers {
			if pod.Status.Phase == "Running" {
				podCPURequest += container.Resources.Requests.Cpu().MilliValue()
				podMemoryRequest += analytics.ConvertToMi(container.Resources.Requests.Memory().String())
			}
		}
		nodeCPUAllocated[nodeName] += podCPURequest
		nodeMemoryAllocated[nodeName] += podMemoryRequest
	}

	instanceCounts := make(map[string]int)
	for _, node := range nodes.Items {
		instanceCounts[node.Labels["beta.kubernetes.io/instance-type"]]++
	}

	pricePerHour := 0.0
	for instanceType, count := range instanceCounts {
		pricePerHour += float64(count) * InitialOnDemandPricesAWS[instanceType]
	}

	// Combine nodes and metrics
	nodeUsages := make([]NodeUsage, 0)
	for _, node := range nodes.Items {
		nodeUsage := NodeUsage{
			NodeGroup:         node.Labels["eks.amazonaws.com/nodegroup"],
			NodeName:          node.Name,
			CPUAllocatable:    node.Status.Allocatable.Cpu().MilliValue(),
			MemoryAllocatable: analytics.ConvertToMi(node.Status.Allocatable.Memory().String()),
			Disk:              node.Status.Capacity.StorageEphemeral().String(),
			// NumPods:           podCount[node.Name],
			NumPods:                  fmt.Sprintf("%d/%d", podCount[node.Name], int(node.Status.Allocatable.Pods().Value())),
			InstanceType:             node.Labels["beta.kubernetes.io/instance-type"],
			CPUAllocated:             nodeCPUAllocated[node.Name],
			MemoryAllocated:          nodeMemoryAllocated[node.Name],
			CPUAllocatablePercent:    float64(nodeCPUAllocated[node.Name]) / float64(node.Status.Allocatable.Cpu().MilliValue()) * 100,
			MemoryAllocatablePercent: float64(nodeMemoryAllocated[node.Name]) / float64(analytics.ConvertToMi(node.Status.Allocatable.Memory().String())) * 100,
			CostPerHour:              pricePerHour,
		}

		for _, nodeMetric := range nodeMetrics.Items {
			if nodeMetric.Name == node.Name {
				nodeUsage.CPUUsed = nodeMetric.Usage.Cpu().MilliValue()
				nodeUsage.MemoryUsed = analytics.ConvertToMi(nodeMetric.Usage.Memory().String())
				nodeUsage.CPUPercent = float64(nodeMetric.Usage.Cpu().MilliValue()) / float64(node.Status.Allocatable.Cpu().MilliValue()) * 100
				nodeUsage.MemoryPercent = float64(nodeMetric.Usage.Memory().Value()) / float64(node.Status.Allocatable.Memory().Value()) * 100
			}
		}

		nodeUsages = append(nodeUsages, nodeUsage)
	}

	totalCPUAllocatable := 0
	totalMemoryAllocatable := 0
	totalCPUUsed := 0
	totalMemoryUsed := 0
	for _, nodeUsage := range nodeUsages {
		totalCPUAllocatable += int(nodeUsage.CPUAllocatable)
		totalCPUUsed += int(nodeUsage.CPUUsed)
	}
	efficiencyCPU := float64(totalCPUUsed) / float64(totalCPUAllocatable) * 100

	for _, nodeUsage := range nodeUsages {
		totalMemoryAllocatable += int(nodeUsage.MemoryAllocatable)
		totalMemoryUsed += int(nodeUsage.MemoryUsed)
	}
	efficiencyMemory := float64(totalMemoryUsed) / float64(totalMemoryAllocatable) * 100
	printNodeUsages(nodeUsages)

	fmt.Printf("\n\x1b[35mTotal cost per hour:\x1b[33m $%.2f \x1b[35m& Total cost per month:\x1b[33m $%.2f\n", pricePerHour, pricePerHour*24*30)
	fmt.Printf("\x1b[35mTotal CPU efficiency: \x1b[33m%.2f%% \x1b[35m& Total Memory efficiency:\x1b[33m %.2f%%\n", efficiencyCPU, efficiencyMemory)

	fmt.Printf(
		"\n\x1b[0mTIP: \nAllocated CPU/Memory: \x1b[33mYellow (Under Allocated) \x1b[0m < 75%% <= \x1b[32mGreen (Well Allocated) \x1b[0m \nCPU: \x1b[33mYellow (Under utilized) \x1b[0m < 15%% <= \x1b[32mGreen (Perfect) \x1b[0m < 75%% <= \x1b[31m Red (High Usage) \x1b[0m\nMemory: \x1b[33m Yellow (Under utilized) \x1b[0m < 25%% <= \x1b[32m Green (Perfect) \x1b[0m < 75%% <= \x1b[31m Red (High Usage) \x1b[0m\n",
	)
}

// Function to print the nodeUsages
func printNodeUsages(nodeUsages []NodeUsage) {
	maxNodeGroupLength := 0
	maxNodeLength := 0
	cpuAllocatedLength := 0
	memoryAllocatedLength := 0
	cpuUsedLength := 0
	memoryUsedLength := 0
	memoryAllocatableLength := 0
	cpuAllocatableLength := 0
	for _, nodeUsage := range nodeUsages {
		if len(nodeUsage.NodeGroup) > maxNodeGroupLength {
			maxNodeGroupLength = len(nodeUsage.NodeGroup)
		}
		if len(nodeUsage.NodeName) > maxNodeLength {
			maxNodeLength = len(nodeUsage.NodeName)
		}
		if len(fmt.Sprintf("%d", nodeUsage.CPUAllocated)) > cpuAllocatedLength {
			cpuAllocatedLength = len(fmt.Sprintf("%d", nodeUsage.CPUAllocated))
		}
		if len(fmt.Sprintf("%d", nodeUsage.MemoryAllocated)) > memoryAllocatedLength {
			memoryAllocatedLength = len(fmt.Sprintf("%d", nodeUsage.MemoryAllocated))
		}
		if len(fmt.Sprintf("%d", nodeUsage.CPUUsed)) > cpuUsedLength {
			cpuUsedLength = len(fmt.Sprintf("%d", nodeUsage.CPUUsed))
		}
		if len(fmt.Sprintf("%.2f", nodeUsage.MemoryPercent)) > memoryUsedLength {
			memoryUsedLength = len(fmt.Sprintf("%.2f", nodeUsage.MemoryPercent))
		}
		if len(fmt.Sprintf("%d", nodeUsage.MemoryAllocatable)) > memoryAllocatableLength {
			memoryAllocatableLength = len(fmt.Sprintf("%d", nodeUsage.MemoryAllocatable))
		}
		if len(fmt.Sprintf("%d", nodeUsage.CPUAllocatable)) > cpuAllocatableLength {
			cpuAllocatableLength = len(fmt.Sprintf("%d", nodeUsage.CPUAllocatable))
		}
	}
	if nodeUsages == nil {
		fmt.Println("No nodes found")
		return
	}
	printColoredHeader(maxNodeGroupLength, maxNodeLength, cpuAllocatableLength+cpuAllocatedLength+11, cpuUsedLength+cpuAllocatableLength+11, memoryAllocatableLength+memoryAllocatedLength+11, memoryUsedLength+memoryAllocatableLength+12)

	// Sort nodeUsages by NodeGroup in ascending order
	sort.Slice(nodeUsages, func(i, j int) bool {
		return nodeUsages[i].NodeGroup < nodeUsages[j].NodeGroup
	})

	for _, nodeUsage := range nodeUsages {
		cpuPrintColor := getCPUPrintColor(nodeUsage.CPUPercent)
		memoryPrintColor := getMemoryPrintColor(nodeUsage.MemoryPercent)
		allocatableCPUPrintColor := getAllocatablePrintColor(nodeUsage.CPUAllocatablePercent)
		allocatableMemoryPrintColor := getAllocatablePrintColor(nodeUsage.MemoryAllocatablePercent)
		fmt.Printf("%-*s %-*s %s%*dm/%-*s %5s%%  %s%*dm/%-*s %5s%%  %s%*dMi/%-*s %5s%% %s%*dMi/%-*s %5s%%\x1b[0m  %-7s %-15s\n",
			maxNodeGroupLength, nodeUsage.NodeGroup, // nodegroup
			maxNodeLength, nodeUsage.NodeName, // node name
			allocatableCPUPrintColor, cpuAllocatedLength, nodeUsage.CPUAllocated, cpuAllocatableLength+1, fmt.Sprintf("%dm", nodeUsage.CPUAllocatable), fmt.Sprintf("%.2f", nodeUsage.CPUAllocatablePercent), // CPU allocated
			cpuPrintColor, cpuUsedLength, nodeUsage.CPUUsed, cpuAllocatableLength+1, fmt.Sprintf("%dm", nodeUsage.CPUAllocatable), fmt.Sprintf("%.2f", nodeUsage.CPUPercent), // CPU used
			allocatableMemoryPrintColor, memoryAllocatedLength, nodeUsage.MemoryAllocated, memoryAllocatableLength+2, fmt.Sprintf("%dMi", nodeUsage.MemoryAllocatable), fmt.Sprintf("%.2f", nodeUsage.MemoryAllocatablePercent),
			memoryPrintColor, memoryUsedLength, nodeUsage.MemoryUsed, memoryAllocatableLength+2, fmt.Sprintf("%dMi", nodeUsage.MemoryAllocatable), fmt.Sprintf("%.2f", nodeUsage.MemoryPercent),
			nodeUsage.NumPods,
			nodeUsage.InstanceType)
	}

}

func printColoredHeader(maxNodeGroupLength int, maxNodeLength int, cpuAllocatedLength int, cpuUsedLength int, memoryAllocatedLength int, memoryUsedLength int) {
	fmt.Printf("\x1b[1;34m%-*s %-*s %-*s %-*s %-*s   %-*s %-8s %-15s\x1b[0m\n", maxNodeGroupLength, "NodeGroup", maxNodeLength, "Nodes", cpuAllocatedLength, "CPU-Allocated", cpuUsedLength, "CPU-Used", memoryAllocatedLength, "Memory-Allocated", memoryUsedLength, "Memory-Used", "#Pods", "Instance-type")
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

func getAllocatablePrintColor(percent float64) string {
	if percent < 75 {
		return "\x1b[33m" // yellow  (under Allocated)
	} else {
		return "\x1b[32m" // greem (Well Allocated)
	}
}

func init() {}
