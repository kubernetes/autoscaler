package target

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/control/k8sclient"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesTarget struct {
	kubeClient kubernetes.Interface
	patcher    k8sclient.ResourcePatcher
}

var _ Interface = &KubernetesTarget{}

func NewKubernetesTarget(kubeClient kubernetes.Interface) (Interface, error) {
	patcher, err := k8sclient.NewKubernetesPatcher(kubeClient)
	if err != nil {
		return nil, err
	}

	t := &KubernetesTarget{
		kubeClient: kubeClient,
		patcher:    patcher,
	}
	return t, nil
}

func (s *KubernetesTarget) Read(kind, namespace, name string) (*v1.PodSpec, error) {
	client := s.kubeClient

	switch strings.ToLower(kind) {
	case "replicaset":
		{
			kind = "ReplicaSet"
			o, err := client.ExtensionsV1beta1().ReplicaSets(namespace).Get(name, meta_v1.GetOptions{})
			if err != nil {
				// TODO: Emit event?
				return nil, err
			}

			return &o.Spec.Template.Spec, nil
		}

	case "daemonset":
		{
			kind = "DaemonSet"
			o, err := client.ExtensionsV1beta1().DaemonSets(namespace).Get(name, meta_v1.GetOptions{})
			if err != nil {
				// TODO: Emit event?
				return nil, err
			}

			return &o.Spec.Template.Spec, nil
		}

	case "deployment":
		{
			kind = "Deployment"
			o, err := client.AppsV1beta1().Deployments(namespace).Get(name, meta_v1.GetOptions{})
			if err != nil {
				// TODO: Emit event?
				return nil, err
			}

			return &o.Spec.Template.Spec, nil
		}

	default:
		return nil, fmt.Errorf("unhandled kind: %q", kind)
	}
}

func (s *KubernetesTarget) UpdateResources(kind, namespace, name string, updates *v1.PodSpec, dryrun bool) error {
	return s.patcher.UpdateResources(kind, namespace, name, updates, dryrun)
}

func (s *KubernetesTarget) ReadClusterState() (*ClusterStats, error) {
	nodes, err := s.kubeClient.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %v", err)
	}

	stats := &ClusterStats{
		NodeCount:          0,
		NodeSumAllocatable: make(v1.ResourceList),
	}
	for i := range nodes.Items {
		node := &nodes.Items[i]

		stats.NodeCount++
		addResourceList(stats.NodeSumAllocatable, node.Status.Allocatable)
	}

	glog.V(4).Infof("kubernetes cluster state: %v", stats)

	return stats, nil
}

func addResourceList(sum v1.ResourceList, inc v1.ResourceList) {
	for k, v := range inc {
		a, found := sum[k]
		if !found {
			sum[k] = v
		} else {
			a.Add(v)
			sum[k] = a
		}
	}
}
