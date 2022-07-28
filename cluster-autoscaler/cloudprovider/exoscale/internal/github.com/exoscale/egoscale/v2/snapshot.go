/*
Copyright 2021 The Kubernetes Authors.

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

package v2

import (
	"context"
	"time"

	apiv2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/api"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/oapi"
)

// SnapshotExport represents exported Snapshot information.
type SnapshotExport struct {
	MD5sum       *string
	PresignedURL *string
}

// Snapshot represents a Snapshot.
type Snapshot struct {
	CreatedAt  *time.Time
	ID         *string `req-for:"update,delete"`
	InstanceID *string
	Name       *string
	Size       *int64
	State      *string
	Zone       *string
}

func snapshotFromAPI(s *oapi.Snapshot, zone string) *Snapshot {
	return &Snapshot{
		CreatedAt:  s.CreatedAt,
		ID:         s.Id,
		InstanceID: s.Instance.Id,
		Name:       s.Name,
		Size:       s.Size,
		State:      (*string)(s.State),
		Zone:       &zone,
	}
}

// DeleteSnapshot deletes a Snapshot.
func (c *Client) DeleteSnapshot(ctx context.Context, zone string, snapshot *Snapshot) error {
	if err := validateOperationParams(snapshot, "delete"); err != nil {
		return err
	}

	resp, err := c.DeleteSnapshotWithResponse(apiv2.WithZone(ctx, zone), *snapshot.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ExportSnapshot exports a Snapshot and returns the exported Snapshot information.
func (c *Client) ExportSnapshot(ctx context.Context, zone string, snapshot *Snapshot) (*SnapshotExport, error) {
	if err := validateOperationParams(snapshot, "update"); err != nil {
		return nil, err
	}

	resp, err := c.ExportSnapshotWithResponse(apiv2.WithZone(ctx, zone), *snapshot.ID)
	if err != nil {
		return nil, err
	}

	res, err := oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	expSnapshot, err := c.GetSnapshotWithResponse(apiv2.WithZone(ctx, zone), *res.(*oapi.Reference).Id)
	if err != nil {
		return nil, err
	}

	return &SnapshotExport{
		MD5sum:       expSnapshot.JSON200.Export.Md5sum,
		PresignedURL: expSnapshot.JSON200.Export.PresignedUrl,
	}, nil
}

// GetSnapshot returns the Snapshot corresponding to the specified ID.
func (c *Client) GetSnapshot(ctx context.Context, zone, id string) (*Snapshot, error) {
	resp, err := c.GetSnapshotWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return snapshotFromAPI(resp.JSON200, zone), nil
}

// ListSnapshots returns the list of existing Snapshots.
func (c *Client) ListSnapshots(ctx context.Context, zone string) ([]*Snapshot, error) {
	list := make([]*Snapshot, 0)

	resp, err := c.ListSnapshotsWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Snapshots != nil {
		for i := range *resp.JSON200.Snapshots {
			list = append(list, snapshotFromAPI(&(*resp.JSON200.Snapshots)[i], zone))
		}
	}

	return list, nil
}
