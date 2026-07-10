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

package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	v1 "k8s.io/api/coordination/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
)

type AzureResourceLocker struct {
	*Cloud
	holder               string
	leaseName            string
	leaseNamespace       string
	leaseDurationSeconds int32
}

func NewAzureResourceLocker(
	cloud *Cloud,
	holder, leaseName, leaseNamespace string,
	leaseDurationSeconds int32,
) *AzureResourceLocker {
	return &AzureResourceLocker{
		Cloud:                cloud,
		holder:               holder,
		leaseName:            leaseName,
		leaseNamespace:       leaseNamespace,
		leaseDurationSeconds: leaseDurationSeconds,
	}
}

// Lock creates a lease if it does not exist and acquires the lease.
// If the lease has not expired yet and is held by another holder, it will return an error.
func (l *AzureResourceLocker) Lock(ctx context.Context) error {
	if err := createLeaseIfNotExists(
		ctx, l.leaseNamespace, l.leaseName, l.leaseDurationSeconds, l.KubeClient,
	); err != nil {
		return err
	}
	if err := l.acquireLease(
		ctx, l.KubeClient, l.holder, l.leaseNamespace, l.leaseName); err != nil {
		return err
	}

	return nil
}

// Unlock releases the lease if needed.
func (l *AzureResourceLocker) Unlock(ctx context.Context) error {
	if err := releaseLease(
		ctx, l.KubeClient, l.leaseNamespace, l.leaseName, l.holder,
	); err != nil {
		return err
	}

	return nil
}

func createLeaseIfNotExists(
	ctx context.Context,
	leaseNamespace, leaseName string,
	leaseDurationSeconds int32,
	clientset kubernetes.Interface,
) error {
	logger := log.FromContextOrBackground(ctx).WithName("createLeaseIfNotExists").
		WithValues("leaseNamespace", leaseNamespace, "leaseName", leaseName,
			"leaseDurationSeconds", leaseDurationSeconds).V(4)

	lease := &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      leaseName,
			Namespace: leaseNamespace,
		},
		Spec: v1.LeaseSpec{
			LeaseDurationSeconds: ptr.To(leaseDurationSeconds),
		},
	}

	_, err := clientset.CoordinationV1().Leases(leaseNamespace).Get(ctx, leaseName, metav1.GetOptions{})
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "failed to get lease")
			return err
		}
		// Lease does not exist, create it
		_, err = clientset.CoordinationV1().Leases(leaseNamespace).Create(ctx, lease, metav1.CreateOptions{})
		if err != nil {
			logger.Error(err, "failed to create lease")
			return err
		}
		logger.Info("Successfully created lease.")
	} else {
		logger.Info("Lease already exists.")
	}

	return nil
}

func (l *AzureResourceLocker) acquireLease(
	ctx context.Context,
	clientset kubernetes.Interface,
	holder, leaseNamespace, leaseName string,
) error {
	logger := log.FromContextOrBackground(ctx).WithName("acquireLease").
		WithValues("holder", holder, "leaseNamespace", leaseNamespace, "leaseName", leaseName).V(4)

	lease, err := clientset.CoordinationV1().Leases(leaseNamespace).Get(ctx, leaseName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get lease")
		return err
	}

	if !hasExpired(lease) {
		// If the lease has not expired, but the previous holder is the same,
		// it means the same previous operation failed and the releaseLease
		// also failed. In this case we should continue acquiring the lease.
		// This should be a rare case, otherwise we should not acquire the lease.
		prevHolder := ptr.Deref(lease.Spec.HolderIdentity, "")
		if !strings.EqualFold(prevHolder, holder) {
			errMsg := "lease has not expired yet, another component such as aks rp may be processing another request, this would be automatically recovered after the lease expires"
			err := errors.New(errMsg)
			logger.Error(err, "failed to acquire lease")
			return err
		}
	}

	lease.Spec.HolderIdentity = ptr.To(holder)
	now := &metav1.MicroTime{Time: time.Now()}
	lease.Spec.AcquireTime = now
	lease.Spec.RenewTime = now
	_, err = clientset.CoordinationV1().
		Leases(leaseNamespace).Update(ctx, lease, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to acquire lease")
		return err
	}

	// invalidate caches if the previous holder was not the cloud controller manager
	if lease.Annotations != nil &&
		!strings.EqualFold(
			lease.Annotations[consts.AzureResourceLockPreviousHolderNameAnnotation],
			consts.AzureResourceLockHolderNameCloudControllerManager,
		) {
		l.lbCache, err = l.newLBCache()
		if err != nil {
			return err
		}
		if err := l.VMSet.RefreshCaches(); err != nil {
			return err
		}
	}

	logger.Info("Successfully acquired lease.")
	return nil
}

func hasExpired(lease *v1.Lease) bool {
	if ptr.Deref(lease.Spec.HolderIdentity, "") == "" {
		return true // No holder, hence expired
	}

	// Calculate the expiration time
	leaseDuration := time.Duration(ptr.Deref(lease.Spec.LeaseDurationSeconds, 0)) * time.Second
	var expirationTime time.Time
	if lease.Spec.RenewTime != nil {
		expirationTime = lease.Spec.RenewTime.Add(leaseDuration)
	}

	return time.Now().After(expirationTime)
}

func releaseLease(
	ctx context.Context,
	clientset kubernetes.Interface,
	leaseNamespace, leaseName, holder string,
) error {
	logger := log.FromContextOrBackground(ctx).WithName("releaseLease").
		WithValues("leaseNamespace", leaseNamespace, "leaseName", leaseName, "holder", holder).V(4)

	lease, err := clientset.CoordinationV1().
		Leases(leaseNamespace).Get(ctx, leaseName, metav1.GetOptions{})
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			logger.Info("lease does not exist, no need to release it.")
			return nil
		}
		logger.Error(err, "failed to get lease")
		return err
	}

	prevHolder := ptr.Deref(lease.Spec.HolderIdentity, "")
	if !strings.EqualFold(prevHolder, holder) {
		logger.Info(
			"lease is already held by a different holder, no need to release it.",
			"requestedHolder", holder, "leaseHolder", prevHolder,
		)
		return nil
	}

	if hasExpired(lease) {
		logger.Info("lease has already expired, no need to release it.")
		return nil
	}

	lease.Spec.HolderIdentity = ptr.To("")
	if lease.Annotations == nil {
		lease.Annotations = make(map[string]string)
	}
	lease.Annotations[consts.AzureResourceLockPreviousHolderNameAnnotation] = consts.AzureResourceLockHolderNameCloudControllerManager
	_, err = clientset.CoordinationV1().
		Leases(leaseNamespace).Update(ctx, lease, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to release lease")
		return err
	}

	logger.Info("Successfully released lease.")
	return nil
}
