// Copyright 2020 Brightbox Systems Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8ssdk

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	brightbox "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox/status"
	klog "k8s.io/klog/v2"
)

// GetServer retrieves a Brightbox Cloud Server
func (c *Cloud) GetServer(ctx context.Context, id string, notFoundError error) (*brightbox.Server, error) {
	klog.V(4).Infof("getServer (%q)", id)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	srv, err := client.Server(id)
	if err != nil {
		if isNotFound(err) {
			return nil, notFoundError
		}
		return nil, err
	}
	return srv, nil
}

func isNotFound(e error) bool {
	switch v := e.(type) {
	case brightbox.ApiError:
		return v.StatusCode == http.StatusNotFound
	default:
		return false
	}
}

// CreateServer creates a Brightbox Cloud Server
func (c *Cloud) CreateServer(newDetails *brightbox.ServerOptions) (*brightbox.Server, error) {
	klog.V(4).Infof("CreateServer (%q)", *newDetails.Name)
	klog.V(6).Infof("%+v", newDetails)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CreateServer(newDetails)
}

func isAlive(lb *brightbox.LoadBalancer) bool {
	return lb.Status == status.Active || lb.Status == status.Creating
}
func trimmed(name string) string {
	return strings.TrimSpace(
		strings.TrimSuffix(
			strings.TrimSpace(name),
			"#type:container",
		),
	)
}

// GetLoadBalancerByName finds a Load Balancer from its name
func (c *Cloud) GetLoadBalancerByName(name string) (*brightbox.LoadBalancer, error) {
	klog.V(4).Infof("GetLoadBalancerByName (%q)", name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	lbList, err := client.LoadBalancers()
	if err != nil {
		return nil, err
	}
	for i := range lbList {
		if isAlive(&lbList[i]) && trimmed(name) == trimmed(lbList[i].Name) {
			return &lbList[i], nil
		}
	}
	return nil, nil
}

// GetLoadBalancerByID finds a Load Balancer from its ID
func (c *Cloud) GetLoadBalancerByID(id string) (*brightbox.LoadBalancer, error) {
	klog.V(4).Infof("GetLoadBalancerById (%q)", id)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.LoadBalancer(id)
}

// CreateLoadBalancer creates a LoadBalancer
func (c *Cloud) CreateLoadBalancer(newDetails *brightbox.LoadBalancerOptions) (*brightbox.LoadBalancer, error) {
	klog.V(4).Infof("CreateLoadBalancer (%q)", *newDetails.Name)
	klog.V(6).Infof("%+v", newDetails)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CreateLoadBalancer(newDetails)
}

// UpdateLoadBalancer updates a LoadBalancer
func (c *Cloud) UpdateLoadBalancer(newDetails *brightbox.LoadBalancerOptions) (*brightbox.LoadBalancer, error) {
	klog.V(4).Infof("UpdateLoadBalancer (%q, %q)", newDetails.Id, *newDetails.Name)
	klog.V(6).Infof("%+v", newDetails)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.UpdateLoadBalancer(newDetails)
}

// GetFirewallPolicyByName get a FirewallPolicy from its name
func (c *Cloud) GetFirewallPolicyByName(name string) (*brightbox.FirewallPolicy, error) {
	klog.V(4).Infof("getFirewallPolicyByName (%q)", name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	firewallPolicyList, err := client.FirewallPolicies()
	if err != nil {
		return nil, err
	}
	var result *brightbox.FirewallPolicy
	for i := range firewallPolicyList {
		if name == firewallPolicyList[i].Name {
			result = &firewallPolicyList[i]
			break
		}
	}
	return result, nil
}

// GetServerTypes obtains the list of Server Types on the account
func (c *Cloud) GetServerTypes() ([]brightbox.ServerType, error) {
	klog.V(4).Info("GetServerTypes")
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ServerTypes()
}

// GetServerType fetches a Server Type from its ID
func (c *Cloud) GetServerType(identifier string) (*brightbox.ServerType, error) {
	klog.V(4).Infof("GetServerType %q", identifier)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ServerType(identifier)
}

// GetImageByName obtains the most recent available image that matches the supplied pattern
func (c *Cloud) GetImageByName(name string) (*brightbox.Image, error) {
	klog.V(4).Infof("GetImageByName %q", name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	klog.V(6).Info("GetImageByName compiling regexp")
	nameRe, err := regexp.Compile(name)
	if err != nil {
		return nil, err
	}
	klog.V(6).Info("GetImageByName retrieving images")
	images, err := client.Images()
	if err != nil {
		return nil, err
	}
	klog.V(6).Info("GetImageByName filtering images")
	filteredImages := filter(
		images,
		func(i brightbox.Image) bool {
			return i.Official &&
				i.Status == status.Available &&
				nameRe.MatchString(i.Name)
		},
	)
	klog.V(6).Infof("GetImageByName images selected (%+v)", filteredImages)
	return mostRecent(filteredImages), nil
}

// GetConfigMaps obtains the list of Config Maps on the account
func (c *Cloud) GetConfigMaps() ([]brightbox.ConfigMap, error) {
	klog.V(4).Info("GetConfigMaps")
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ConfigMaps()
}

// GetConfigMap fetches a Config Map from its ID
func (c *Cloud) GetConfigMap(identifier string) (*brightbox.ConfigMap, error) {
	klog.V(4).Infof("GetConfigMap %q", identifier)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ConfigMap(identifier)
}

// GetServerGroups obtains the list of Server Groups on the account
func (c *Cloud) GetServerGroups() ([]brightbox.ServerGroup, error) {
	klog.V(4).Info("GetServerGroups")
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ServerGroups()
}

// GetServerGroup fetches a Server Group from its ID
func (c *Cloud) GetServerGroup(identifier string) (*brightbox.ServerGroup, error) {
	klog.V(4).Infof("GetServerGroup %q", identifier)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.ServerGroup(identifier)
}

// GetServerGroupByName fetches a Server Group from its name
func (c *Cloud) GetServerGroupByName(name string) (*brightbox.ServerGroup, error) {
	klog.V(4).Infof("GetServerGroupByName (%q)", name)
	serverGroupList, err := c.GetServerGroups()
	if err != nil {
		return nil, err
	}
	var result *brightbox.ServerGroup
	for i := range serverGroupList {
		if name == serverGroupList[i].Name {
			result = &serverGroupList[i]
			break
		}
	}
	return result, nil
}

// CreateServerGroup creates a Server Group
func (c *Cloud) CreateServerGroup(name string) (*brightbox.ServerGroup, error) {
	klog.V(4).Infof("CreateServerGroup (%q)", name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CreateServerGroup(&brightbox.ServerGroupOptions{Name: &name})
}

// CreateFirewallPolicy creates a Firewall Policy
func (c *Cloud) CreateFirewallPolicy(group *brightbox.ServerGroup) (*brightbox.FirewallPolicy, error) {
	klog.V(4).Infof("createFirewallPolicy (%q)", group.Name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CreateFirewallPolicy(&brightbox.FirewallPolicyOptions{Name: &group.Name, ServerGroup: &group.Id})
}

// CreateFirewallRule creates a Firewall Rule
func (c *Cloud) CreateFirewallRule(newDetails *brightbox.FirewallRuleOptions) (*brightbox.FirewallRule, error) {
	klog.V(4).Infof("createFirewallRule (%q)", *newDetails.Description)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CreateFirewallRule(newDetails)
}

// UpdateFirewallRule updates a Firewall Rule
func (c *Cloud) UpdateFirewallRule(newDetails *brightbox.FirewallRuleOptions) (*brightbox.FirewallRule, error) {
	klog.V(4).Infof("updateFirewallRule (%q, %q)", newDetails.Id, *newDetails.Description)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.UpdateFirewallRule(newDetails)
}

// EnsureMappedCloudIP checks to make sure the Cloud IP is mapped to the Load Balancer.
// This function is idempotent.
func (c *Cloud) EnsureMappedCloudIP(lb *brightbox.LoadBalancer, cip *brightbox.CloudIP) error {
	klog.V(4).Infof("EnsureMappedCloudIP (%q, %q)", lb.Id, cip.Id)
	if alreadyMapped(cip, lb.Id) {
		return nil
	} else if cip.Status == status.Mapped {
		return fmt.Errorf("Unexplained mapping of %q (%q)", cip.Id, cip.PublicIP)
	}
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.MapCloudIP(cip.Id, lb.Id)
}

func alreadyMapped(cip *brightbox.CloudIP, loadBalancerID string) bool {
	return cip.LoadBalancer != nil && cip.LoadBalancer.Id == loadBalancerID
}

// AllocateCloudIP allocates a new Cloud IP and gives it the name specified
func (c *Cloud) AllocateCloudIP(name string) (*brightbox.CloudIP, error) {
	klog.V(4).Infof("AllocateCloudIP %q", name)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	opts := &brightbox.CloudIPOptions{
		Name: &name,
	}
	return client.CreateCloudIP(opts)
}

// GetCloudIPs obtains the list of allocated Cloud IPs
func (c *Cloud) GetCloudIPs() ([]brightbox.CloudIP, error) {
	klog.V(4).Infof("GetCloudIPs")
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CloudIPs()
}

// Get a cloudIp by id
func (c *Cloud) getCloudIP(id string) (*brightbox.CloudIP, error) {
	klog.V(4).Infof("getCloudIP (%q)", id)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	return client.CloudIP(id)
}

// Destroy things

// DestroyLoadBalancer removes a Load Balancer
func (c *Cloud) DestroyLoadBalancer(id string) error {
	klog.V(4).Infof("DestroyLoadBalancer %q", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.DestroyLoadBalancer(id)
}

// DestroyServer removes a Server
func (c *Cloud) DestroyServer(id string) error {
	klog.V(4).Infof("DestroyServer %q", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.DestroyServer(id)
}

// DestroyServerGroup removes a Server Group
func (c *Cloud) DestroyServerGroup(id string) error {
	klog.V(4).Infof("DestroyServerGroup %q", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.DestroyServerGroup(id)
}

// DestroyFirewallPolicy removes a Firewall Policy
func (c *Cloud) DestroyFirewallPolicy(id string) error {
	klog.V(4).Infof("DestroyFirewallPolicy %q", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.DestroyFirewallPolicy(id)
}

// DestroyCloudIP removes a Cloud IP allocation
func (c *Cloud) DestroyCloudIP(id string) error {
	klog.V(4).Infof("DestroyCloudIP (%q)", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.DestroyCloudIP(id)
}

// unmapCloudIP removes a mapping to a Cloud IP
func (c *Cloud) unmapCloudIP(id string) error {
	klog.V(4).Infof("unmapCloudIP (%q)", id)
	client, err := c.CloudClient()
	if err != nil {
		return err
	}
	return client.UnMapCloudIP(id)
}

// DestroyCloudIPs matching 'name' from a supplied list of cloudIPs
func (c *Cloud) DestroyCloudIPs(cloudIPList []brightbox.CloudIP, name string) error {
	klog.V(4).Infof("DestroyCloudIPs (%q)", name)
	for _, v := range cloudIPList {
		if v.Name == name {
			if err := c.DestroyCloudIP(v.Id); err != nil {
				klog.V(4).Infof("Error destroying CloudIP %q", v.Id)
				return err
			}
		}
	}
	return nil
}

// EnsureOldCloudIPsDeposed unmaps any CloudIPs mapped to the loadbalancer
// that isn't the current cloudip and matches 'name'
func (c *Cloud) EnsureOldCloudIPsDeposed(cloudIPList []brightbox.CloudIP, currentIPID string, name string) error {
	klog.V(4).Infof("EnsureOldCloudIPsDeposed (%q, %q)", currentIPID, name)
	for _, v := range cloudIPList {
		if v.Name == name && v.Id != currentIPID {
			if err := c.unmapCloudIP(v.Id); err != nil {
				klog.V(4).Infof("Error unmapping CloudIP %q", v.Id)
				return err
			}
		}
	}
	return nil
}

func mapServersToServerIds(servers []brightbox.Server) []string {
	result := make([]string, len(servers))
	for i := range servers {
		result[i] = servers[i].Id
	}
	return result
}

// SyncServerGroup ensures a Brightbox Server Group contains the supplied
// list of Servers and nothing else
func (c *Cloud) SyncServerGroup(group *brightbox.ServerGroup, newServerIds []string) (*brightbox.ServerGroup, error) {
	oldServerIds := mapServersToServerIds(group.Servers)
	klog.V(4).Infof("SyncServerGroup (%v, %v, %v)", group.Id, oldServerIds, newServerIds)
	client, err := c.CloudClient()
	if err != nil {
		return nil, err
	}
	serverIdsToInsert, serverIdsToDelete := getSyncLists(oldServerIds, newServerIds)
	result := group
	if len(serverIdsToInsert) > 0 {
		klog.V(4).Infof("Adding Servers %v", serverIdsToInsert)
		result, err = client.AddServersToServerGroup(group.Id, serverIdsToInsert)
	}
	if err == nil && len(serverIdsToDelete) > 0 {
		klog.V(4).Infof("Removing Servers %v", serverIdsToDelete)
		result, err = client.RemoveServersFromServerGroup(group.Id, serverIdsToDelete)
	}
	return result, err
}

// IsUpdateLoadBalancerRequired checks whether a set of LoadBalancerOptions
// warrants an API update call.
func IsUpdateLoadBalancerRequired(lb *brightbox.LoadBalancer, newDetails brightbox.LoadBalancerOptions) bool {
	klog.V(6).Infof("Update LoadBalancer Required (%v, %v)", *newDetails.Name, lb.Name)
	return (newDetails.Name != nil && *newDetails.Name != lb.Name) ||
		(newDetails.Healthcheck != nil && isUpdateLoadBalancerHealthcheckRequired(newDetails.Healthcheck, &lb.Healthcheck)) ||
		isUpdateLoadBalancerNodeRequired(newDetails.Nodes, lb.Nodes) ||
		isUpdateLoadBalancerListenerRequired(newDetails.Listeners, lb.Listeners) ||
		isUpdateLoadBalancerDomainsRequired(newDetails.Domains, lb.Acme)
}

func isUpdateLoadBalancerHealthcheckRequired(newHealthCheck *brightbox.LoadBalancerHealthcheck, oldHealthCheck *brightbox.LoadBalancerHealthcheck) bool {
	klog.V(6).Infof("Update LoadBalancer Healthcheck Required (%#v, %#v)", *newHealthCheck, *oldHealthCheck)
	return (newHealthCheck.Type != oldHealthCheck.Type) ||
		(newHealthCheck.Port != oldHealthCheck.Port) ||
		(newHealthCheck.Request != oldHealthCheck.Request)
}

func isUpdateLoadBalancerNodeRequired(a []brightbox.LoadBalancerNode, b []brightbox.Server) bool {
	klog.V(6).Infof("Update LoadBalancer Node Required (%v, %v)", a, b)
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return true
	}
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i].Node != b[i].Id {
			return true
		}
	}
	return false
}

func isUpdateLoadBalancerListenerRequired(a []brightbox.LoadBalancerListener, b []brightbox.LoadBalancerListener) bool {
	klog.V(6).Infof("Update LoadBalancer Listener Required (%v, %v)", a, b)
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return true
	}
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if (a[i].Protocol != b[i].Protocol) ||
			(a[i].In != b[i].In) ||
			(a[i].Out != b[i].Out) ||
			(a[i].Timeout != 0 && b[i].Timeout != 0 && a[i].Timeout != b[i].Timeout) ||
			(a[i].ProxyProtocol != b[i].ProxyProtocol) {
			return true
		}
	}
	return false
}

func isUpdateLoadBalancerDomainsRequired(a *[]string, acme *brightbox.LoadBalancerAcme) bool {
	klog.V(6).Infof("Update LoadBalancer Domains Required (%v)", a)
	if acme == nil {
		return a != nil
	}
	if a == nil {
		return false
	}
	b := make([]string, len(acme.Domains))
	for i, domain := range acme.Domains {
		b[i] = domain.Identifier
	}
	return !sameStringSlice(*a, b)
}

// ErrorIfNotErased returns an appropriate error if the Load Balancer has not been erased
func ErrorIfNotErased(lb *brightbox.LoadBalancer) error {
	switch {
	case lb == nil:
		return nil
	case lb.CloudIPs != nil && len(lb.CloudIPs) > 0:
		return fmt.Errorf("CloudIPs still mapped to load balancer %q", lb.Id)
	case !isAlive(lb):
		return nil
	}
	return fmt.Errorf("Unknown reason why %q has not deleted", lb.Id)
}

// ErrorIfNotComplete returns an appropriate error if the Load Balancer has not yet built
func ErrorIfNotComplete(lb *brightbox.LoadBalancer, cipID, name string) error {
	switch {
	case lb == nil:
		return fmt.Errorf("Load Balancer for %q is missing", name)
	case !isAlive(lb):
		return fmt.Errorf("Load Balancer %q still building", lb.Id)
	case !containsCIP(lb.CloudIPs, cipID):
		return fmt.Errorf("Mapping of CloudIP %q to %q not complete", cipID, lb.Id)
	}
	return ErrorIfAcmeNotComplete(lb.Acme)
}

// Look for a CIP Id in a list of cloudIPs
func containsCIP(cloudIPList []brightbox.CloudIP, cipID string) bool {
	for _, v := range cloudIPList {
		if v.Id == cipID {
			return true
		}
	}
	return false
}

// ErrorIfAcmeNotComplete returns an appropriate error if ACME has not yet validated
func ErrorIfAcmeNotComplete(acme *brightbox.LoadBalancerAcme) error {
	if acme != nil {
		for _, domain := range acme.Domains {
			if domain.Status != ValidAcmeDomainStatus {
				return fmt.Errorf("Domain %q has not yet been validated for SSL use (%q:%q)", domain.Identifier, domain.Status, domain.LastMessage)
			}
		}
	}
	return nil
}

// Returns the most recent item out of a slice of items with Dates
// or nil if there are no items
func mostRecent(items []brightbox.Image) *brightbox.Image {
	if len(items) == 0 {
		return nil
	}
	sortedItems := items
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Unix() > items[j].CreatedAt.Unix()
	})
	return &sortedItems[0]
}

// filter returns a new slice with all elements from the
// input elements for which the provided predicate function returns true.
func filter[T any](input []T, pred func(T) bool) (output []T) {
	for _, v := range input {
		if pred(v) {
			output = append(output, v)
		}
	}
	return output
}
