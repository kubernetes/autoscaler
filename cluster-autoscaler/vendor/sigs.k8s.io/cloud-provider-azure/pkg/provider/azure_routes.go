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

package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
)

var _ cloudprovider.Routes = (*Cloud)(nil)

// routeOperation defines the allowed operations for route updating.
type routeOperation string

// copied to minimize the number of cross reference
// and exceptions in publishing and allowed imports.
const (
	// Route operations.
	routeOperationAdd             routeOperation = "add"
	routeOperationDelete          routeOperation = "delete"
	routeTableOperationUpdateTags routeOperation = "updateRouteTableTags"
)

// delayedRouteOperation defines a delayed route operation which is used in delayedRouteUpdater.
type delayedRouteOperation struct {
	route          network.Route
	routeTableTags map[string]*string
	operation      routeOperation
	result         chan batchOperationResult
}

// wait waits for the operation completion and returns the result.
func (op *delayedRouteOperation) wait() batchOperationResult {
	return <-op.result
}

// delayedRouteUpdater defines a delayed route updater, which batches all the
// route updating operations within "interval" period.
// Example usage:
// op, err := updater.addRouteOperation(routeOperationAdd, route)
// err = op.wait()
type delayedRouteUpdater struct {
	az       *Cloud
	interval time.Duration

	lock           sync.Mutex
	routesToUpdate []batchOperation
}

// newDelayedRouteUpdater creates a new delayedRouteUpdater.
func newDelayedRouteUpdater(az *Cloud, interval time.Duration) batchProcessor {
	return &delayedRouteUpdater{
		az:             az,
		interval:       interval,
		routesToUpdate: make([]batchOperation, 0),
	}
}

// run starts the updater reconciling loop.
func (d *delayedRouteUpdater) run(ctx context.Context) {
	klog.Info("delayedRouteUpdater: started")
	err := wait.PollUntilContextCancel(ctx, d.interval, true, func(ctx context.Context) (bool, error) {
		d.updateRoutes()
		return false, nil
	})
	klog.Infof("delayedRouteUpdater: stopped due to %s", err.Error())
}

// updateRoutes invokes route table client to update all routes.
func (d *delayedRouteUpdater) updateRoutes() {
	d.lock.Lock()
	defer d.lock.Unlock()

	// No need to do any updating.
	if len(d.routesToUpdate) == 0 {
		klog.V(4).Info("updateRoutes: nothing to update, returning")
		return
	}

	var err error
	defer func() {
		// Notify all the goroutines.
		for _, op := range d.routesToUpdate {
			rt := op.(*delayedRouteOperation)
			rt.result <- newBatchOperationResult("", false, err)
		}
		// Clear all the jobs.
		d.routesToUpdate = make([]batchOperation, 0)
	}()

	var (
		routeTable       network.RouteTable
		existsRouteTable bool
	)
	routeTable, existsRouteTable, err = d.az.getRouteTable(azcache.CacheReadTypeDefault)
	if err != nil {
		klog.Errorf("getRouteTable() failed with error: %v", err)
		return
	}

	// create route table if it doesn't exists yet.
	if !existsRouteTable {
		err = d.az.createRouteTable()
		if err != nil {
			klog.Errorf("createRouteTable() failed with error: %v", err)
			return
		}

		routeTable, _, err = d.az.getRouteTable(azcache.CacheReadTypeDefault)
		if err != nil {
			klog.Errorf("getRouteTable() failed with error: %v", err)
			return
		}
	}

	// reconcile routes.
	dirty, onlyUpdateTags := false, true
	routes := []network.Route{}
	if routeTable.RouteTablePropertiesFormat != nil && routeTable.RouteTablePropertiesFormat.Routes != nil {
		routes = *routeTable.Routes
	}

	routes, dirty = d.cleanupOutdatedRoutes(routes)
	if dirty {
		onlyUpdateTags = false
	}

	for _, op := range d.routesToUpdate {
		rt := op.(*delayedRouteOperation)
		if rt.operation == routeTableOperationUpdateTags {
			routeTable.Tags = rt.routeTableTags
			dirty = true
			continue
		}

		routeMatch := false
		onlyUpdateTags = false
		for i, existingRoute := range routes {
			if strings.EqualFold(pointer.StringDeref(existingRoute.Name, ""), pointer.StringDeref(rt.route.Name, "")) {
				// delete the name-matched routes here (missing routes would be added later if the operation is add).
				routes = append(routes[:i], routes[i+1:]...)
				if existingRoute.RoutePropertiesFormat != nil &&
					rt.route.RoutePropertiesFormat != nil &&
					strings.EqualFold(pointer.StringDeref(existingRoute.AddressPrefix, ""), pointer.StringDeref(rt.route.AddressPrefix, "")) &&
					strings.EqualFold(pointer.StringDeref(existingRoute.NextHopIPAddress, ""), pointer.StringDeref(rt.route.NextHopIPAddress, "")) {
					routeMatch = true
				}
				if rt.operation == routeOperationDelete {
					dirty = true
				}
				break
			}
		}
		if rt.operation == routeOperationDelete && !dirty {
			klog.Warningf("updateRoutes: route to be deleted %s does not match any of the existing route", pointer.StringDeref(rt.route.Name, ""))
		}

		// Add missing routes if the operation is add.
		if rt.operation == routeOperationAdd {
			routes = append(routes, rt.route)
			if !routeMatch {
				dirty = true
			}
			continue
		}
	}

	if dirty {
		if !onlyUpdateTags {
			klog.V(2).Infof("updateRoutes: updating routes")
			routeTable.Routes = &routes
		}
		err = d.az.CreateOrUpdateRouteTable(routeTable)
		if err != nil {
			klog.Errorf("CreateOrUpdateRouteTable() failed with error: %v", err)
			return
		}

		// wait a while for route updates to take effect.
		time.Sleep(time.Duration(d.az.Config.RouteUpdateWaitingInSeconds) * time.Second)
	}
}

// cleanupOutdatedRoutes deletes all non-dualstack routes when dualstack is enabled,
// and deletes all dualstack routes when dualstack is not enabled.
func (d *delayedRouteUpdater) cleanupOutdatedRoutes(existingRoutes []network.Route) (routes []network.Route, changed bool) {
	for i := len(existingRoutes) - 1; i >= 0; i-- {
		existingRouteName := pointer.StringDeref(existingRoutes[i].Name, "")
		split := strings.Split(existingRouteName, consts.RouteNameSeparator)

		klog.V(4).Infof("cleanupOutdatedRoutes: checking route %s", existingRouteName)

		// filter out unmanaged routes
		deleteRoute := false
		if d.az.nodeNames.Has(split[0]) {
			if d.az.ipv6DualStackEnabled && len(split) == 1 {
				klog.V(2).Infof("cleanupOutdatedRoutes: deleting outdated non-dualstack route %s", existingRouteName)
				deleteRoute = true
			} else if !d.az.ipv6DualStackEnabled && len(split) == 2 {
				klog.V(2).Infof("cleanupOutdatedRoutes: deleting outdated dualstack route %s", existingRouteName)
				deleteRoute = true
			}

			if deleteRoute {
				existingRoutes = append(existingRoutes[:i], existingRoutes[i+1:]...)
				changed = true
			}
		}
	}

	return existingRoutes, changed
}

func getAddRouteOperation(route network.Route) batchOperation {
	return &delayedRouteOperation{
		route:     route,
		operation: routeOperationAdd,
		result:    make(chan batchOperationResult),
	}
}

func getDeleteRouteOperation(route network.Route) batchOperation {
	return &delayedRouteOperation{
		route:     route,
		operation: routeOperationDelete,
		result:    make(chan batchOperationResult),
	}
}

func getUpdateRouteTableTagsOperation(tags map[string]*string) batchOperation {
	return &delayedRouteOperation{
		routeTableTags: tags,
		operation:      routeTableOperationUpdateTags,
		result:         make(chan batchOperationResult),
	}
}

// addOperation adds the routeOperation to delayedRouteUpdater and returns a delayedRouteOperation.
func (d *delayedRouteUpdater) addOperation(operation batchOperation) batchOperation {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.routesToUpdate = append(d.routesToUpdate, operation)
	return operation
}

func (d *delayedRouteUpdater) removeOperation(_ string) {}

// ListRoutes lists all managed routes that belong to the specified clusterName
// implements cloudprovider.Routes.ListRoutes
func (az *Cloud) ListRoutes(_ context.Context, clusterName string) ([]*cloudprovider.Route, error) {
	klog.V(10).Infof("ListRoutes: START clusterName=%q", clusterName)
	routeTable, existsRouteTable, err := az.getRouteTable(azcache.CacheReadTypeDefault)
	routes, err := processRoutes(az.ipv6DualStackEnabled, routeTable, existsRouteTable, err)
	if err != nil {
		return nil, err
	}

	// Compose routes for unmanaged routes so that node controller won't retry creating routes for them.
	unmanagedNodes, err := az.GetUnmanagedNodes()
	if err != nil {
		return nil, err
	}
	az.routeCIDRsLock.Lock()
	defer az.routeCIDRsLock.Unlock()
	for _, nodeName := range unmanagedNodes.UnsortedList() {
		if cidr, ok := az.routeCIDRs[nodeName]; ok {
			routes = append(routes, &cloudprovider.Route{
				Name:            nodeName,
				TargetNode:      MapRouteNameToNodeName(az.ipv6DualStackEnabled, nodeName),
				DestinationCIDR: cidr,
			})
		}
	}

	// ensure the route table is tagged as configured
	tags, changed := az.ensureRouteTableTagged(&routeTable)
	if changed {
		klog.V(2).Infof("ListRoutes: updating tags on route table %s", pointer.StringDeref(routeTable.Name, ""))
		op := az.routeUpdater.addOperation(getUpdateRouteTableTagsOperation(tags))

		// Wait for operation complete.
		err = op.wait().err
		if err != nil {
			klog.Errorf("ListRoutes: failed to update route table tags with error: %v", err)
			return nil, err
		}
	}

	return routes, nil
}

// Injectable for testing
func processRoutes(ipv6DualStackEnabled bool, routeTable network.RouteTable, exists bool, err error) ([]*cloudprovider.Route, error) {
	if err != nil {
		return nil, err
	}
	if !exists {
		return []*cloudprovider.Route{}, nil
	}

	var kubeRoutes []*cloudprovider.Route
	if routeTable.RouteTablePropertiesFormat != nil && routeTable.Routes != nil {
		kubeRoutes = make([]*cloudprovider.Route, len(*routeTable.Routes))
		for i, route := range *routeTable.Routes {
			instance := MapRouteNameToNodeName(ipv6DualStackEnabled, *route.Name)
			cidr := *route.AddressPrefix
			klog.V(10).Infof("ListRoutes: * instance=%q, cidr=%q", instance, cidr)

			kubeRoutes[i] = &cloudprovider.Route{
				Name:            *route.Name,
				TargetNode:      instance,
				DestinationCIDR: cidr,
			}
		}
	}

	klog.V(10).Info("ListRoutes: FINISH")
	return kubeRoutes, nil
}

func (az *Cloud) createRouteTable() error {
	routeTable := network.RouteTable{
		Name:                       pointer.String(az.RouteTableName),
		Location:                   pointer.String(az.Location),
		RouteTablePropertiesFormat: &network.RouteTablePropertiesFormat{},
	}

	klog.V(3).Infof("createRouteTableIfNotExists: creating routetable. routeTableName=%q", az.RouteTableName)
	err := az.CreateOrUpdateRouteTable(routeTable)
	if err != nil {
		return err
	}

	// Invalidate the cache right after updating
	_ = az.rtCache.Delete(az.RouteTableName)
	return nil
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
// implements cloudprovider.Routes.CreateRoute
func (az *Cloud) CreateRoute(_ context.Context, clusterName string, _ string, kubeRoute *cloudprovider.Route) error {
	mc := metrics.NewMetricContext("routes", "create_route", az.ResourceGroup, az.getNetworkResourceSubscriptionID(), string(kubeRoute.TargetNode))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	// Returns  for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	var targetIP string
	nodeName := string(kubeRoute.TargetNode)
	unmanaged, err := az.IsNodeUnmanaged(nodeName)
	if err != nil {
		return err
	}
	if unmanaged {
		klog.V(2).Infof("CreateRoute: omitting unmanaged node %q", kubeRoute.TargetNode)
		az.routeCIDRsLock.Lock()
		defer az.routeCIDRsLock.Unlock()
		az.routeCIDRs[nodeName] = kubeRoute.DestinationCIDR
		return nil
	}

	CIDRv6 := utilnet.IsIPv6CIDRString(kubeRoute.DestinationCIDR)
	// if single stack IPv4 then get the IP for the primary ip config
	// single stack IPv6 is supported on dual stack host. So the IPv6 IP is secondary IP for both single stack IPv6 and dual stack
	// Get all private IPs for the machine and find the first one that matches the IPv6 family
	if !az.ipv6DualStackEnabled && !CIDRv6 {
		targetIP, _, err = az.getIPForMachine(kubeRoute.TargetNode)
		if err != nil {
			return err
		}
	} else {
		// for dual stack and single stack IPv6 we need to select
		// a private ip that matches family of the cidr
		klog.V(4).Infof("CreateRoute: create route instance=%q cidr=%q is in dual stack mode", kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
		nodePrivateIPs, err := az.getPrivateIPsForMachine(kubeRoute.TargetNode)
		if nil != err {
			klog.V(3).Infof("CreateRoute: create route: failed(GetPrivateIPsByNodeName) instance=%q cidr=%q with error=%v", kubeRoute.TargetNode, kubeRoute.DestinationCIDR, err)
			return err
		}

		targetIP, err = findFirstIPByFamily(nodePrivateIPs, CIDRv6)
		if nil != err {
			klog.V(3).Infof("CreateRoute: create route: failed(findFirstIpByFamily) instance=%q cidr=%q with error=%v", kubeRoute.TargetNode, kubeRoute.DestinationCIDR, err)
			return err
		}
	}
	routeName := mapNodeNameToRouteName(az.ipv6DualStackEnabled, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	route := network.Route{
		Name: pointer.String(routeName),
		RoutePropertiesFormat: &network.RoutePropertiesFormat{
			AddressPrefix:    pointer.String(kubeRoute.DestinationCIDR),
			NextHopType:      network.RouteNextHopTypeVirtualAppliance,
			NextHopIPAddress: pointer.String(targetIP),
		},
	}

	klog.V(2).Infof("CreateRoute: creating route for clusterName=%q instance=%q cidr=%q", clusterName, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	op := az.routeUpdater.addOperation(getAddRouteOperation(route))

	// Wait for operation complete.
	err = op.wait().err
	if err != nil {
		klog.Errorf("CreateRoute failed for node %q with error: %v", kubeRoute.TargetNode, err)
		return err
	}

	klog.V(2).Infof("CreateRoute: route created. clusterName=%q instance=%q cidr=%q", clusterName, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	isOperationSucceeded = true

	return nil
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
// implements cloudprovider.Routes.DeleteRoute
func (az *Cloud) DeleteRoute(_ context.Context, clusterName string, kubeRoute *cloudprovider.Route) error {
	mc := metrics.NewMetricContext("routes", "delete_route", az.ResourceGroup, az.getNetworkResourceSubscriptionID(), string(kubeRoute.TargetNode))
	isOperationSucceeded := false
	defer func() {
		mc.ObserveOperationWithResult(isOperationSucceeded)
	}()

	// Returns  for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	nodeName := string(kubeRoute.TargetNode)
	unmanaged, err := az.IsNodeUnmanaged(nodeName)
	if err != nil {
		return err
	}
	if unmanaged {
		klog.V(2).Infof("DeleteRoute: omitting unmanaged node %q", kubeRoute.TargetNode)
		az.routeCIDRsLock.Lock()
		defer az.routeCIDRsLock.Unlock()
		delete(az.routeCIDRs, nodeName)
		return nil
	}

	routeName := mapNodeNameToRouteName(az.ipv6DualStackEnabled, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	klog.V(2).Infof("DeleteRoute: deleting route. clusterName=%q instance=%q cidr=%q routeName=%q", clusterName, kubeRoute.TargetNode, kubeRoute.DestinationCIDR, routeName)
	route := network.Route{
		Name:                  pointer.String(routeName),
		RoutePropertiesFormat: &network.RoutePropertiesFormat{},
	}
	op := az.routeUpdater.addOperation(getDeleteRouteOperation(route))

	// Wait for operation complete.
	err = op.wait().err
	if err != nil {
		klog.Errorf("DeleteRoute failed for node %q with error: %v", kubeRoute.TargetNode, err)
		return err
	}

	// Remove outdated ipv4 routes as well
	if az.ipv6DualStackEnabled {
		routeNameWithoutIPV6Suffix := strings.Split(routeName, consts.RouteNameSeparator)[0]
		klog.V(2).Infof("DeleteRoute: deleting route. clusterName=%q instance=%q cidr=%q routeName=%q", clusterName, kubeRoute.TargetNode, kubeRoute.DestinationCIDR, routeNameWithoutIPV6Suffix)
		route := network.Route{
			Name:                  pointer.String(routeNameWithoutIPV6Suffix),
			RoutePropertiesFormat: &network.RoutePropertiesFormat{},
		}
		op := az.routeUpdater.addOperation(getDeleteRouteOperation(route))

		// Wait for operation complete.
		err = op.wait().err
		if err != nil {
			klog.Errorf("DeleteRoute failed for node %q with error: %v", kubeRoute.TargetNode, err)
			return err
		}
	}

	klog.V(2).Infof("DeleteRoute: route deleted. clusterName=%q instance=%q cidr=%q", clusterName, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	isOperationSucceeded = true

	return nil
}

// This must be kept in sync with MapRouteNameToNodeName.
// These two functions enable stashing the instance name in the route
// and then retrieving it later when listing. This is needed because
// Azure does not let you put tags/descriptions on the Route itself.
func mapNodeNameToRouteName(ipv6DualStackEnabled bool, nodeName types.NodeName, cidr string) string {
	if !ipv6DualStackEnabled {
		return string(nodeName)
	}
	return fmt.Sprintf(consts.RouteNameFmt, nodeName, cidrtoRfc1035(cidr))
}

// MapRouteNameToNodeName is used with mapNodeNameToRouteName.
// See comment on mapNodeNameToRouteName for detailed usage.
func MapRouteNameToNodeName(ipv6DualStackEnabled bool, routeName string) types.NodeName {
	if !ipv6DualStackEnabled {
		return types.NodeName(routeName)
	}
	parts := strings.Split(routeName, consts.RouteNameSeparator)
	nodeName := parts[0]
	return types.NodeName(nodeName)

}

// given a list of ips, return the first one
// that matches the family requested
// error if no match, or failure to parse
// any of the ips
func findFirstIPByFamily(ips []string, v6 bool) (string, error) {
	for _, ip := range ips {
		bIPv6 := utilnet.IsIPv6String(ip)
		if v6 == bIPv6 {
			return ip, nil
		}
	}
	return "", fmt.Errorf("no match found matching the ipfamily requested")
}

// strips : . /
func cidrtoRfc1035(cidr string) string {
	cidr = strings.ReplaceAll(cidr, ":", "")
	cidr = strings.ReplaceAll(cidr, ".", "")
	cidr = strings.ReplaceAll(cidr, "/", "")
	return cidr
}

// ensureRouteTableTagged ensures the route table is tagged as configured
func (az *Cloud) ensureRouteTableTagged(rt *network.RouteTable) (map[string]*string, bool) {
	if !strings.EqualFold(az.RouteTableResourceGroup, az.ResourceGroup) {
		return nil, false
	}

	if az.Tags == "" && (az.TagsMap == nil || len(az.TagsMap) == 0) {
		return nil, false
	}
	tags := parseTags(az.Tags, az.TagsMap)
	if rt.Tags == nil {
		rt.Tags = make(map[string]*string)
	}

	tags, changed := az.reconcileTags(rt.Tags, tags)
	rt.Tags = tags

	return rt.Tags, changed
}
