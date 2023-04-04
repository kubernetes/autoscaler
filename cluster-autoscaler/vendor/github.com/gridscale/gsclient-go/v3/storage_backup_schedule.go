package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// StorageBackupScheduleOperator provides an interface for operations on backup schedules.
type StorageBackupScheduleOperator interface {
	GetStorageBackupScheduleList(ctx context.Context, id string) ([]StorageBackupSchedule, error)
	GetStorageBackupSchedule(ctx context.Context, storageID, scheduleID string) (StorageBackupSchedule, error)
	CreateStorageBackupSchedule(ctx context.Context, id string, body StorageBackupScheduleCreateRequest)
	UpdateStorageBackupSchedule(ctx context.Context, storageID, scheduleID string, body StorageBackupScheduleUpdateRequest) error
	DeleteStorageBackupSchedule(ctx context.Context, storageID, scheduleID string) error
}

// StorageBackupScheduleList contains a list of storage backup schedules.
type StorageBackupScheduleList struct {
	List map[string]StorageBackupScheduleProperties `json:"schedule_storage_backups"`
}

// StorageBackupSchedule represents a single storage backup schedule.
type StorageBackupSchedule struct {
	Properties StorageBackupScheduleProperties `json:"schedule_storage_backup"`
}

// StorageBackupScheduleProperties contains properties of a storage backup schedule.
type StorageBackupScheduleProperties struct {
	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The amount of backups to keep before overwriting the last created backup.
	// value >= 1.
	KeepBackups int `json:"keep_backups"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The date and time that the storage backup schedule will be run.
	NextRuntime GSTime `json:"next_runtime"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Related backups (backups taken by this storage backup schedule)
	Relations StorageBackupScheduleRelations `json:"relations"`

	// The interval at which the schedule will run (in minutes)
	// value >= 60.
	RunInterval int `json:"run_interval"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// UUID of the storage that will be used for making taking backups
	StorageUUID string `json:"storage_uuid"`

	// Status of the schedule.
	Active bool `json:"active"`

	// The UUID of the location where your backup is stored.
	BackupLocationUUID string `json:"backup_location_uuid"`

	// The human-readable name of backup location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	BackupLocationName string `json:"backup_location_name"`
}

// StorageBackupScheduleRelations contains a list of relations between a storage backup schedule and storage backups.
type StorageBackupScheduleRelations struct {
	// Array of all related backups (backups taken by this storage backup schedule).
	StorageBackups []StorageBackupScheduleRelation `json:"storages_backups"`
}

// StorageBackupScheduleRelation represents a relation between a storage backup schedule and a storage backup.
type StorageBackupScheduleRelation struct {
	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// StorageBackupScheduleCreateRequest represents a request for creating a storage backup schedule.
type StorageBackupScheduleCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The interval at which the schedule will run (in minutes).
	// Allowed value >= 60.
	RunInterval int `json:"run_interval"`

	// The amount of backups to keep before overwriting the last created backup.
	// value >= 1.
	KeepBackups int `json:"keep_backups"`

	// The date and time that the storage backup schedule will be run.
	NextRuntime GSTime `json:"next_runtime"`

	// Status of the schedule.
	Active bool `json:"active"`

	// The UUID of the location where your backup is stored.
	BackupLocationUUID string `json:"backup_location_uuid,omitempty"`
}

// StorageBackupScheduleCreateResponse represents a response for creating a storage backup schedule.
type StorageBackupScheduleCreateResponse struct {
	// UUID of the request.
	RequestUUID string `json:"request_uuid"`

	// UUID of the storage backup schedule being created.
	ObjectUUID string `json:"object_uuid"`
}

// StorageBackupScheduleUpdateRequest represents a request for updating a storage backup schedule.
type StorageBackupScheduleUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	// Optional.
	Name string `json:"name,omitempty"`

	// The interval at which the schedule will run (in minutes). Optional.
	// Allowed value >= 60
	RunInterval int `json:"run_interval,omitempty"`

	// The amount of backups to keep before overwriting the last created backup. Optional.
	// value >= 1
	KeepBackups int `json:"keep_backups,omitempty"`

	// The date and time that the storage backup schedule will be run. Optional.
	NextRuntime *GSTime `json:"next_runtime,omitempty"`

	// Status of the schedule. Optional.
	Active *bool `json:"active,omitempty"`
}

// StorageBackupLocationList contains a list of available location to store your backup.
type StorageBackupLocationList struct {
	List map[string]StorageBackupLocationProperties `json:"backup_locations"`
}

//StorageBackupLocation represents a backup location.
type StorageBackupLocation struct {
	Properties StorageBackupLocationProperties
}

// StorageBackupLocationProperties represents a backup location's properties.
type StorageBackupLocationProperties struct {
	// UUID of the location.
	ObjectUUID string `json:"object_uuid"`

	// Name of the location.
	Name string `json:"name"`
}

// GetStorageBackupScheduleList returns a list of available storage backup schedules based on a given storage's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getStorageBackupSchedules
func (c *Client) GetStorageBackupScheduleList(ctx context.Context, id string) ([]StorageBackupSchedule, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiStorageBase, id, "backup_schedules"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageBackupScheduleList
	var schedules []StorageBackupSchedule
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		schedules = append(schedules, StorageBackupSchedule{Properties: properties})
	}
	return schedules, err
}

// GetStorageBackupSchedule returns a specific storage backup schedule based on a given storage's id and a backup schedule's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getStorageBackupSchedules
func (c *Client) GetStorageBackupSchedule(ctx context.Context, storageID, scheduleID string) (StorageBackupSchedule, error) {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return StorageBackupSchedule{}, errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiStorageBase, storageID, "backup_schedules", scheduleID),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageBackupSchedule
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateStorageBackupSchedule creates a storage backup schedule based on a given storage UUID.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getStorageBackupSchedule
func (c *Client) CreateStorageBackupSchedule(ctx context.Context, id string, body StorageBackupScheduleCreateRequest) (
	StorageBackupScheduleCreateResponse, error) {
	if !isValidUUID(id) {
		return StorageBackupScheduleCreateResponse{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, id, "backup_schedules"),
		method: http.MethodPost,
		body:   body,
	}
	var response StorageBackupScheduleCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateStorageBackupSchedule updates a specific storage backup schedule based on a given storage's id and a backup schedule's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateStorageBackupSchedule
func (c *Client) UpdateStorageBackupSchedule(ctx context.Context, storageID, scheduleID string,
	body StorageBackupScheduleUpdateRequest) error {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "backup_schedules", scheduleID),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteStorageBackupSchedule removes a specific storage backup scheduler based on a given storage's id and a backup schedule's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteStorageBackupSchedule
func (c *Client) DeleteStorageBackupSchedule(ctx context.Context, storageID, scheduleID string) error {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "backup_schedules", scheduleID),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetStorageBackupLocationList returns a list of available locations to store your backup.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/GetBackupLocations
func (c *Client) GetStorageBackupLocationList(ctx context.Context) ([]StorageBackupLocation, error) {
	r := gsRequest{
		uri:                 apiBackupLocationBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageBackupLocationList
	var locationList []StorageBackupLocation
	err := r.execute(ctx, *c, &response)
	for _, locationProperties := range response.List {
		locationList = append(locationList, StorageBackupLocation{
			Properties: locationProperties,
		})
	}
	return locationList, err
}
