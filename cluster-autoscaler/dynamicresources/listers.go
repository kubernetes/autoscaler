/*
Copyright 2024 The Kubernetes Authors.

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

package dynamicresources

import (
	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/labels"
	resourceapilisters "k8s.io/client-go/listers/resource/v1alpha3"
)

// These listers are introduced for Provider dependency injection.

// providerClaimLister is a subset of ResourceClaimLister needed by the Provider.
type providerClaimLister interface {
	List() ([]*resourceapi.ResourceClaim, error)
}

// providerClaimLister is a subset of ResourceSliceLister needed by the Provider.
type providerSliceLister interface {
	List() ([]*resourceapi.ResourceSlice, error)
}

// providerClassLister is a subset of DeviceClassLister needed by the Provider.
type providerClassLister interface {
	List() ([]*resourceapi.DeviceClass, error)
}

// resourceClaimApiLister implements providerClaimLister using a real ResourceClaimLister listing from the API.
type resourceClaimApiLister struct {
	apiLister resourceapilisters.ResourceClaimLister
}

// List lists all ResourceClaims.
func (l *resourceClaimApiLister) List() ([]*resourceapi.ResourceClaim, error) {
	return l.apiLister.List(labels.Everything())
}

// resourceSliceApiLister implements providerSliceLister using a real ResourceSliceLister listing from the API.
type resourceSliceApiLister struct {
	apiLister resourceapilisters.ResourceSliceLister
}

// List lists all ResourceSlices.
func (l *resourceSliceApiLister) List() (ret []*resourceapi.ResourceSlice, err error) {
	return l.apiLister.List(labels.Everything())
}

// deviceClassApiLister implements providerClassLister using a real DeviceClassLister listing from the API.
type deviceClassApiLister struct {
	apiLister resourceapilisters.DeviceClassLister
}

// List lists all DeviceClasses.
func (l *deviceClassApiLister) List() (ret []*resourceapi.DeviceClass, err error) {
	return l.apiLister.List(labels.Everything())
}
