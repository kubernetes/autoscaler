// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	snapshotPath = "/snapshots"
)

var _ SnapshotService = (*snapshot)(nil)

// SnapshotService is an interface to interact with BizFly API Snapshot endpoint.
type SnapshotService interface {
	List(ctx context.Context, opts *ListOptions) ([]*Snapshot, error)
	Create(ctx context.Context, scr *SnapshotCreateRequest) (*Snapshot, error)
	Get(ctx context.Context, id string) (*Snapshot, error)
	Delete(ctx context.Context, id string) error
}

// SnapshotCreateRequest represents create new volume request payload.
type SnapshotCreateRequest struct {
	Name     string `json:"name"`
	VolumeId string `json:"volume_id"`
	Force    bool   `json:"force"`
}

// Snapshot contains snapshot information
type Snapshot struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	VolumeTypeId     string            `json:"volume_type_id"`
	VolumeId         string            `json:"volume_id"`
	Size             int               `json:"size"`
	Progress         string            `json:"os-extended-snapshot-attributes:progress"`
	TenantId         string            `json:"os-extended-snapshot-attributes:project_id"`
	Metadata         map[string]string `json:"metadata"`
	Description      string            `json:"description"`
	IsUsingAutoscale bool              `json:"is_using_autoscale"`
	UpdatedAt        string            `json:"updated_at"`
	CreateAt         string            `json:"created_at"`
	FromVolume       Volume            `json:"volume"`
	Category         string            `json:"category"`
}

type snapshot struct {
	client *Client
}

// Get gets a snapshot
func (s *snapshot) Get(ctx context.Context, id string) (*Snapshot, error) {
	var snapshot *Snapshot
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, strings.Join([]string{snapshotPath, id}, "/"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// Delete deletes a snapshot
func (s *snapshot) Delete(ctx context.Context, id string) error {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, serverServiceName, strings.Join([]string{snapshotPath, id}, "/"), nil)

	if err != nil {
		return err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		fmt.Println("error send req")
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

// Create creates a new snapshot
func (s *snapshot) Create(ctx context.Context, scr *SnapshotCreateRequest) (*Snapshot, error) {
	var snapshot *Snapshot
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, snapshotPath, &scr)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// List lists all snapshot of user
func (s *snapshot) List(ctx context.Context, opts *ListOptions) ([]*Snapshot, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, snapshotPath, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var snapshots []*Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshots); err != nil {
		return nil, err
	}

	return snapshots, nil
}
