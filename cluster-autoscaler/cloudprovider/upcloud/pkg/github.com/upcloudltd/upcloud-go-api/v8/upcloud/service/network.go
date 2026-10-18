package service

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/request"
)

type Network interface {
	GetNetworks(ctx context.Context, f ...request.QueryFilter) (*upcloud.Networks, error)
	GetNetworksInZone(ctx context.Context, r *request.GetNetworksInZoneRequest) (*upcloud.Networks, error)
	CreateNetwork(ctx context.Context, r *request.CreateNetworkRequest) (*upcloud.Network, error)
	GetNetworkDetails(ctx context.Context, r *request.GetNetworkDetailsRequest) (*upcloud.Network, error)
	ModifyNetwork(ctx context.Context, r *request.ModifyNetworkRequest) (*upcloud.Network, error)
	DeleteNetwork(ctx context.Context, r *request.DeleteNetworkRequest) error
	AttachNetworkRouter(ctx context.Context, r *request.AttachNetworkRouterRequest) error
	DetachNetworkRouter(ctx context.Context, r *request.DetachNetworkRouterRequest) error
	GetServerNetworks(ctx context.Context, r *request.GetServerNetworksRequest) (*upcloud.Networking, error)
	CreateNetworkInterface(ctx context.Context, r *request.CreateNetworkInterfaceRequest) (*upcloud.Interface, error)
	ModifyNetworkInterface(ctx context.Context, r *request.ModifyNetworkInterfaceRequest) (*upcloud.Interface, error)
	DeleteNetworkInterface(ctx context.Context, r *request.DeleteNetworkInterfaceRequest) error
	AssignIPAddressToNetworkInterface(ctx context.Context, r *request.AssignIPAddressToNetworkInterfaceRequest) (*upcloud.IPAddress, error)
	DeleteIPAddressFromNetworkInterface(ctx context.Context, r *request.DeleteIPAddressFromNetworkInterfaceRequest) error
	GetRouters(ctx context.Context, f ...request.QueryFilter) (*upcloud.Routers, error)
	GetRouterDetails(ctx context.Context, r *request.GetRouterDetailsRequest) (*upcloud.Router, error)
	CreateRouter(ctx context.Context, r *request.CreateRouterRequest) (*upcloud.Router, error)
	ModifyRouter(ctx context.Context, r *request.ModifyRouterRequest) (*upcloud.Router, error)
	DeleteRouter(ctx context.Context, r *request.DeleteRouterRequest) error
}

// GetNetworks returns the all the available networks
func (s *Service) GetNetworks(ctx context.Context, f ...request.QueryFilter) (*upcloud.Networks, error) {
	req := request.GetNetworksRequest{Filters: f}
	networks := upcloud.Networks{}
	return &networks, s.get(ctx, req.RequestURL(), &networks)
}

// GetNetworksInZone returns the all the available networks within the specified zone.
func (s *Service) GetNetworksInZone(ctx context.Context, r *request.GetNetworksInZoneRequest) (*upcloud.Networks, error) {
	networks := upcloud.Networks{}
	return &networks, s.get(ctx, r.RequestURL(), &networks)
}

// CreateNetwork creates a new network and returns the network details for the new network.
func (s *Service) CreateNetwork(ctx context.Context, r *request.CreateNetworkRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	return &network, s.create(ctx, r, &network)
}

// GetNetworkDetails returns the details for the specified network.
func (s *Service) GetNetworkDetails(ctx context.Context, r *request.GetNetworkDetailsRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	return &network, s.get(ctx, r.RequestURL(), &network)
}

// ModifyNetwork modifies the existing specified network.
func (s *Service) ModifyNetwork(ctx context.Context, r *request.ModifyNetworkRequest) (*upcloud.Network, error) {
	network := upcloud.Network{}
	return &network, s.replace(ctx, r, &network)
}

// DeleteNetwork deletes the specified network.
func (s *Service) DeleteNetwork(ctx context.Context, r *request.DeleteNetworkRequest) error {
	return s.delete(ctx, r)
}

// AttachNetworkRouter attaches a router to the specified network.
func (s *Service) AttachNetworkRouter(ctx context.Context, r *request.AttachNetworkRouterRequest) error {
	return s.replace(ctx, r, nil)
}

// DetachNetworkRouter detaches a router from the specified network.
func (s *Service) DetachNetworkRouter(ctx context.Context, r *request.DetachNetworkRouterRequest) error {
	return s.replace(ctx, r, nil)
}

// GetServerNetworks returns all the networks associated with the specified server.
func (s *Service) GetServerNetworks(ctx context.Context, r *request.GetServerNetworksRequest) (*upcloud.Networking, error) {
	networking := upcloud.Networking{}
	return &networking, s.get(ctx, r.RequestURL(), &networking)
}

// CreateNetworkInterface creates a new network interface on the specified server.
func (s *Service) CreateNetworkInterface(ctx context.Context, r *request.CreateNetworkInterfaceRequest) (*upcloud.Interface, error) {
	iface := upcloud.Interface{}
	return &iface, s.create(ctx, r, &iface)
}

// ModifyNetworkInterface modifies the specified network interface on the specified server.
func (s *Service) ModifyNetworkInterface(ctx context.Context, r *request.ModifyNetworkInterfaceRequest) (*upcloud.Interface, error) {
	iface := upcloud.Interface{}
	return &iface, s.replace(ctx, r, &iface)
}

// DeleteNetworkInterface removes the specified network interface from the specified server.
func (s *Service) DeleteNetworkInterface(ctx context.Context, r *request.DeleteNetworkInterfaceRequest) error {
	return s.delete(ctx, r)
}

func (s *Service) AssignIPAddressToNetworkInterface(ctx context.Context, r *request.AssignIPAddressToNetworkInterfaceRequest) (*upcloud.IPAddress, error) {
	ip := upcloud.IPAddress{}
	return &ip, s.create(ctx, r, &ip)
}

func (s *Service) DeleteIPAddressFromNetworkInterface(ctx context.Context, r *request.DeleteIPAddressFromNetworkInterfaceRequest) error {
	return s.delete(ctx, r)
}

// GetRouters returns the all the available routers
func (s *Service) GetRouters(ctx context.Context, f ...request.QueryFilter) (*upcloud.Routers, error) {
	r := request.GetRoutersRequest{Filters: f}
	routers := upcloud.Routers{}
	return &routers, s.get(ctx, r.RequestURL(), &routers)
}

// GetRouterDetails returns the details for the specified router.
func (s *Service) GetRouterDetails(ctx context.Context, r *request.GetRouterDetailsRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	return &router, s.get(ctx, r.RequestURL(), &router)
}

// CreateRouter creates a new router.
func (s *Service) CreateRouter(ctx context.Context, r *request.CreateRouterRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	return &router, s.create(ctx, r, &router)
}

// ModifyRouter modifies the configuration of the specified existing router.
func (s *Service) ModifyRouter(ctx context.Context, r *request.ModifyRouterRequest) (*upcloud.Router, error) {
	router := upcloud.Router{}
	return &router, s.modify(ctx, r, &router)
}

// DeleteRouter deletes the specified router.
func (s *Service) DeleteRouter(ctx context.Context, r *request.DeleteRouterRequest) error {
	return s.delete(ctx, r)
}
