package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// StorageBackupOperator provides an interface for operations on storage backups.
type StorageBackupOperator interface {
	GetStorageBackupList(ctx context.Context, id string) ([]StorageBackup, error)
	DeleteStorageBackup(ctx context.Context, storageID, backupID string) error
	RollbackStorageBackup(ctx context.Context, storageID, backupID string, body StorageRollbackRequest) error
}

// StorageBackupList holds of a list of storage backups.
type StorageBackupList struct {
	// Array of backups.
	List map[string]StorageBackupProperties `json:"backups"`
}

// StorageBackup represents a single storage backup.
type StorageBackup struct {
	// Properties of a backup.
	Properties StorageBackupProperties `json:"backup"`
}

// StorageBackupProperties holds the properties of a single backup.
type StorageBackupProperties struct {
	// The UUID of a backup is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The name of the backup equals schedule name plus backup UUID.
	Name string `json:"name"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The size of a backup in GB.
	Capacity int `json:"capacity"`
}

// GetStorageBackupList gets a list of available storage backups.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getStorageBackups
func (c *Client) GetStorageBackupList(ctx context.Context, id string) ([]StorageBackup, error) {
	r := gsRequest{
		uri:                 path.Join(apiStorageBase, id, "backups"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageBackupList
	var storageBackups []StorageBackup
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		storageBackups = append(storageBackups, StorageBackup{
			Properties: properties,
		})
	}
	return storageBackups, err
}

// DeleteStorageBackup removes a specific storage's backup.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteStorageBackup
func (c *Client) DeleteStorageBackup(ctx context.Context, storageID, backupID string) error {
	if !isValidUUID(storageID) || !isValidUUID(backupID) {
		return errors.New("'storageID' or 'backupID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "backups", backupID),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// RollbackStorageBackup rollbacks a storage from a storage backup.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/rollbackStorageBackup
func (c *Client) RollbackStorageBackup(ctx context.Context, storageID, backupID string, body StorageRollbackRequest) error {
	if !isValidUUID(storageID) || !isValidUUID(backupID) {
		return errors.New("'storageID' or 'backupID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "backups", backupID, "rollback"),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}
