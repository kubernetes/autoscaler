package utho

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// newInClusterClient returns a Kubernetes clientset using in-cluster config.
func newInClusterClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return clientset, nil
}

// getNodeLabel retrieves the value of the given labelKey from the first node
// in the cluster that has it.
func getNodeLabel(labelKey string) (string, error) {
	clientset, err := newInClusterClient()
	if err != nil {
		return "", err
	}

	nodeList, err := clientset.CoreV1().
		Nodes().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}
	if len(nodeList.Items) == 0 {
		return "", errors.New("no nodes found in cluster")
	}

	for _, node := range nodeList.Items {
		if val, ok := node.Labels[labelKey]; ok {
			return val, nil
		}
	}

	return "", fmt.Errorf("label %q not found on any node", labelKey)
}

// normalizeID strips URL schemes (e.g. "utho://") or other prefixes from an ID
func normalizeID(id string) string {
	return strings.TrimPrefix(id, "utho://")
}

// readyConditions marks the synthetic node as Ready so the scheduler
// can count its resources during the scale-up simulation.
func readyConditions() []apiv1.NodeCondition {
	return []apiv1.NodeCondition{
		{
			Type:   apiv1.NodeReady,
			Status: apiv1.ConditionTrue,
		},
	}
}

// buildKubeProxy returns a tiny Pod object representing kube-proxy.
// CA adds it to every template so that DaemonSet resources are included.
func buildKubeProxy(nodeGroupID string) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-" + nodeGroupID,
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-proxy"},
		},
	}
}

// join merges two string maps (src wins on duplicate keys).
func join(dst, src map[string]string) map[string]string {
	if dst == nil {
		dst = map[string]string{}
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
