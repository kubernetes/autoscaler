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
	"net"
	"time"

	apicoordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	typedcoordinationv1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	"k8s.io/utils/pointer"
)

const (
	// AdmissionControllerStatusName is the name of
	// the Admission Controller status object.
	AdmissionControllerStatusName = "vpa-admission-controller"
	// AdmissionControllerStatusNamespace is the namespace of
	// the Admission Controller status object.
	AdmissionControllerStatusNamespace = "kube-system"
	// AdmissionControllerStatusTimeout is a time after which
	// if not updated the Admission Controller status is no longer valid.
	AdmissionControllerStatusTimeout = 1 * time.Minute

	// Parameters for retrying with exponential backoff.
	retryBackoffInitialDuration = 100 * time.Millisecond
	retryBackoffFactor          = 3
	retryBackoffJitter          = 0
	retryBackoffSteps           = 3
)

// Client for the status object.
type Client struct {
	client               typedcoordinationv1.LeaseInterface
	leaseName            string
	leaseNamespace       string
	leaseDurationSeconds int32
	holderIdentity       string
}

// Validator for the status object.
type Validator interface {
	IsStatusValid(statusTimeout time.Duration) (bool, error)
}

// NewClient returns a client for the status object.
func NewClient(c clientset.Interface, leaseName, leaseNamespace string, leaseDuration time.Duration, holderIdentity string) *Client {
	return &Client{
		client:               c.CoordinationV1().Leases(leaseNamespace),
		leaseName:            leaseName,
		leaseNamespace:       leaseNamespace,
		leaseDurationSeconds: int32(leaseDuration.Seconds()),
		holderIdentity:       holderIdentity,
	}
}

// NewValidator returns a validator for the status object.
func NewValidator(c clientset.Interface, leaseName, leaseNamespace string) Validator {
	return &Client{
		client:               c.CoordinationV1().Leases(leaseNamespace),
		leaseName:            leaseName,
		leaseNamespace:       leaseNamespace,
		leaseDurationSeconds: 0,
		holderIdentity:       "",
	}
}

// UpdateStatus renews status object lease.
// Status object will be created if it doesn't exist.
func (c *Client) UpdateStatus() error {
	updateFn := func() error {
		lease, err := c.client.Get(c.leaseName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// Create lease if it doesn't exist.
			return c.create()
		} else if err != nil {
			return err
		}
		lease.Spec.RenewTime = &metav1.MicroTime{Time: time.Now()}
		lease.Spec.HolderIdentity = pointer.StringPtr(c.holderIdentity)
		_, err = c.client.Update(lease)
		if apierrors.IsConflict(err) {
			// Lease was updated by an another replica of the component.
			// No error should be returned.
			return nil
		}
		return err
	}
	return retryWithExponentialBackOff(updateFn)
}

func (c *Client) create() error {
	_, err := c.client.Create(c.newLease())
	if apierrors.IsAlreadyExists(err) {
		// Lease was created by an another replica of the component.
		// No error should be returned.
		return nil
	}
	return err
}

func (c *Client) newLease() *apicoordinationv1.Lease {
	return &apicoordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.leaseName,
			Namespace: c.leaseNamespace,
		},
		Spec: apicoordinationv1.LeaseSpec{
			HolderIdentity:       pointer.StringPtr(c.holderIdentity),
			LeaseDurationSeconds: pointer.Int32Ptr(c.leaseDurationSeconds),
		},
	}
}

// IsStatusValid verifies if the current status object
// was updated before lease timing out.
func (c *Client) IsStatusValid(statusTimeout time.Duration) (bool, error) {
	status, err := c.getStatus()
	if err != nil {
		return false, err
	}
	return isStatusValid(status, statusTimeout, time.Now()), nil
}

func (c *Client) getStatus() (*apicoordinationv1.Lease, error) {
	var lease *apicoordinationv1.Lease
	getFn := func() error {
		var err error
		lease, err = c.client.Get(c.leaseName, metav1.GetOptions{})
		return err
	}
	err := retryWithExponentialBackOff(getFn)
	return lease, err
}

func isStatusValid(status *apicoordinationv1.Lease, leaseTimeout time.Duration, now time.Time) bool {
	return status.CreationTimestamp.Add(leaseTimeout).After(now) ||
		(status.Spec.RenewTime != nil &&
			status.Spec.RenewTime.Add(leaseTimeout).After(now))
}

func isRetryableAPIError(err error) bool {
	// These errors may indicate a transient error that we can retry.
	if apierrors.IsInternalError(err) || apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) ||
		apierrors.IsTooManyRequests(err) || utilnet.IsProbableEOF(err) || utilnet.IsConnectionReset(err) {
		return true
	}
	// If the error sends the Retry-After header, we respect it as an explicit confirmation we should retry.
	if _, shouldRetry := apierrors.SuggestsClientDelay(err); shouldRetry {
		return true
	}
	return false
}

func isRetryableNetError(err error) bool {
	if netError, ok := err.(net.Error); ok {
		return netError.Temporary() || netError.Timeout()
	}
	return false
}

func retryWithExponentialBackOff(fn func() error) error {
	backoff := wait.Backoff{
		Duration: retryBackoffInitialDuration,
		Factor:   retryBackoffFactor,
		Jitter:   retryBackoffJitter,
		Steps:    retryBackoffSteps,
	}
	retryFn := func(fn func() error) func() (bool, error) {
		return func() (bool, error) {
			err := fn()
			if err == nil {
				return true, nil
			}
			if isRetryableAPIError(err) || isRetryableNetError(err) {
				return false, nil
			}
			return false, err
		}
	}
	return wait.ExponentialBackoff(backoff, retryFn(fn))
}
