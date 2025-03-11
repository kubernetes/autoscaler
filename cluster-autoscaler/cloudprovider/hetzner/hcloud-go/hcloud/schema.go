package hcloud

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/schema"
)

// This file provides converter functions to convert models in the
// schema package to models in the hcloud package and vice versa.

var c converter

// ActionFromSchema converts a schema.Action to an Action.
func ActionFromSchema(s schema.Action) *Action {
	return c.ActionFromSchema(s)
}

// SchemaFromAction converts an Action to a schema.Action.
func SchemaFromAction(a *Action) schema.Action {
	return c.SchemaFromAction(a)
}

// ActionsFromSchema converts a slice of schema.Action to a slice of Action.
func ActionsFromSchema(s []schema.Action) []*Action {
	return c.ActionsFromSchema(s)
}

// SchemaFromActions converts a slice of Action to a slice of schema.Action.
func SchemaFromActions(a []*Action) []schema.Action {
	return c.SchemaFromActions(a)
}

// FloatingIPFromSchema converts a schema.FloatingIP to a FloatingIP.
func FloatingIPFromSchema(s schema.FloatingIP) *FloatingIP {
	return c.FloatingIPFromSchema(s)
}

// SchemaFromFloatingIP converts a FloatingIP to a schema.FloatingIP.
func SchemaFromFloatingIP(f *FloatingIP) schema.FloatingIP {
	return c.SchemaFromFloatingIP(f)
}

// PrimaryIPFromSchema converts a schema.PrimaryIP to a PrimaryIP.
func PrimaryIPFromSchema(s schema.PrimaryIP) *PrimaryIP {
	return c.PrimaryIPFromSchema(s)
}

// SchemaFromPrimaryIP converts a PrimaryIP to a schema.PrimaryIP.
func SchemaFromPrimaryIP(p *PrimaryIP) schema.PrimaryIP {
	return c.SchemaFromPrimaryIP(p)
}

// ISOFromSchema converts a schema.ISO to an ISO.
func ISOFromSchema(s schema.ISO) *ISO {
	return c.ISOFromSchema(s)
}

// SchemaFromISO converts an ISO to a schema.ISO.
func SchemaFromISO(i *ISO) schema.ISO {
	return c.SchemaFromISO(i)
}

// LocationFromSchema converts a schema.Location to a Location.
func LocationFromSchema(s schema.Location) *Location {
	return c.LocationFromSchema(s)
}

// SchemaFromLocation converts a Location to a schema.Location.
func SchemaFromLocation(l *Location) schema.Location {
	return c.SchemaFromLocation(l)
}

// DatacenterFromSchema converts a schema.Datacenter to a Datacenter.
func DatacenterFromSchema(s schema.Datacenter) *Datacenter {
	return c.DatacenterFromSchema(s)
}

// SchemaFromDatacenter converts a Datacenter to a schema.Datacenter.
func SchemaFromDatacenter(d *Datacenter) schema.Datacenter {
	return c.SchemaFromDatacenter(d)
}

// ServerFromSchema converts a schema.Server to a Server.
func ServerFromSchema(s schema.Server) *Server {
	return c.ServerFromSchema(s)
}

// SchemaFromServer converts a Server to a schema.Server.
func SchemaFromServer(s *Server) schema.Server {
	return c.SchemaFromServer(s)
}

// ServerPublicNetFromSchema converts a schema.ServerPublicNet to a ServerPublicNet.
func ServerPublicNetFromSchema(s schema.ServerPublicNet) ServerPublicNet {
	return c.ServerPublicNetFromSchema(s)
}

// SchemaFromServerPublicNet converts a ServerPublicNet to a schema.ServerPublicNet.
func SchemaFromServerPublicNet(s ServerPublicNet) schema.ServerPublicNet {
	return c.SchemaFromServerPublicNet(s)
}

// ServerPublicNetIPv4FromSchema converts a schema.ServerPublicNetIPv4 to
// a ServerPublicNetIPv4.
func ServerPublicNetIPv4FromSchema(s schema.ServerPublicNetIPv4) ServerPublicNetIPv4 {
	return c.ServerPublicNetIPv4FromSchema(s)
}

// SchemaFromServerPublicNetIPv4 converts a ServerPublicNetIPv4 to
// a schema.ServerPublicNetIPv4.
func SchemaFromServerPublicNetIPv4(s ServerPublicNetIPv4) schema.ServerPublicNetIPv4 {
	return c.SchemaFromServerPublicNetIPv4(s)
}

// ServerPublicNetIPv6FromSchema converts a schema.ServerPublicNetIPv6 to
// a ServerPublicNetIPv6.
func ServerPublicNetIPv6FromSchema(s schema.ServerPublicNetIPv6) ServerPublicNetIPv6 {
	return c.ServerPublicNetIPv6FromSchema(s)
}

// SchemaFromServerPublicNetIPv6 converts a ServerPublicNetIPv6 to
// a schema.ServerPublicNetIPv6.
func SchemaFromServerPublicNetIPv6(s ServerPublicNetIPv6) schema.ServerPublicNetIPv6 {
	return c.SchemaFromServerPublicNetIPv6(s)
}

// ServerPrivateNetFromSchema converts a schema.ServerPrivateNet to a ServerPrivateNet.
func ServerPrivateNetFromSchema(s schema.ServerPrivateNet) ServerPrivateNet {
	return c.ServerPrivateNetFromSchema(s)
}

// SchemaFromServerPrivateNet converts a ServerPrivateNet to a schema.ServerPrivateNet.
func SchemaFromServerPrivateNet(s ServerPrivateNet) schema.ServerPrivateNet {
	return c.SchemaFromServerPrivateNet(s)
}

// ServerTypeFromSchema converts a schema.ServerType to a ServerType.
func ServerTypeFromSchema(s schema.ServerType) *ServerType {
	return c.ServerTypeFromSchema(s)
}

// SchemaFromServerType converts a ServerType to a schema.ServerType.
func SchemaFromServerType(s *ServerType) schema.ServerType {
	return c.SchemaFromServerType(s)
}

// SSHKeyFromSchema converts a schema.SSHKey to a SSHKey.
func SSHKeyFromSchema(s schema.SSHKey) *SSHKey {
	return c.SSHKeyFromSchema(s)
}

// SchemaFromSSHKey converts a SSHKey to a schema.SSHKey.
func SchemaFromSSHKey(s *SSHKey) schema.SSHKey {
	return c.SchemaFromSSHKey(s)
}

// ImageFromSchema converts a schema.Image to an Image.
func ImageFromSchema(s schema.Image) *Image {
	return c.ImageFromSchema(s)
}

// SchemaFromImage converts an Image to a schema.Image.
func SchemaFromImage(i *Image) schema.Image {
	return c.SchemaFromImage(i)
}

// VolumeFromSchema converts a schema.Volume to a Volume.
func VolumeFromSchema(s schema.Volume) *Volume {
	return c.VolumeFromSchema(s)
}

// SchemaFromVolume converts a Volume to a schema.Volume.
func SchemaFromVolume(v *Volume) schema.Volume {
	return c.SchemaFromVolume(v)
}

// NetworkFromSchema converts a schema.Network to a Network.
func NetworkFromSchema(s schema.Network) *Network {
	return c.NetworkFromSchema(s)
}

// SchemaFromNetwork converts a Network to a schema.Network.
func SchemaFromNetwork(n *Network) schema.Network {
	return c.SchemaFromNetwork(n)
}

// NetworkSubnetFromSchema converts a schema.NetworkSubnet to a NetworkSubnet.
func NetworkSubnetFromSchema(s schema.NetworkSubnet) NetworkSubnet {
	return c.NetworkSubnetFromSchema(s)
}

// SchemaFromNetworkSubnet converts a NetworkSubnet to a schema.NetworkSubnet.
func SchemaFromNetworkSubnet(n NetworkSubnet) schema.NetworkSubnet {
	return c.SchemaFromNetworkSubnet(n)
}

// NetworkRouteFromSchema converts a schema.NetworkRoute to a NetworkRoute.
func NetworkRouteFromSchema(s schema.NetworkRoute) NetworkRoute {
	return c.NetworkRouteFromSchema(s)
}

// SchemaFromNetworkRoute converts a NetworkRoute to a schema.NetworkRoute.
func SchemaFromNetworkRoute(n NetworkRoute) schema.NetworkRoute {
	return c.SchemaFromNetworkRoute(n)
}

// LoadBalancerTypeFromSchema converts a schema.LoadBalancerType to a LoadBalancerType.
func LoadBalancerTypeFromSchema(s schema.LoadBalancerType) *LoadBalancerType {
	return c.LoadBalancerTypeFromSchema(s)
}

// SchemaFromLoadBalancerType converts a LoadBalancerType to a schema.LoadBalancerType.
func SchemaFromLoadBalancerType(l *LoadBalancerType) schema.LoadBalancerType {
	return c.SchemaFromLoadBalancerType(l)
}

// LoadBalancerFromSchema converts a schema.LoadBalancer to a LoadBalancer.
func LoadBalancerFromSchema(s schema.LoadBalancer) *LoadBalancer {
	return c.LoadBalancerFromSchema(s)
}

// SchemaFromLoadBalancer converts a LoadBalancer to a schema.LoadBalancer.
func SchemaFromLoadBalancer(l *LoadBalancer) schema.LoadBalancer {
	return c.SchemaFromLoadBalancer(l)
}

// LoadBalancerServiceFromSchema converts a schema.LoadBalancerService to a LoadBalancerService.
func LoadBalancerServiceFromSchema(s schema.LoadBalancerService) LoadBalancerService {
	return c.LoadBalancerServiceFromSchema(s)
}

// SchemaFromLoadBalancerService converts a LoadBalancerService to a schema.LoadBalancerService.
func SchemaFromLoadBalancerService(l LoadBalancerService) schema.LoadBalancerService {
	return c.SchemaFromLoadBalancerService(l)
}

// LoadBalancerServiceHealthCheckFromSchema converts a schema.LoadBalancerServiceHealthCheck to a LoadBalancerServiceHealthCheck.
func LoadBalancerServiceHealthCheckFromSchema(s *schema.LoadBalancerServiceHealthCheck) LoadBalancerServiceHealthCheck {
	return c.LoadBalancerServiceHealthCheckFromSchema(s)
}

// SchemaFromLoadBalancerServiceHealthCheck converts a LoadBalancerServiceHealthCheck to a schema.LoadBalancerServiceHealthCheck.
func SchemaFromLoadBalancerServiceHealthCheck(l LoadBalancerServiceHealthCheck) *schema.LoadBalancerServiceHealthCheck {
	return c.SchemaFromLoadBalancerServiceHealthCheck(l)
}

// LoadBalancerTargetFromSchema converts a schema.LoadBalancerTarget to a LoadBalancerTarget.
func LoadBalancerTargetFromSchema(s schema.LoadBalancerTarget) LoadBalancerTarget {
	return c.LoadBalancerTargetFromSchema(s)
}

// SchemaFromLoadBalancerTarget converts a LoadBalancerTarget to a schema.LoadBalancerTarget.
func SchemaFromLoadBalancerTarget(l LoadBalancerTarget) schema.LoadBalancerTarget {
	return c.SchemaFromLoadBalancerTarget(l)
}

// LoadBalancerTargetHealthStatusFromSchema converts a schema.LoadBalancerTarget to a LoadBalancerTarget.
func LoadBalancerTargetHealthStatusFromSchema(s schema.LoadBalancerTargetHealthStatus) LoadBalancerTargetHealthStatus {
	return c.LoadBalancerTargetHealthStatusFromSchema(s)
}

// SchemaFromLoadBalancerTargetHealthStatus converts a LoadBalancerTarget to a schema.LoadBalancerTarget.
func SchemaFromLoadBalancerTargetHealthStatus(l LoadBalancerTargetHealthStatus) schema.LoadBalancerTargetHealthStatus {
	return c.SchemaFromLoadBalancerTargetHealthStatus(l)
}

// CertificateFromSchema converts a schema.Certificate to a Certificate.
func CertificateFromSchema(s schema.Certificate) *Certificate {
	return c.CertificateFromSchema(s)
}

// SchemaFromCertificate converts a Certificate to a schema.Certificate.
func SchemaFromCertificate(ct *Certificate) schema.Certificate {
	return c.SchemaFromCertificate(ct)
}

// PaginationFromSchema converts a schema.MetaPagination to a Pagination.
func PaginationFromSchema(s schema.MetaPagination) Pagination {
	return c.PaginationFromSchema(s)
}

// SchemaFromPagination converts a Pagination to a schema.MetaPagination.
func SchemaFromPagination(p Pagination) schema.MetaPagination {
	return c.SchemaFromPagination(p)
}

// ErrorFromSchema converts a schema.Error to an Error.
func ErrorFromSchema(s schema.Error) Error {
	return c.ErrorFromSchema(s)
}

// SchemaFromError converts an Error to a schema.Error.
func SchemaFromError(e Error) schema.Error {
	return c.SchemaFromError(e)
}

// PricingFromSchema converts a schema.Pricing to a Pricing.
func PricingFromSchema(s schema.Pricing) Pricing {
	return c.PricingFromSchema(s)
}

// SchemaFromPricing converts a Pricing to a schema.Pricing.
func SchemaFromPricing(p Pricing) schema.Pricing {
	return c.SchemaFromPricing(p)
}

// FirewallFromSchema converts a schema.Firewall to a Firewall.
func FirewallFromSchema(s schema.Firewall) *Firewall {
	return c.FirewallFromSchema(s)
}

// SchemaFromFirewall converts a Firewall to a schema.Firewall.
func SchemaFromFirewall(f *Firewall) schema.Firewall {
	return c.SchemaFromFirewall(f)
}

// PlacementGroupFromSchema converts a schema.PlacementGroup to a PlacementGroup.
func PlacementGroupFromSchema(s schema.PlacementGroup) *PlacementGroup {
	return c.PlacementGroupFromSchema(s)
}

// SchemaFromPlacementGroup converts a PlacementGroup to a schema.PlacementGroup.
func SchemaFromPlacementGroup(p *PlacementGroup) schema.PlacementGroup {
	return c.SchemaFromPlacementGroup(p)
}

// DeprecationFromSchema converts a [schema.DeprecationInfo] to a [DeprecationInfo].
func DeprecationFromSchema(s *schema.DeprecationInfo) *DeprecationInfo {
	return c.DeprecationFromSchema(s)
}

// SchemaFromDeprecation converts a [DeprecationInfo] to a [schema.DeprecationInfo].
func SchemaFromDeprecation(d *DeprecationInfo) *schema.DeprecationInfo {
	return c.SchemaFromDeprecation(d)
}

func placementGroupCreateOptsToSchema(opts PlacementGroupCreateOpts) schema.PlacementGroupCreateRequest {
	return c.SchemaFromPlacementGroupCreateOpts(opts)
}

func loadBalancerCreateOptsToSchema(opts LoadBalancerCreateOpts) schema.LoadBalancerCreateRequest {
	return c.SchemaFromLoadBalancerCreateOpts(opts)
}

func loadBalancerAddServiceOptsToSchema(opts LoadBalancerAddServiceOpts) schema.LoadBalancerActionAddServiceRequest {
	return c.SchemaFromLoadBalancerAddServiceOpts(opts)
}

func loadBalancerUpdateServiceOptsToSchema(opts LoadBalancerUpdateServiceOpts) schema.LoadBalancerActionUpdateServiceRequest {
	return c.SchemaFromLoadBalancerUpdateServiceOpts(opts)
}

func firewallCreateOptsToSchema(opts FirewallCreateOpts) schema.FirewallCreateRequest {
	return c.SchemaFromFirewallCreateOpts(opts)
}

func firewallSetRulesOptsToSchema(opts FirewallSetRulesOpts) schema.FirewallActionSetRulesRequest {
	return c.SchemaFromFirewallSetRulesOpts(opts)
}

func firewallResourceToSchema(resource FirewallResource) schema.FirewallResource {
	return c.SchemaFromFirewallResource(resource)
}

func serverMetricsFromSchema(s *schema.ServerGetMetricsResponse) (*ServerMetrics, error) {
	return c.ServerMetricsFromSchema(s)
}

func loadBalancerMetricsFromSchema(s *schema.LoadBalancerGetMetricsResponse) (*LoadBalancerMetrics, error) {
	return c.LoadBalancerMetricsFromSchema(s)
}
