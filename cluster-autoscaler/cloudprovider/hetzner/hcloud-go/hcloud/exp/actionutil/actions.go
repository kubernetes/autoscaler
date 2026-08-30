package actionutil

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"

// AppendNext return the action and the next actions in a new slice.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func AppendNext(action *hcloud.Action, nextActions []*hcloud.Action) []*hcloud.Action {
	all := make([]*hcloud.Action, 0, 1+len(nextActions))
	all = append(all, action)
	all = append(all, nextActions...)
	return all
}
