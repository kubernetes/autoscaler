/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	discovery_v1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
	utilsets "sigs.k8s.io/cloud-provider-azure/pkg/util/sets"
)

// batchProcessor collects operations in a certain interval and then processes them in batches.
type batchProcessor interface {
	// run starts the batchProcessor, and stops if the context exits.
	run(ctx context.Context)

	// addOperation adds an operation to the batchProcessor.
	addOperation(operation batchOperation) batchOperation

	// removeOperation removes all operations targeting to the specified service.
	removeOperation(name string)
}

// batchOperation is an operation that can be added to a batchProcessor.
type batchOperation interface {
	wait() batchOperationResult
}

// loadBalancerBackendPoolUpdateOperation is an operation that updates the backend pool of a load balancer.
type loadBalancerBackendPoolUpdateOperation struct {
	serviceName      string
	loadBalancerName string
	backendPoolName  string
	kind             consts.LoadBalancerBackendPoolUpdateOperation
	nodeIPs          []string
}

func (op *loadBalancerBackendPoolUpdateOperation) wait() batchOperationResult {
	return batchOperationResult{}
}

// loadBalancerBackendPoolUpdater is a batchProcessor that updates the backend pool of a load balancer.
type loadBalancerBackendPoolUpdater struct {
	az         *Cloud
	interval   time.Duration
	lock       sync.Mutex
	operations []batchOperation
}

// newLoadBalancerBackendPoolUpdater creates a new loadBalancerBackendPoolUpdater.
func newLoadBalancerBackendPoolUpdater(az *Cloud, interval time.Duration) *loadBalancerBackendPoolUpdater {
	return &loadBalancerBackendPoolUpdater{
		az:         az,
		interval:   interval,
		operations: make([]batchOperation, 0),
	}
}

// run starts the loadBalancerBackendPoolUpdater, and stops if the context exits.
func (updater *loadBalancerBackendPoolUpdater) run(ctx context.Context) {
	klog.V(2).Info("loadBalancerBackendPoolUpdater.run: started")
	err := wait.PollUntilContextCancel(ctx, updater.interval, false, func(ctx context.Context) (bool, error) {
		updater.process()
		return false, nil
	})
	klog.Infof("loadBalancerBackendPoolUpdater.run: stopped due to %s", err.Error())
}

// getAddIPsToBackendPoolOperation creates a new loadBalancerBackendPoolUpdateOperation
// that adds nodeIPs to the backend pool.
func getAddIPsToBackendPoolOperation(serviceName, loadBalancerName, backendPoolName string, nodeIPs []string) *loadBalancerBackendPoolUpdateOperation {
	return &loadBalancerBackendPoolUpdateOperation{
		serviceName:      serviceName,
		loadBalancerName: loadBalancerName,
		backendPoolName:  backendPoolName,
		kind:             consts.LoadBalancerBackendPoolUpdateOperationAdd,
		nodeIPs:          nodeIPs,
	}
}

// getRemoveIPsFromBackendPoolOperation creates a new loadBalancerBackendPoolUpdateOperation
// that removes nodeIPs from the backend pool.
func getRemoveIPsFromBackendPoolOperation(serviceName, loadBalancerName, backendPoolName string, nodeIPs []string) *loadBalancerBackendPoolUpdateOperation {
	return &loadBalancerBackendPoolUpdateOperation{
		serviceName:      serviceName,
		loadBalancerName: loadBalancerName,
		backendPoolName:  backendPoolName,
		kind:             consts.LoadBalancerBackendPoolUpdateOperationRemove,
		nodeIPs:          nodeIPs,
	}
}

// addOperation adds an operation to the loadBalancerBackendPoolUpdater.
func (updater *loadBalancerBackendPoolUpdater) addOperation(operation batchOperation) batchOperation {
	updater.lock.Lock()
	defer updater.lock.Unlock()

	op := operation.(*loadBalancerBackendPoolUpdateOperation)
	klog.V(4).InfoS("loadBalancerBackendPoolUpdater.addOperation",
		"kind", op.kind,
		"service name", op.serviceName,
		"load balancer name", op.loadBalancerName,
		"backend pool name", op.backendPoolName,
		"node IPs", strings.Join(op.nodeIPs, ","))
	updater.operations = append(updater.operations, operation)
	return operation
}

// removeOperation removes all operations targeting to the specified service.
func (updater *loadBalancerBackendPoolUpdater) removeOperation(serviceName string) {
	updater.lock.Lock()
	defer updater.lock.Unlock()

	for i := len(updater.operations) - 1; i >= 0; i-- {
		op := updater.operations[i].(*loadBalancerBackendPoolUpdateOperation)
		if strings.EqualFold(op.serviceName, serviceName) {
			klog.V(4).InfoS("loadBalancerBackendPoolUpdater.removeOperation",
				"kind", op.kind,
				"service name", op.serviceName,
				"load balancer name", op.loadBalancerName,
				"backend pool name", op.backendPoolName,
				"node IPs", strings.Join(op.nodeIPs, ","))
			updater.operations = append(updater.operations[:i], updater.operations[i+1:]...)
		}
	}
}

// process processes all operations in the loadBalancerBackendPoolUpdater.
// It merges operations that have the same loadBalancerName and backendPoolName,
// and then processes them in batches. If an operation fails, it will be retried
// if it is retriable, otherwise all operations in the batch targeting to
// this backend pool will fail.
func (updater *loadBalancerBackendPoolUpdater) process() {
	updater.lock.Lock()
	defer updater.lock.Unlock()

	if len(updater.operations) == 0 {
		klog.V(4).Infof("loadBalancerBackendPoolUpdater.process: no operations to process")
		return
	}

	// Group operations by loadBalancerName:backendPoolName
	groups := make(map[string][]batchOperation)
	for _, op := range updater.operations {
		lbOp := op.(*loadBalancerBackendPoolUpdateOperation)
		si, found := updater.az.getLocalServiceInfo(strings.ToLower(lbOp.serviceName))
		if !found {
			klog.V(4).Infof("loadBalancerBackendPoolUpdater.process: service %s is not a local service, skip the operation", lbOp.serviceName)
			continue
		}
		if !strings.EqualFold(si.lbName, lbOp.loadBalancerName) {
			klog.V(4).InfoS("loadBalancerBackendPoolUpdater.process: service is not associated with the load balancer, skip the operation",
				"service", lbOp.serviceName,
				"previous load balancer", lbOp.loadBalancerName,
				"current load balancer", si.lbName)
			continue
		}

		key := fmt.Sprintf("%s:%s", lbOp.loadBalancerName, lbOp.backendPoolName)
		groups[key] = append(groups[key], op)
	}

	// Clear all jobs.
	updater.operations = make([]batchOperation, 0)

	for key, ops := range groups {
		parts := strings.Split(key, ":")
		lbName, poolName := parts[0], parts[1]
		operationName := fmt.Sprintf("%s/%s", lbName, poolName)
		bp, rerr := updater.az.LoadBalancerClient.GetLBBackendPool(context.Background(), updater.az.ResourceGroup, lbName, poolName, "")
		if rerr != nil {
			updater.processError(rerr, operationName, ops...)
			continue
		}

		var changed bool
		for _, op := range ops {
			lbOp := op.(*loadBalancerBackendPoolUpdateOperation)
			switch lbOp.kind {
			case consts.LoadBalancerBackendPoolUpdateOperationRemove:
				removed := removeNodeIPAddressesFromBackendPool(bp, lbOp.nodeIPs, false, true)
				changed = changed || removed
			case consts.LoadBalancerBackendPoolUpdateOperationAdd:
				added := updater.az.addNodeIPAddressesToBackendPool(&bp, lbOp.nodeIPs)
				changed = changed || added
			default:
				panic("loadBalancerBackendPoolUpdater.process: unknown operation type")
			}
		}
		// To keep the code clean, ignore the case when `changed` is true
		// but the backend pool object is not changed after multiple times of removal and re-adding.
		if changed {
			klog.V(2).Infof("loadBalancerBackendPoolUpdater.process: updating backend pool %s/%s", lbName, poolName)
			rerr = updater.az.LoadBalancerClient.CreateOrUpdateBackendPools(context.Background(), updater.az.ResourceGroup, lbName, poolName, bp, pointer.StringDeref(bp.Etag, ""))
			if rerr != nil {
				updater.processError(rerr, operationName, ops...)
				continue
			}
		}
		updater.notify(newBatchOperationResult(operationName, true, nil), ops...)
	}
}

// processError mark the operations as retriable if the error is retriable,
// and fail all operations if the error is not retriable.
func (updater *loadBalancerBackendPoolUpdater) processError(
	rerr *retry.Error,
	operationName string,
	operations ...batchOperation,
) {
	if rerr.IsNotFound() {
		klog.V(4).Infof("backend pool not found for operation %s, skip updating", operationName)
		return
	}

	if rerr.Retriable {
		// Retry if retriable.
		updater.operations = append(updater.operations, operations...)
	} else {
		// Fail all operations if not retriable.
		updater.notify(newBatchOperationResult(operationName, false, rerr.Error()), operations...)
	}
}

// notify notifies the operations with the result.
func (updater *loadBalancerBackendPoolUpdater) notify(res batchOperationResult, operations ...batchOperation) {
	for _, op := range operations {
		updater.az.processBatchOperationResult(op, res)
		break
	}
}

// batchOperationResult is the result of a batch operation.
type batchOperationResult struct {
	name    string
	success bool
	err     error
}

// newBatchOperationResult creates a new batchOperationResult.
func newBatchOperationResult(name string, success bool, err error) batchOperationResult {
	return batchOperationResult{
		name:    name,
		success: success,
		err:     err,
	}
}

func (az *Cloud) getLocalServiceInfo(serviceName string) (*serviceInfo, bool) {
	data, ok := az.localServiceNameToServiceInfoMap.Load(serviceName)
	if !ok {
		return &serviceInfo{}, false
	}
	return data.(*serviceInfo), true
}

// setUpEndpointSlicesInformer creates an informer for EndpointSlices of local services.
// It watches the update events and send backend pool update operations to the batch updater.
func (az *Cloud) setUpEndpointSlicesInformer(informerFactory informers.SharedInformerFactory) {
	endpointSlicesInformer := informerFactory.Discovery().V1().EndpointSlices().Informer()
	_, _ = endpointSlicesInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				es := obj.(*discovery_v1.EndpointSlice)
				az.endpointSlicesCache.Store(strings.ToLower(fmt.Sprintf("%s/%s", es.Namespace, es.Name)), es)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				previousES := oldObj.(*discovery_v1.EndpointSlice)
				newES := newObj.(*discovery_v1.EndpointSlice)

				svcName := getServiceNameOfEndpointSlice(newES)
				if svcName == "" {
					klog.V(4).Infof("EndpointSlice %s/%s does not have service name label, skip updating load balancer backend pool", newES.Namespace, newES.Name)
					return
				}

				klog.V(4).Infof("Detecting EndpointSlice %s/%s update", newES.Namespace, newES.Name)
				az.endpointSlicesCache.Store(strings.ToLower(fmt.Sprintf("%s/%s", newES.Namespace, newES.Name)), newES)

				key := strings.ToLower(fmt.Sprintf("%s/%s", newES.Namespace, svcName))
				si, found := az.getLocalServiceInfo(key)
				if !found {
					klog.V(4).Infof("EndpointSlice %s/%s belongs to service %s, but the service is not a local service, or has not finished the initial reconciliation loop. Skip updating load balancer backend pool", newES.Namespace, newES.Name, key)
					return
				}
				lbName, ipFamily := si.lbName, si.ipFamily

				var previousIPs, currentIPs, previousNodeNames, currentNodeNames []string
				if previousES != nil {
					for _, ep := range previousES.Endpoints {
						previousNodeNames = append(previousNodeNames, pointer.StringDeref(ep.NodeName, ""))
					}
				}
				if newES != nil {
					for _, ep := range newES.Endpoints {
						currentNodeNames = append(currentNodeNames, pointer.StringDeref(ep.NodeName, ""))
					}
				}
				for _, previousNodeName := range previousNodeNames {
					nodeIPsSet := az.nodePrivateIPs[strings.ToLower(previousNodeName)]
					previousIPs = append(previousIPs, nodeIPsSet.UnsortedList()...)
				}
				for _, currentNodeName := range currentNodeNames {
					nodeIPsSet := az.nodePrivateIPs[strings.ToLower(currentNodeName)]
					currentIPs = append(currentIPs, nodeIPsSet.UnsortedList()...)
				}

				if az.backendPoolUpdater != nil {
					var bpNames []string
					bpNameIPv4 := getLocalServiceBackendPoolName(key, false)
					bpNameIPv6 := getLocalServiceBackendPoolName(key, true)
					switch strings.ToLower(ipFamily) {
					case strings.ToLower(consts.IPVersionIPv4String):
						bpNames = append(bpNames, bpNameIPv4)
					case strings.ToLower(consts.IPVersionIPv6String):
						bpNames = append(bpNames, bpNameIPv6)
					default:
						bpNames = append(bpNames, bpNameIPv4, bpNameIPv6)
					}
					currentIPsInBackendPools := make(map[string][]string)
					for _, bpName := range bpNames {
						currentIPsInBackendPools[bpName] = previousIPs
					}
					az.applyIPChangesAmongLocalServiceBackendPoolsByIPFamily(lbName, key, currentIPsInBackendPools, currentIPs)
				}
			},
			DeleteFunc: func(obj interface{}) {
				es := obj.(*discovery_v1.EndpointSlice)
				az.endpointSlicesCache.Delete(strings.ToLower(fmt.Sprintf("%s/%s", es.Namespace, es.Name)))
			},
		})
}

func (az *Cloud) processBatchOperationResult(op batchOperation, res batchOperationResult) {
	lbOp := op.(*loadBalancerBackendPoolUpdateOperation)
	var svc *v1.Service
	svc, _, _ = az.getLatestService(lbOp.serviceName, false)
	if svc == nil {
		klog.Warningf("Service %s not found, skip sending event", lbOp.serviceName)
		return
	}
	if !res.success {
		var errStr string
		if res.err != nil {
			errStr = res.err.Error()
		}
		az.Event(svc, v1.EventTypeWarning, "LoadBalancerBackendPoolUpdateFailed", errStr)
	} else {
		az.Event(svc, v1.EventTypeNormal, "LoadBalancerBackendPoolUpdated", "Load balancer backend pool updated successfully")
	}
}

// getServiceNameOfEndpointSlice gets the service name of an EndpointSlice.
func getServiceNameOfEndpointSlice(es *discovery_v1.EndpointSlice) string {
	if es.Labels != nil {
		return es.Labels[consts.ServiceNameLabel]
	}
	return ""
}

// compareNodeIPs compares the previous and current node IPs and returns the IPs to be deleted.
func compareNodeIPs(previousIPs, currentIPs []string) []string {
	previousIPSet := sets.NewString(previousIPs...)
	currentIPSet := sets.NewString(currentIPs...)
	return previousIPSet.Difference(currentIPSet).List()
}

// getLocalServiceBackendPoolName gets the name of the backend pool of a local service.
func getLocalServiceBackendPoolName(serviceName string, ipv6 bool) string {
	serviceName = strings.ToLower(strings.Replace(serviceName, "/", "-", -1))
	if ipv6 {
		return fmt.Sprintf("%s-%s", serviceName, consts.IPVersionIPv6StringLower)
	}
	return serviceName
}

// getBackendPoolNameForService determine the expected backend pool name
// by checking the external traffic policy of the service.
func (az *Cloud) getBackendPoolNameForService(service *v1.Service, clusterName string, ipv6 bool) string {
	if !isLocalService(service) || !az.useMultipleStandardLoadBalancers() {
		return getBackendPoolName(clusterName, ipv6)
	}
	return getLocalServiceBackendPoolName(getServiceName(service), ipv6)
}

// getBackendPoolNamesForService determine the expected backend pool names
// by checking the external traffic policy of the service.
func (az *Cloud) getBackendPoolNamesForService(service *v1.Service, clusterName string) map[bool]string {
	if !isLocalService(service) || !az.useMultipleStandardLoadBalancers() {
		return getBackendPoolNames(clusterName)
	}
	return map[bool]string{
		consts.IPVersionIPv4: getLocalServiceBackendPoolName(getServiceName(service), false),
		consts.IPVersionIPv6: getLocalServiceBackendPoolName(getServiceName(service), true),
	}
}

// getBackendPoolIDsForService determine the expected backend pool IDs
// by checking the external traffic policy of the service.
func (az *Cloud) getBackendPoolIDsForService(service *v1.Service, clusterName, lbName string) map[bool]string {
	if !isLocalService(service) || !az.useMultipleStandardLoadBalancers() {
		return az.getBackendPoolIDs(clusterName, lbName)
	}
	return map[bool]string{
		consts.IPVersionIPv4: az.getLocalServiceBackendPoolID(getServiceName(service), lbName, false),
		consts.IPVersionIPv6: az.getLocalServiceBackendPoolID(getServiceName(service), lbName, true),
	}
}

// getLocalServiceBackendPoolID gets the ID of the backend pool of a local service.
func (az *Cloud) getLocalServiceBackendPoolID(serviceName string, lbName string, ipv6 bool) string {
	return az.getBackendPoolID(lbName, getLocalServiceBackendPoolName(serviceName, ipv6))
}

// localServiceOwnsBackendPool checks if a backend pool is owned by a local service.
func localServiceOwnsBackendPool(serviceName, bpName string) bool {
	prefix := strings.Replace(serviceName, "/", "-", -1)
	return strings.HasPrefix(strings.ToLower(bpName), strings.ToLower(prefix))
}

type serviceInfo struct {
	ipFamily string
	lbName   string
}

func newServiceInfo(ipFamily, lbName string) *serviceInfo {
	return &serviceInfo{
		ipFamily: ipFamily,
		lbName:   lbName,
	}
}

// getLocalServiceEndpointsNodeNames gets the node names that host all endpoints of the local service.
func (az *Cloud) getLocalServiceEndpointsNodeNames(service *v1.Service) (*utilsets.IgnoreCaseSet, error) {
	var (
		ep           *discovery_v1.EndpointSlice
		foundInCache bool
	)
	az.endpointSlicesCache.Range(func(key, value interface{}) bool {
		endpointSlice := value.(*discovery_v1.EndpointSlice)
		if strings.EqualFold(getServiceNameOfEndpointSlice(endpointSlice), service.Name) &&
			strings.EqualFold(endpointSlice.Namespace, service.Namespace) {
			ep = endpointSlice
			foundInCache = true
			return false
		}
		return true
	})
	if ep == nil {
		klog.Infof("EndpointSlice for service %s/%s not found, try to list EndpointSlices", service.Namespace, service.Name)
		eps, err := az.KubeClient.DiscoveryV1().EndpointSlices(service.Namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list EndpointSlices for service %s/%s: %s", service.Namespace, service.Name, err.Error())
			return nil, err
		}
		for _, endpointSlice := range eps.Items {
			endpointSlice := endpointSlice
			if strings.EqualFold(getServiceNameOfEndpointSlice(&endpointSlice), service.Name) {
				ep = &endpointSlice
				break
			}
		}
	}
	if ep == nil {
		return nil, fmt.Errorf("failed to find EndpointSlice for service %s/%s", service.Namespace, service.Name)
	}
	if !foundInCache {
		az.endpointSlicesCache.Store(strings.ToLower(fmt.Sprintf("%s/%s", ep.Namespace, ep.Name)), ep)
	}

	var nodeNames []string
	for _, endpoint := range ep.Endpoints {
		klog.V(4).Infof("EndpointSlice %s/%s has endpoint %s on node %s", ep.Namespace, ep.Name, endpoint.Addresses, pointer.StringDeref(endpoint.NodeName, ""))
		nodeNames = append(nodeNames, pointer.StringDeref(endpoint.NodeName, ""))
	}

	return utilsets.NewString(nodeNames...), nil
}

// cleanupLocalServiceBackendPool cleans up the backend pool of
// a local service among given load balancers.
func (az *Cloud) cleanupLocalServiceBackendPool(
	svc *v1.Service,
	nodes []*v1.Node,
	lbs *[]network.LoadBalancer,
	clusterName string,
) (newLBs *[]network.LoadBalancer, err error) {
	var changed bool
	if lbs != nil {
		for _, lb := range *lbs {
			lbName := pointer.StringDeref(lb.Name, "")
			if lb.BackendAddressPools != nil {
				for _, bp := range *lb.BackendAddressPools {
					bpName := pointer.StringDeref(bp.Name, "")
					if localServiceOwnsBackendPool(getServiceName(svc), bpName) {
						if err := az.DeleteLBBackendPool(lbName, bpName); err != nil {
							return nil, err
						}
						changed = true
					}
				}
			}
		}
	}
	if changed {
		// Refresh the list of existing LBs after cleanup to update etags for the LBs.
		klog.V(4).Info("Refreshing the list of existing LBs")
		lbs, err = az.ListManagedLBs(svc, nodes, clusterName)
		if err != nil {
			return nil, fmt.Errorf("reconcileLoadBalancer: failed to list managed LB: %w", err)
		}
	}
	return lbs, nil
}

// checkAndApplyLocalServiceBackendPoolUpdates if the IPs in the backend pool are aligned
// with the corresponding endpointslice, and update the backend pool if necessary.
func (az *Cloud) checkAndApplyLocalServiceBackendPoolUpdates(lb network.LoadBalancer, service *v1.Service) error {
	serviceName := getServiceName(service)
	endpointsNodeNames, err := az.getLocalServiceEndpointsNodeNames(service)
	if err != nil {
		return err
	}
	var expectedIPs []string
	for _, nodeName := range endpointsNodeNames.UnsortedList() {
		ips := az.nodePrivateIPs[strings.ToLower(nodeName)]
		expectedIPs = append(expectedIPs, ips.UnsortedList()...)
	}
	currentIPsInBackendPools := make(map[string][]string)
	for _, bp := range *lb.BackendAddressPools {
		bpName := pointer.StringDeref(bp.Name, "")
		if localServiceOwnsBackendPool(serviceName, bpName) {
			var currentIPs []string
			for _, address := range *bp.LoadBalancerBackendAddresses {
				currentIPs = append(currentIPs, *address.IPAddress)
			}
			currentIPsInBackendPools[bpName] = currentIPs
		}
	}
	az.applyIPChangesAmongLocalServiceBackendPoolsByIPFamily(*lb.Name, serviceName, currentIPsInBackendPools, expectedIPs)

	return nil
}

// applyIPChangesAmongLocalServiceBackendPoolsByIPFamily reconciles IPs by IP family
// amone the backend pools of a local service.
func (az *Cloud) applyIPChangesAmongLocalServiceBackendPoolsByIPFamily(
	lbName, serviceName string,
	currentIPsInBackendPools map[string][]string,
	expectedIPs []string,
) {
	currentIPsInBackendPoolsIPv4 := make(map[string][]string)
	currentIPsInBackendPoolsIPv6 := make(map[string][]string)
	for bpName, ips := range currentIPsInBackendPools {
		if managedResourceHasIPv6Suffix(bpName) {
			currentIPsInBackendPoolsIPv6[bpName] = ips
		} else {
			currentIPsInBackendPoolsIPv4[bpName] = ips
		}
	}

	var ipv4, ipv6 []string
	for _, ip := range expectedIPs {
		if utilnet.IsIPv6String(ip) {
			ipv6 = append(ipv6, ip)
		} else {
			ipv4 = append(ipv4, ip)
		}
	}
	az.reconcileIPsInLocalServiceBackendPoolsAsync(lbName, serviceName, currentIPsInBackendPoolsIPv6, ipv6)
	az.reconcileIPsInLocalServiceBackendPoolsAsync(lbName, serviceName, currentIPsInBackendPoolsIPv4, ipv4)
}

// reconcileIPsInLocalServiceBackendPoolsAsync reconciles IPs in the backend pools of a local service.
func (az *Cloud) reconcileIPsInLocalServiceBackendPoolsAsync(
	lbName, serviceName string,
	currentIPsInBackendPools map[string][]string,
	expectedIPs []string,
) {
	for bpName, currentIPs := range currentIPsInBackendPools {
		ipsToBeDeleted := compareNodeIPs(currentIPs, expectedIPs)
		if len(ipsToBeDeleted) == 0 && len(currentIPs) == len(expectedIPs) {
			klog.V(4).Infof("No IP change detected for service %s, skip updating load balancer backend pool", serviceName)
			return
		}
		if len(ipsToBeDeleted) > 0 {
			az.backendPoolUpdater.addOperation(getRemoveIPsFromBackendPoolOperation(serviceName, lbName, bpName, ipsToBeDeleted))
		}
		if len(expectedIPs) > 0 {
			az.backendPoolUpdater.addOperation(getAddIPsToBackendPoolOperation(serviceName, lbName, bpName, expectedIPs))
		}
	}
}
