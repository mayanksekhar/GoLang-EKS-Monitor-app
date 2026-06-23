package metrics

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeMetrics struct {
	Name        string
	CPUCapacity int64
	CPUUsage    int64
	CPUPercent  float64
	MemCapacity int64
	MemUsage    int64
	MemPercent  float64
	Role        string
}

type ClusterMetrics struct {
	ClusterName  string
	Nodes        []NodeMetrics
	TotalPods    int
	RunningPods  int
	PendingPods  int
	Namespaces   int
	TotalCPU     float64
	TotalMem     float64
}

type Collector struct {
	client        *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
}

func NewCollector() (*Collector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig for local development
		home, _ := os.UserHomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &Collector{client: client, metricsClient: metricsClient}, nil
}

func (c *Collector) Collect(ctx context.Context) (*ClusterMetrics, error) {
	cm := &ClusterMetrics{
		ClusterName: os.Getenv("CLUSTER_NAME"),
	}

	// Get nodes
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get node metrics
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Build metrics map
	metricsMap := make(map[string]interface{})
	for _, nm := range nodeMetrics.Items {
		metricsMap[nm.Name] = nm
	}

	for _, node := range nodes.Items {
		nm := NodeMetrics{Name: node.Name}

		// Capacity
		cpuCap := node.Status.Capacity[corev1.ResourceCPU]
		memCap := node.Status.Capacity[corev1.ResourceMemory]
		nm.CPUCapacity = cpuCap.MilliValue()
		nm.MemCapacity = memCap.Value() / 1024 / 1024 // MB

		// Usage from metrics-server
		if raw, ok := metricsMap[node.Name]; ok {
			if typed, ok := raw.(interface {
				GetName() string
			}); ok {
				_ = typed
			}
			// type assertion via the nodeMetrics items
			for _, item := range nodeMetrics.Items {
				if item.Name == node.Name {
					cpuUsage := item.Usage[corev1.ResourceCPU]
					memUsage := item.Usage[corev1.ResourceMemory]
					nm.CPUUsage = cpuUsage.MilliValue()
					nm.MemUsage = memUsage.Value() / 1024 / 1024
					if nm.CPUCapacity > 0 {
						nm.CPUPercent = float64(nm.CPUUsage) / float64(nm.CPUCapacity) * 100
					}
					if nm.MemCapacity > 0 {
						nm.MemPercent = float64(nm.MemUsage) / float64(nm.MemCapacity) * 100
					}
				}
			}
		}

		// Node role
		nm.Role = "worker"
		for label := range node.Labels {
			if label == "node-role.kubernetes.io/control-plane" {
				nm.Role = "control-plane"
			}
		}

		cm.Nodes = append(cm.Nodes, nm)
		cm.TotalCPU += nm.CPUPercent
		cm.TotalMem += nm.MemPercent
	}

	if len(cm.Nodes) > 0 {
		cm.TotalCPU /= float64(len(cm.Nodes))
		cm.TotalMem /= float64(len(cm.Nodes))
	}

	// Get pods
	pods, err := c.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	cm.TotalPods = len(pods.Items)
	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			cm.RunningPods++
		case corev1.PodPending:
			cm.PendingPods++
		}
	}

	// Get namespaces
	ns, err := c.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	cm.Namespaces = len(ns.Items)

	return cm, nil
}
