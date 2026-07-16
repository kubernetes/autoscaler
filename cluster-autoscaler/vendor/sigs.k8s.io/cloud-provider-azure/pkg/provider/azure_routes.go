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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	utilnet "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
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
	route          *armnetwork.Route
	routeTableTags map[string]*string
	operation      routeOperation
	result         chan batchOperationResult
	nodeName       string
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
	logger := log.FromContextOrBackground(ctx).WithName("delayedRouteUpdater.run")
	logger.Info("Started")
	err := wait.PollUntilContextCancel(ctx, d.interval, true, func(ctx context.Context) (bool, error) {
		d.updateRoutes(ctx)
		return false, nil
	})
	logger.Error(err, "stopped")
}

// updateRoutes invokes route table client to update all routes.
func (d *delayedRouteUpdater) updateRoutes(ctx context.Context) {
	logger := log.FromContextOrBackground(ctx).WithName("updateRoutes")
	d.lock.Lock()
	defer d.lock.Unlock()

	// No need to do any updating.
	if len(d.routesToUpdate) == 0 {
		logger.V(4).Info("nothing to update, returning")
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
		routeTable *armnetwork.RouteTable
	)
	routeTable, err = d.az.routeTableRepo.Get(ctx, d.az.RouteTableName, azcache.CacheReadTypeDefault)
	if err != nil {
		logger.Error(err, "getRouteTable() failed")
		return
	}

	// create route table if it doesn't exists yet.
	if routeTable == nil {
		err = d.az.createRouteTable(ctx)
		if err != nil {
			logger.Error(err, "createRouteTable() failed")
			return
		}

		routeTable, err = d.az.routeTableRepo.Get(ctx, d.az.RouteTableName, azcache.CacheReadTypeDefault)
		if err != nil {
			logger.Error(err, "getRouteTable() failed")
			return
		}
	}

	// reconcile routes.
	dirty, onlyUpdateTags := false, true
	var routes []*armnetwork.Route
	if routeTable.Properties != nil {
		routes = routeTable.Properties.Routes
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
			if strings.EqualFold(ptr.Deref(existingRoute.Name, ""), ptr.Deref(rt.route.Name, "")) {
				// delete the name-matched routes here (missing routes would be added later if the operation is add).
				routes = append(routes[:i], routes[i+1:]...)
				if existingRoute.Properties != nil &&
					rt.route.Properties != nil &&
					strings.EqualFold(ptr.Deref(existingRoute.Properties.AddressPrefix, ""), ptr.Deref(rt.route.Properties.AddressPrefix, "")) &&
					strings.EqualFold(ptr.Deref(existingRoute.Properties.NextHopIPAddress, ""), ptr.Deref(rt.route.Properties.NextHopIPAddress, "")) {
					routeMatch = true
				}
				if rt.operation == routeOperationDelete {
					dirty = true
				}
				break
			}
		}
		// After removing the matched routes (if any), loop again to remove the outdated routes,
		// whose name and IP addresses may not match but target the same node.
		for i := len(routes) - 1; i >= 0; i-- {
			existingRoute := routes[i]
			// remove all routes that target the node when the operation is delete
			if strings.HasPrefix(ptr.Deref(existingRoute.Name, ""), ptr.Deref(rt.route.Name, "")) {
				if rt.operation == routeOperationDelete {
					routes = append(routes[:i], routes[i+1:]...)
					dirty = true
					logger.V(2).Info("found outdated route targeting node, removing it", "route", ptr.Deref(rt.route.Name, ""), "node", rt.nodeName)
				}
			}
		}
		if rt.operation == routeOperationDelete && !dirty {
			logger.Info("route to be deleted does not match any of the existing route", "route", ptr.Deref(rt.route.Name, ""))
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
			logger.V(2).Info("updating routes")
			routeTable.Properties.Routes = routes
		}
		_, err := d.az.routeTableRepo.CreateOrUpdate(ctx, *routeTable)
		if err != nil {
			logger.Error(err, "CreateOrUpdateRouteTable() failed")
			return
		}

		// wait a while for route updates to take effect.
		time.Sleep(time.Duration(d.az.RouteUpdateWaitingInSeconds) * time.Second)
	}
}

// cleanupOutdatedRoutes deletes all non-dualstack routes when dualstack is enabled,
// and deletes all dualstack routes when dualstack is not enabled.
func (d *delayedRouteUpdater) cleanupOutdatedRoutes(existingRoutes []*armnetwork.Route) (routes []*armnetwork.Route, changed bool) {
	logger := log.Background().WithName("cleanupOutdatedRoutes")
	for i := len(existingRoutes) - 1; i >= 0; i-- {
		existingRouteName := ptr.Deref(existingRoutes[i].Name, "")
		split := strings.Split(existingRouteName, consts.RouteNameSeparator)

		logger.V(4).Info("checking route", "route", existingRouteName)

		// filter out unmanaged routes
		deleteRoute := false
		if d.az.nodeNames.Has(split[0]) {
			if d.az.ipv6DualStackEnabled && len(split) == 1 {
				logger.V(2).Info("deleting outdated non-dualstack route", "route", existingRouteName)
				deleteRoute = true
			} else if !d.az.ipv6DualStackEnabled && len(split) == 2 {
				logger.V(2).Info("deleting outdated dualstack route", "route", existingRouteName)
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

func getAddRouteOperation(route *armnetwork.Route, nodeName string) batchOperation {
	return &delayedRouteOperation{
		route:     route,
		nodeName:  nodeName,
		operation: routeOperationAdd,
		result:    make(chan batchOperationResult),
	}
}

func getDeleteRouteOperation(route *armnetwork.Route, nodeName string) batchOperation {
	return &delayedRouteOperation{
		route:     route,
		nodeName:  nodeName,
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
func (az *Cloud) ListRoutes(ctx context.Context, clusterName string) ([]*cloudprovider.Route, error) {
	logger := log.FromContextOrBackground(ctx).WithName("ListRoutes")
	logger.V(10).Info("START", "clusterName", clusterName)
	routeTable, err := az.routeTableRepo.Get(ctx, az.RouteTableName, azcache.CacheReadTypeDefault)
	routes, err := processRoutes(az.ipv6DualStackEnabled, routeTable, err)
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
	tags, changed := az.ensureRouteTableTagged(routeTable)
	if changed {
		logger.V(2).Info("updating tags on route table", "routeTableName", ptr.Deref(routeTable.Name, ""))
		op := az.routeUpdater.addOperation(getUpdateRouteTableTagsOperation(tags))

		// Wait for operation complete.
		err = op.wait().err
		if err != nil {
			logger.Error(err, "failed to update route table tags")
			return nil, err
		}
	}

	return routes, nil
}

// Injectable for testing
func processRoutes(ipv6DualStackEnabled bool, routeTable *armnetwork.RouteTable, err error) ([]*cloudprovider.Route, error) {
	logger := log.Background().WithName("processRoutes")
	if err != nil {
		return nil, err
	}
	if routeTable == nil {
		return []*cloudprovider.Route{}, nil
	}

	var kubeRoutes []*cloudprovider.Route
	if routeTable.Properties != nil {
		kubeRoutes = make([]*cloudprovider.Route, len(routeTable.Properties.Routes))
		for i, route := range routeTable.Properties.Routes {
			instance := MapRouteNameToNodeName(ipv6DualStackEnabled, *route.Name)
			cidr := *route.Properties.AddressPrefix
			logger.V(10).Info("Got route", "instance", instance, "cidr", cidr)

			kubeRoutes[i] = &cloudprovider.Route{
				Name:            *route.Name,
				TargetNode:      instance,
				DestinationCIDR: cidr,
			}
		}
	}

	logger.V(10).Info("FINISH")
	return kubeRoutes, nil
}

func (az *Cloud) createRouteTable(ctx context.Context) error {
	logger := log.FromContextOrBackground(ctx).WithName("createRouteTable")
	routeTable := armnetwork.RouteTable{
		Name:       ptr.To(az.RouteTableName),
		Location:   ptr.To(az.Location),
		Properties: &armnetwork.RouteTablePropertiesFormat{},
	}

	logger.V(3).Info("creating routetable if not exists", "routeTableName", az.RouteTableName)
	_, err := az.routeTableRepo.CreateOrUpdate(ctx, routeTable)
	return err
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
// implements cloudprovider.Routes.CreateRoute
func (az *Cloud) CreateRoute(ctx context.Context, clusterName string, _ string, kubeRoute *cloudprovider.Route) error {
	logger := log.FromContextOrBackground(ctx).WithName("CreateRoute")
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
		logger.V(2).Info("omitting unmanaged node", "node", kubeRoute.TargetNode)
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
		targetIP, _, err = az.getIPForMachine(ctx, kubeRoute.TargetNode)
		if err != nil {
			return err
		}
	} else {
		// for dual stack and single stack IPv6 we need to select
		// a private ip that matches family of the cidr
		logger.V(4).Info("create route instance in dual stack mode", "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
		nodePrivateIPs, err := az.getPrivateIPsForMachine(ctx, kubeRoute.TargetNode)
		if nil != err {
			logger.V(3).Error(err, "failed(GetPrivateIPsByNodeName)", "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
			return err
		}

		targetIP, err = findFirstIPByFamily(nodePrivateIPs, CIDRv6)
		if nil != err {
			logger.V(3).Error(err, "failed(findFirstIpByFamily)", "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
			return err
		}
	}
	routeName := mapNodeNameToRouteName(az.ipv6DualStackEnabled, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	route := &armnetwork.Route{
		Name: ptr.To(routeName),
		Properties: &armnetwork.RoutePropertiesFormat{
			AddressPrefix:    ptr.To(kubeRoute.DestinationCIDR),
			NextHopType:      ptr.To(armnetwork.RouteNextHopTypeVirtualAppliance),
			NextHopIPAddress: ptr.To(targetIP),
		},
	}

	logger.V(2).Info("creating route", "clusterName", clusterName, "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
	op := az.routeUpdater.addOperation(getAddRouteOperation(route, string(kubeRoute.TargetNode)))

	// Wait for operation complete.
	err = op.wait().err
	if err != nil {
		logger.Error(err, "failed for node", "node", kubeRoute.TargetNode)
		return err
	}

	logger.V(2).Info("route created", "clusterName", clusterName, "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
	isOperationSucceeded = true

	return nil
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
// implements cloudprovider.Routes.DeleteRoute
func (az *Cloud) DeleteRoute(ctx context.Context, clusterName string, kubeRoute *cloudprovider.Route) error {
	logger := log.FromContextOrBackground(ctx).WithName("DeleteRoute")
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
		logger.V(2).Info("omitting unmanaged node", "node", kubeRoute.TargetNode)
		az.routeCIDRsLock.Lock()
		defer az.routeCIDRsLock.Unlock()
		delete(az.routeCIDRs, nodeName)
		return nil
	}

	routeName := mapNodeNameToRouteName(az.ipv6DualStackEnabled, kubeRoute.TargetNode, kubeRoute.DestinationCIDR)
	logger.V(2).Info("deleting route", "clusterName", clusterName, "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR, "routeName", routeName)
	route := &armnetwork.Route{
		Name:       ptr.To(routeName),
		Properties: &armnetwork.RoutePropertiesFormat{},
	}
	op := az.routeUpdater.addOperation(getDeleteRouteOperation(route, string(kubeRoute.TargetNode)))

	// Wait for operation complete.
	err = op.wait().err
	if err != nil {
		logger.Error(err, "failed for node", "node", kubeRoute.TargetNode)
		return err
	}

	// Remove outdated ipv4 routes as well
	if az.ipv6DualStackEnabled {
		routeNameWithoutIPV6Suffix := strings.Split(routeName, consts.RouteNameSeparator)[0]
		logger.V(2).Info("deleting route", "clusterName", clusterName, "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR, "routeName", routeNameWithoutIPV6Suffix)
		route := &armnetwork.Route{
			Name:       ptr.To(routeNameWithoutIPV6Suffix),
			Properties: &armnetwork.RoutePropertiesFormat{},
		}
		op := az.routeUpdater.addOperation(getDeleteRouteOperation(route, string(kubeRoute.TargetNode)))

		// Wait for operation complete.
		err = op.wait().err
		if err != nil {
			logger.Error(err, "failed for node", "node", kubeRoute.TargetNode)
			return err
		}
	}

	logger.V(2).Info("route deleted", "clusterName", clusterName, "instance", kubeRoute.TargetNode, "cidr", kubeRoute.DestinationCIDR)
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
func (az *Cloud) ensureRouteTableTagged(rt *armnetwork.RouteTable) (map[string]*string, bool) {
	if !strings.EqualFold(az.RouteTableResourceGroup, az.ResourceGroup) {
		return nil, false
	}

	if az.Tags == "" && (len(az.TagsMap) == 0) {
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
