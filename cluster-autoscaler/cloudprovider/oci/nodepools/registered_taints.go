/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"regexp"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	taintutil "k8s.io/kubernetes/pkg/util/taints"

	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
)

// RegisteredTaintsGetter returns the initial registered taints for the node pool.
type RegisteredTaintsGetter interface {
	Get(*oke.NodePool) ([]apiv1.Taint, error)
}

// CreateRegisteredTaintsGetter creates a client that can retrieve the defined taints on a node pool
func CreateRegisteredTaintsGetter() RegisteredTaintsGetter {
	return &registeredTaintsGetterImpl{}
}

type registeredTaintsGetterImpl struct{}

func (otg *registeredTaintsGetterImpl) Get(np *oke.NodePool) ([]apiv1.Taint, error) {
	metadata := np.NodeMetadata
	kubeletArgs, ok := metadata["kubelet-extra-args"]

	// if user didn't specify any extra args on kubelet, then we know there can't be any initial taints
	if !ok || len(kubeletArgs) == 0 {
		return []apiv1.Taint{}, nil
	}

	// expression accounts for flags starting with one or two dashes
	// and accounts for whether the value was specified with an equal or a space
	re, err := regexp.Compile(`-?-register-with-taints[= ](\S*)`)
	if err != nil {
		return []apiv1.Taint{}, err
	}

	submatches := re.FindStringSubmatch(kubeletArgs)

	// if we found a match, then the match indexes will be
	// [0] regex match
	// [1] taint info
	if len(submatches) == 2 {
		registerFlagValue := submatches[1]
		taintStrs := strings.Split(registerFlagValue, ",")
		// deleteTaints don't mean anything in this context
		addTaints, _, err := taintutil.ParseTaints(taintStrs)
		if err != nil {
			return []apiv1.Taint{}, err
		}
		return addTaints, nil
	}

	return []apiv1.Taint{}, nil
}
