/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"context"
	"fmt"

	"github.com/awslabs/operatorpkg/serrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPersistentVolumeClaim(ctx context.Context, kubeClient client.Client, pod *v1.Pod, volume v1.Volume) (*v1.PersistentVolumeClaim, error) {
	var pvcName string
	switch {
	case volume.PersistentVolumeClaim != nil:
		pvcName = volume.PersistentVolumeClaim.ClaimName
	case volume.Ephemeral != nil:
		// generated name per https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#persistentvolumeclaim-naming
		pvcName = fmt.Sprintf("%s-%s", pod.Name, volume.Name)
	default:
		return nil, nil
	}

	pvc := &v1.PersistentVolumeClaim{}
	if err := kubeClient.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: pvcName}, pvc); err != nil {
		return nil, serrors.Wrap(fmt.Errorf("getting persistent volume claim, %w", err), "PersistentVolumeClaim", klog.KRef("", pvcName))
	}
	return pvc, nil
}
