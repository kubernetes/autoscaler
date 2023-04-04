package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// ServerStorageRelationOperator provides an interface for operations on server-storage relations.
type ServerStorageRelationOperator interface {
	GetServerStorageList(ctx context.Context, id string) ([]ServerStorageRelationProperties, error)
	GetServerStorage(ctx context.Context, serverID, storageID string) (ServerStorageRelationProperties, error)
	CreateServerStorage(ctx context.Context, id string, body ServerStorageRelationCreateRequest) error
	UpdateServerStorage(ctx context.Context, serverID, storageID string, body ServerStorageRelationUpdateRequest) error
	DeleteServerStorage(ctx context.Context, serverID, storageID string) error
	LinkStorage(ctx context.Context, serverID string, storageID string, bootdevice bool) error
	UnlinkStorage(ctx context.Context, serverID string, storageID string) error
}

// ServerStorageRelationList holds a list of relations between a server and storages.
type ServerStorageRelationList struct {
	// Array of relations between a server and storages.
	List []ServerStorageRelationProperties `json:"storage_relations"`
}

// ServerStorageRelationSingle represents a single relation between a server and a storage.
type ServerStorageRelationSingle struct {
	// Properties of a relation between a server and a storage.
	Properties ServerStorageRelationProperties `json:"storage_relation"`
}

// ServerStorageRelationProperties holds properties of a relation between a server and a storage.
type ServerStorageRelationProperties struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`

	// The capacity of a storage/ISO image/template/snapshot in GB.
	Capacity int `json:"capacity"`

	// Indicates the speed of the storage. This may be (storage, storage_high or storage_insane).
	StorageType string `json:"storage_type"`

	// Defines the SCSI target ID. The SCSI defines transmission routes like Serial Attached SCSI (SAS), Fibre Channel and iSCSI.
	// The target ID is a device (e.g. disk).
	Target int `json:"target"`

	// Is the common SCSI abbreviation of the Logical Unit Number. A LUN is a unique identifier for a single disk or a composite of disks.
	Lun int `json:"lun"`

	// Defines the SCSI controller id. The SCSI defines transmission routes such as Serial Attached SCSI (SAS), Fibre Channel and iSCSI.
	Controller int `json:"controller"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Defines if this object is the boot device. Storages, Networks and ISO images can have a boot device configured,
	// but only one boot device per Storage, Network or ISO image.
	// The boot order is as follows => Network > ISO image > Storage.
	BootDevice bool `json:"bootdevice"`

	// The SCSI bus id. The SCSI defines transmission routes like Serial Attached SCSI (SAS), Fibre Channel and iSCSI.
	// Each SCSI device is addressed via a specific number. Each SCSI bus can have multiple SCSI devices connected to it.
	Bus int `json:"bus"`

	// Indicates the UUID of the last used template on this storage (inherited from snapshots).
	LastUsedTemplate string `json:"last_used_template"`

	// If a template has been used that requires a license key (e.g. Windows Servers)
	// this shows the product_no of the license (see the /prices endpoint for more details).
	LicenseProductNo int `json:"license_product_no"`

	// The same as the object_uuid.
	ServerUUID string `json:"server_uuid"`
}

// ServerStorageRelationCreateRequest represents a request for creating a relation between a server and a storage.
type ServerStorageRelationCreateRequest struct {
	// The UUID of the storage you are requesting. If server's hardware profile is default, nested, q35 or q35_nested,
	// you are allowed to attached 8 servers. Only 2 storage are allowed to be attached to server with other hardware profile.
	ObjectUUID string `json:"object_uuid"`

	// Whether the server will boot from this storage device or not. Optional.
	BootDevice bool `json:"bootdevice,omitempty"`
}

// ServerStorageRelationUpdateRequest represents a request for updating a relation between a server and a storage.
type ServerStorageRelationUpdateRequest struct {
	// The ordering of the network interfaces. Lower numbers have lower PCI-IDs. Optional.
	Ordering int `json:"ordering,omitempty"`

	// Whether the server boots from this network or not. Optional.
	BootDevice bool `json:"bootdevice,omitempty"`

	// Defines information about IP prefix spoof protection (it allows source traffic only from the IPv4/IPv4 network prefixes).
	// If empty, it allow no IPv4/IPv6 source traffic. If set to null, l3security is disabled (default). Optional.
	L3security []string `json:"l3security,omitempty"`
}

// GetServerStorageList gets a list of a specific server's storages.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getServerLinkedStorages
func (c *Client) GetServerStorageList(ctx context.Context, id string) ([]ServerStorageRelationProperties, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiServerBase, id, "storages"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ServerStorageRelationList
	err := r.execute(ctx, *c, &response)
	return response.List, err
}

// GetServerStorage gets a storage of a specific server.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getServerLinkedStorage
func (c *Client) GetServerStorage(ctx context.Context, serverID, storageID string) (ServerStorageRelationProperties, error) {
	if !isValidUUID(serverID) || !isValidUUID(storageID) {
		return ServerStorageRelationProperties{}, errors.New("'serverID' or 'storageID' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiServerBase, serverID, "storages", storageID),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ServerStorageRelationSingle
	err := r.execute(ctx, *c, &response)
	return response.Properties, err
}

// UpdateServerStorage updates a link between a storage and a server.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateServerLinkedStorage
func (c *Client) UpdateServerStorage(ctx context.Context, serverID, storageID string, body ServerStorageRelationUpdateRequest) error {
	if !isValidUUID(serverID) || !isValidUUID(storageID) {
		return errors.New("'serverID' or 'storageID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiServerBase, serverID, "storages", storageID),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// CreateServerStorage creates a link between a server and a storage.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/linkStorageToServer
func (c *Client) CreateServerStorage(ctx context.Context, id string, body ServerStorageRelationCreateRequest) error {
	if !isValidUUID(id) || !isValidUUID(body.ObjectUUID) {
		return errors.New("'server_id' or 'storage_id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiServerBase, id, "storages"),
		method: http.MethodPost,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteServerStorage removes a link between a storage and a server.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/unlinkStorageFromServer
func (c *Client) DeleteServerStorage(ctx context.Context, serverID, storageID string) error {
	if !isValidUUID(serverID) || !isValidUUID(storageID) {
		return errors.New("'serverID' or 'storageID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiServerBase, serverID, "storages", storageID),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// LinkStorage attaches a storage to a server.
func (c *Client) LinkStorage(ctx context.Context, serverID string, storageID string, bootdevice bool) error {
	body := ServerStorageRelationCreateRequest{
		ObjectUUID: storageID,
		BootDevice: bootdevice,
	}
	return c.CreateServerStorage(ctx, serverID, body)
}

// UnlinkStorage detaches a storage from a server.
func (c *Client) UnlinkStorage(ctx context.Context, serverID string, storageID string) error {
	return c.DeleteServerStorage(ctx, serverID, storageID)
}
