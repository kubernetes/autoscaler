package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// StorageSnapshotScheduleOperator provides an interface for operations on snapshot schedules.
type StorageSnapshotScheduleOperator interface {
	GetStorageSnapshotScheduleList(ctx context.Context, id string) ([]StorageSnapshotSchedule, error)
	GetStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string) (StorageSnapshotSchedule, error)
	CreateStorageSnapshotSchedule(ctx context.Context, id string, body StorageSnapshotScheduleCreateRequest)
	UpdateStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string, body StorageSnapshotScheduleUpdateRequest)
	DeleteStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string) error
}

// StorageSnapshotScheduleList holds a list of storage snapshot schedules.
type StorageSnapshotScheduleList struct {
	// Array of storage snapshot schedules.
	List map[string]StorageSnapshotScheduleProperties `json:"snapshot_schedules"`
}

// StorageSnapshotSchedule represents a single storage snapshot schedule.
type StorageSnapshotSchedule struct {
	// Properties of a storage snapshot schedule.
	Properties StorageSnapshotScheduleProperties `json:"snapshot_schedule"`
}

// StorageSnapshotScheduleProperties holds properties of a single storage snapshot schedule.
type StorageSnapshotScheduleProperties struct {
	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The amount of Snapshots to keep before overwriting the last created Snapshot.
	// value >= 1.
	KeepSnapshots int `json:"keep_snapshots"`

	// List of labels.
	Labels []string `json:"labels"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The date and time that the snapshot schedule will be run.
	NextRuntime GSTime `json:"next_runtime"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Related snapshots (snapshots taken by this snapshot schedule).
	Relations StorageSnapshotScheduleRelations `json:"relations"`

	// The interval at which the schedule will run (in minutes).
	// value >= 60.
	RunInterval int `json:"run_interval"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// UUID of the storage that will be used for taking snapshots.
	StorageUUID string `json:"storage_uuid"`
}

// StorageSnapshotScheduleRelations holds a list of relations between a storage snapshot schedule and storage snapshots.
type StorageSnapshotScheduleRelations struct {
	// Array of all related snapshots (snapshots taken by this snapshot schedule).
	Snapshots []StorageSnapshotScheduleRelation `json:"snapshots"`
}

// StorageSnapshotScheduleRelation represents a relation between a storage snapshot schedule and a storage snapshot.
type StorageSnapshotScheduleRelation struct {
	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// StorageSnapshotScheduleCreateRequest represents a request for creating a storage snapshot schedule.
type StorageSnapshotScheduleCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// List of labels. Optional.
	Labels []string `json:"labels,omitempty"`

	// The interval at which the schedule will run (in minutes).
	// Allowed value >= 60
	RunInterval int `json:"run_interval"`

	// The amount of Snapshots to keep before overwriting the last created Snapshot.
	// Allowed value >= 1
	KeepSnapshots int `json:"keep_snapshots"`

	// The date and time that the snapshot schedule will be run. Optional.
	NextRuntime *GSTime `json:"next_runtime,omitempty"`
}

// StorageSnapshotScheduleCreateResponse represents a response for creating a storage snapshot schedule.
type StorageSnapshotScheduleCreateResponse struct {
	// UUID of the request.
	RequestUUID string `json:"request_uuid"`

	// UUID of the snapshot schedule being created.
	ObjectUUID string `json:"object_uuid"`
}

// StorageSnapshotScheduleUpdateRequest represents a request for updating a storage snapshot schedule.
type StorageSnapshotScheduleUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	// Optional.
	Name string `json:"name,omitempty"`

	// List of labels. Optional.
	Labels *[]string `json:"labels,omitempty"`

	// The interval at which the schedule will run (in minutes). Optional.
	// Allowed value >= 60
	RunInterval int `json:"run_interval,omitempty"`

	// The amount of Snapshots to keep before overwriting the last created Snapshot. Optional.
	// Allowed value >= 1
	KeepSnapshots int `json:"keep_snapshots,omitempty"`

	// The date and time that the snapshot schedule will be run. Optional.
	NextRuntime *GSTime `json:"next_runtime,omitempty"`
}

// GetStorageSnapshotScheduleList gets a list of available storage snapshot schedules based on a given storage's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getSnapshotSchedules
func (c *Client) GetStorageSnapshotScheduleList(ctx context.Context, id string) ([]StorageSnapshotSchedule, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiStorageBase, id, "snapshot_schedules"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageSnapshotScheduleList
	var schedules []StorageSnapshotSchedule
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		schedules = append(schedules, StorageSnapshotSchedule{Properties: properties})
	}
	return schedules, err
}

// GetStorageSnapshotSchedule gets a specific storage snapshot schedule based on a given storage's id and schedule's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getSnapshotSchedule
func (c *Client) GetStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string) (StorageSnapshotSchedule, error) {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return StorageSnapshotSchedule{}, errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiStorageBase, storageID, "snapshot_schedules", scheduleID),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response StorageSnapshotSchedule
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateStorageSnapshotSchedule creates a storage snapshot schedule.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createSnapshotSchedule
func (c *Client) CreateStorageSnapshotSchedule(ctx context.Context, id string, body StorageSnapshotScheduleCreateRequest) (
	StorageSnapshotScheduleCreateResponse, error) {
	if !isValidUUID(id) {
		return StorageSnapshotScheduleCreateResponse{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, id, "snapshot_schedules"),
		method: http.MethodPost,
		body:   body,
	}
	var response StorageSnapshotScheduleCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateStorageSnapshotSchedule updates specific storage snapshot schedule based on a given storage's id and schedule's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateSnapshotSchedule
func (c *Client) UpdateStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string,
	body StorageSnapshotScheduleUpdateRequest) error {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "snapshot_schedules", scheduleID),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteStorageSnapshotSchedule removes specific storage snapshot scheduler based on a given storage's id and scheduler's id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteSnapshotSchedule
func (c *Client) DeleteStorageSnapshotSchedule(ctx context.Context, storageID, scheduleID string) error {
	if !isValidUUID(storageID) || !isValidUUID(scheduleID) {
		return errors.New("'storageID' or 'scheduleID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiStorageBase, storageID, "snapshot_schedules", scheduleID),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}
