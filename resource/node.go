package resource

import (
	"context"
	"fmt"

	"github.com/feo0o/kubeview/kube"
	"github.com/jedib0t/go-pretty/v6/table"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeresource "k8s.io/cli-runtime/pkg/resource"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Node struct {
	Name              string
	Capacity          Resource
	Allocatable       Resource
	AllocatedRequests Resource
	AllocatedLimits   Resource
}

type Resource struct {
	CPU              resource.Quantity
	Memory           resource.Quantity
	StorageEphemeral resource.Quantity
	//HugePages1Gi     *resource.Quantity
	//HugePages2Mi     *resource.Quantity
	Pods resource.Quantity
}

func nodesResources() (nodes []Node, err error) {
	nodeList, err := kube.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		var nodeResource Node
		var capacity, allocatable, allocatedRequests, allocatedLimits Resource
		nodeResource.Name = node.Name

		capacity.CPU = *node.Status.Capacity.Cpu()
		capacity.Memory = *node.Status.Capacity.Memory()
		capacity.StorageEphemeral = *node.Status.Capacity.StorageEphemeral()
		capacity.Pods = *node.Status.Capacity.Pods()
		nodeResource.Capacity = capacity

		allocatable.CPU = *node.Status.Allocatable.Cpu()
		allocatable.Memory = *node.Status.Allocatable.Memory()
		allocatable.StorageEphemeral = *node.Status.Allocatable.StorageEphemeral()
		allocatable.Pods = *node.Status.Allocatable.Pods()
		nodeResource.Allocatable = allocatable

		fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name + ",status.phase!=" + string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
		if err != nil {
			return nil, err
		}
		opts := metav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		}
		nodeNonTerminatedPodsList, err := getPodsInChunks(kube.ClientSet.CoreV1().Pods(""), opts)
		if err != nil {
			return nil, err
		}
		for _, pod := range nodeNonTerminatedPodsList.Items {
			for _, container := range pod.Spec.Containers {
				allocatedRequests.CPU.Add(*container.Resources.Requests.Cpu())
				allocatedRequests.Memory.Add(*container.Resources.Requests.Memory())
				allocatedRequests.StorageEphemeral.Add(*container.Resources.Requests.StorageEphemeral())

				allocatedLimits.CPU.Add(*container.Resources.Limits.Cpu())
				allocatedLimits.Memory.Add(*container.Resources.Requests.Memory())
				allocatedLimits.StorageEphemeral.Add(*container.Resources.Limits.StorageEphemeral())
			}
		}
		allocatedRequests.Pods = *resource.NewQuantity(int64(len(nodeNonTerminatedPodsList.Items)), resource.DecimalSI)
		allocatedLimits.Pods = *resource.NewQuantity(int64(len(nodeNonTerminatedPodsList.Items)), resource.DecimalSI)

		nodeResource.AllocatedRequests = allocatedRequests
		nodeResource.AllocatedLimits = allocatedLimits

		nodes = append(nodes, nodeResource)
	}
	return nodes, nil
}

func getPodsInChunks(c corev1client.PodInterface, initialOpts metav1.ListOptions) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	err := runtimeresource.FollowContinue(&initialOpts,
		func(options metav1.ListOptions) (runtime.Object, error) {
			newList, err := c.List(context.TODO(), options)
			if err != nil {
				return nil, runtimeresource.EnhanceListError(err, options, corev1.ResourcePods.String())
			}
			podList.Items = append(podList.Items, newList.Items...)
			return newList, nil
		})
	return podList, err
}

func PrintNodesResourcesToStdout() {
	nodes, err := nodesResources()
	if err != nil {
		fmt.Println(err.Error())
	}

	twCapacity := table.NewWriter()
	twCapacity.AppendHeader(table.Row{"NODE", "CPU", "Memory", "Ephemeral Storage", "Pods"})

	twAllocatable := table.NewWriter()
	twAllocatable.AppendHeader(table.Row{"NODE", "CPU", "Memory", "Ephemeral Storage", "Pods"})

	twAllocatedRequests := table.NewWriter()
	twAllocatedRequests.AppendHeader(
		table.Row{
			"NODE",
			"CPU",
			"CPU Percent",
			"Memory",
			"Mem Percent",
			"Ephemeral Storage",
			"Storage Percent",
			"Pods",
			"Pods Percent",
		},
	)

	twAllocatedLimits := table.NewWriter()
	twAllocatedLimits.AppendHeader(
		table.Row{
			"NODE",
			"CPU",
			"CPU Percent",
			"Memory",
			"Mem Percent",
			"Ephemeral Storage",
			"Storage Percent",
			"Pods",
			"Pods Percent",
		},
	)

	for _, node := range nodes {
		twCapacity.AppendRow(
			table.Row{
				node.Name,
				node.Capacity.CPU.String(),
				node.Capacity.Memory.String(),
				node.Capacity.StorageEphemeral.String(),
				node.Capacity.Pods.String(),
			},
		)
		twAllocatable.AppendRow(
			table.Row{
				node.Name,
				node.Allocatable.CPU.String(),
				node.Allocatable.Memory.String(),
				node.Allocatable.StorageEphemeral.String(),
				node.Allocatable.Pods.String(),
			},
		)
		twAllocatedRequests.AppendRow(
			table.Row{
				node.Name,
				node.AllocatedRequests.CPU.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedRequests.CPU.AsApproximateFloat64()/node.Allocatable.CPU.AsApproximateFloat64()),
				node.AllocatedRequests.Memory.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedRequests.Memory.AsApproximateFloat64()/node.Allocatable.Memory.AsApproximateFloat64()),
				node.AllocatedRequests.StorageEphemeral.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedRequests.StorageEphemeral.AsApproximateFloat64()/node.Allocatable.StorageEphemeral.AsApproximateFloat64()),
				node.AllocatedRequests.Pods.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedRequests.Pods.AsApproximateFloat64()/node.Allocatable.Pods.AsApproximateFloat64()),
			},
		)
		twAllocatedLimits.AppendRow(
			table.Row{
				node.Name,
				node.AllocatedLimits.CPU.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedLimits.CPU.AsApproximateFloat64()/node.Allocatable.CPU.AsApproximateFloat64()),
				node.AllocatedLimits.Memory.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedLimits.Memory.AsApproximateFloat64()/node.Allocatable.Memory.AsApproximateFloat64()),
				node.AllocatedLimits.StorageEphemeral.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedLimits.StorageEphemeral.AsApproximateFloat64()/node.Allocatable.StorageEphemeral.AsApproximateFloat64()),
				node.AllocatedLimits.Pods.String(),
				fmt.Sprintf("%.2f%%", node.AllocatedLimits.Pods.AsApproximateFloat64()/node.Allocatable.Pods.AsApproximateFloat64()),
			},
		)
	}
	fmt.Printf("\nNodes Capacity:\n%s\n", twCapacity.Render())
	fmt.Printf("\nNodes Allocatable:\n%s\n", twAllocatable.Render())
	fmt.Printf("\nNodes Allocated Requests:\n%s\n", twAllocatedRequests.Render())
	fmt.Printf("\nNodes Allocated Limits:\n%s\n", twAllocatedLimits.Render())
}
