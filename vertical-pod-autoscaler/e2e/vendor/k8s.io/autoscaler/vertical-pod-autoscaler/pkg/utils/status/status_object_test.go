/*
Copyright 2020 The Kubernetes Authors.

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

package status

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apicoordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
)

func TestUpdateStatus(t *testing.T) {
	const (
		leaseName      = "lease"
		leaseNamespace = "default"
	)
	tests := []struct {
		name          string
		updateReactor func(action testcore.Action) (bool, runtime.Object, error)
		wantErr       bool
	}{
		{
			name: "updating status object",
			updateReactor: func(action testcore.Action) (bool, runtime.Object, error) {
				if action.GetResource().Resource == "leases" {
					return true, &apicoordinationv1.Lease{}, nil
				}

				return true, nil, fmt.Errorf("unsupported action")
			},
			wantErr: false,
		},
		{
			name: "updating status object - creating status",
			updateReactor: func() func(action testcore.Action) (bool, runtime.Object, error) {
				i := 0
				return func(action testcore.Action) (bool, runtime.Object, error) {
					if action.GetResource().Resource == "leases" {
						i++
						switch i {
						case 1:
							return true, nil, apierrors.NewNotFound(schema.GroupResource{}, leaseName)
						default:
							return true, &apicoordinationv1.Lease{}, nil
						}
					}

					return true, nil, fmt.Errorf("unsupported action")
				}
			}(),
			wantErr: false,
		},
		{
			// Status doesn't exist but will be created by an other component in the meantime.
			name: "updating status object - creating status, already exists",
			updateReactor: func() func(action testcore.Action) (bool, runtime.Object, error) {
				i := 0
				return func(action testcore.Action) (bool, runtime.Object, error) {
					if action.GetResource().Resource == "leases" {
						i++
						switch i {
						case 1:
							return true, nil, apierrors.NewNotFound(schema.GroupResource{}, leaseName)
						default:
							return true, nil, apierrors.NewAlreadyExists(schema.GroupResource{}, leaseName)
						}
					}

					return true, nil, fmt.Errorf("unsupported action")
				}
			}(),
			wantErr: false,
		},
		{
			// Status is updated by an other component in the meantime.
			name: "updating status object - updating conflict",
			updateReactor: func() func(action testcore.Action) (bool, runtime.Object, error) {
				i := 0
				return func(action testcore.Action) (bool, runtime.Object, error) {
					if action.GetResource().Resource == "leases" {
						i++
						switch i {
						case 1:
							return true, &apicoordinationv1.Lease{}, nil
						default:
							return true, nil, apierrors.NewConflict(schema.GroupResource{}, leaseName, nil)
						}
					}

					return true, nil, fmt.Errorf("unsupported action")
				}
			}(),
			wantErr: false,
		},
		{
			name: "updating status object - endless error",
			updateReactor: func(action testcore.Action) (bool, runtime.Object, error) {
				if action.GetResource().Resource == "leases" {
					return true, nil, apierrors.NewNotFound(schema.GroupResource{}, leaseName)
				}
				return true, nil, fmt.Errorf("unsupported action")
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fc := fake.NewSimpleClientset()
			fc.PrependReactor("get", "leases", tc.updateReactor)
			fc.PrependReactor("create", "leases", tc.updateReactor)
			fc.PrependReactor("update", "leases", tc.updateReactor)
			client := NewClient(fc, leaseName, leaseNamespace, 10*time.Second, leaseName)
			err := client.UpdateStatus()
			assert.True(t, (err != nil) == tc.wantErr, fmt.Sprintf("gotErr: %v, wantErr: %v", (err != nil), tc.wantErr))
		})
	}
}

func TestGetStatus(t *testing.T) {
	const (
		leaseName      = "lease"
		leaseNamespace = "default"
	)
	tests := []struct {
		name       string
		getReactor func(action testcore.Action) (bool, runtime.Object, error)
		wantErr    bool
	}{
		{
			name: "getting status object",
			getReactor: func(action testcore.Action) (bool, runtime.Object, error) {
				if action.GetVerb() == "get" && action.GetResource().Resource == "leases" {
					return true, &apicoordinationv1.Lease{}, nil
				}

				return true, nil, fmt.Errorf("unsupported action")
			},
			wantErr: false,
		},
		{
			name: "getting status object - retryable error",
			getReactor: func() func(action testcore.Action) (bool, runtime.Object, error) {
				i := 0
				return func(action testcore.Action) (bool, runtime.Object, error) {
					if action.GetVerb() == "get" && action.GetResource().Resource == "leases" {
						i++
						switch i {
						case 1, 2:
							return true, nil, syscall.ECONNRESET
						default:
							return true, &apicoordinationv1.Lease{}, nil
						}
					}

					return true, nil, fmt.Errorf("unsupported action")
				}
			}(),
			wantErr: false,
		},
		{
			name: "getting status object - non-retryable error",
			getReactor: func() func(action testcore.Action) (bool, runtime.Object, error) {
				i := 0
				return func(action testcore.Action) (bool, runtime.Object, error) {
					if action.GetVerb() == "get" && action.GetResource().Resource == "leases" {
						i++
						switch i {
						case 1:
							return true, nil, fmt.Errorf("non-retryable error")
						default:
							return true, &apicoordinationv1.Lease{}, nil
						}
					}

					return true, nil, fmt.Errorf("unsupported action")
				}
			}(),
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			fc := fake.NewSimpleClientset()
			fc.PrependReactor("get", "leases", tc.getReactor)
			client := NewClient(fc, leaseName, leaseNamespace, 10*time.Second, leaseName)
			_, err := client.getStatus()
			assert.True(t, (err != nil) == tc.wantErr)
		})
	}
}

func TestIsStatusValid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		lease         *apicoordinationv1.Lease
		leaseTimeout  time.Duration
		expectedValid bool
	}{
		{
			name: "Valid CreationTimestamp",
			lease: &apicoordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: now},
				},
			},
			leaseTimeout:  10 * time.Second,
			expectedValid: true,
		},
		{
			name: "Outdated CreationTimestamp with no RenewTime",
			lease: &apicoordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: now.Add(-time.Minute)},
				},
			},
			leaseTimeout:  10 * time.Second,
			expectedValid: false,
		},
		{
			name: "Valid RenewTime",
			lease: &apicoordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: now.Add(-time.Minute)},
				},
				Spec: apicoordinationv1.LeaseSpec{
					RenewTime: &metav1.MicroTime{Time: now},
				},
			},
			leaseTimeout:  10 * time.Second,
			expectedValid: true,
		},
		{
			name: "Outdated CreationTimestamp and RenewTime",
			lease: &apicoordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.Time{Time: now.Add(-time.Minute)},
				},
				Spec: apicoordinationv1.LeaseSpec{
					RenewTime: &metav1.MicroTime{Time: now.Add(-time.Minute)},
				},
			},
			leaseTimeout:  10 * time.Second,
			expectedValid: false,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			assert.Equal(t, isStatusValid(tc.lease, tc.leaseTimeout, now), tc.expectedValid)
		})
	}
}
