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

// DatabaseBackupConfig represents a Database Backup configuration.
type DatabaseBackupConfig struct {
	FrequentIntervalMinutes    *int64
	FrequentOldestAgeMinutes   *int64
	InfrequentIntervalMinutes  *int64
	InfrequentOldestAgeMinutes *int64
	Interval                   *int64
	MaxCount                   *int64
	RecoveryMode               *string
}

func databaseBackupConfigFromAPI(c *oapi.DbaasBackupConfig) *DatabaseBackupConfig {
	return &DatabaseBackupConfig{
		FrequentIntervalMinutes:    c.FrequentIntervalMinutes,
		FrequentOldestAgeMinutes:   c.FrequentOldestAgeMinutes,
		InfrequentIntervalMinutes:  c.InfrequentIntervalMinutes,
		InfrequentOldestAgeMinutes: c.InfrequentOldestAgeMinutes,
		Interval:                   c.Interval,
		MaxCount:                   c.MaxCount,
		RecoveryMode:               c.RecoveryMode,
	}
}

// DatabasePlan represents a Database Plan.
type DatabasePlan struct {
	Authorized       *bool
	BackupConfig     *DatabaseBackupConfig
	DiskSpace        *int64
	MaxMemoryPercent *int64
	Name             *string
	Nodes            *int64
	NodeCPUs         *int64
	NodeMemory       *int64
}

func databasePlanFromAPI(p *oapi.DbaasPlan) *DatabasePlan {
	return &DatabasePlan{
		Authorized:       p.Authorized,
		BackupConfig:     databaseBackupConfigFromAPI(p.BackupConfig),
		DiskSpace:        p.DiskSpace,
		MaxMemoryPercent: p.MaxMemoryPercent,
		Name:             p.Name,
		Nodes:            p.NodeCount,
		NodeCPUs:         p.NodeCpuCount,
		NodeMemory:       p.NodeMemory,
	}
}

// DatabaseServiceType represents a Database Service type.
type DatabaseServiceType struct {
	AvailableVersions *[]string
	DefaultVersion    *string
	Description       *string
	Name              *string
	Plans             []*DatabasePlan
}

func databaseServiceTypeFromAPI(t *oapi.DbaasServiceType) *DatabaseServiceType {
	return &DatabaseServiceType{
		AvailableVersions: t.AvailableVersions,
		DefaultVersion:    t.DefaultVersion,
		Description:       t.Description,
		Name:              (*string)(t.Name),
		Plans: func() []*DatabasePlan {
			plans := make([]*DatabasePlan, 0)
			if t.Plans != nil {
				for _, plan := range *t.Plans {
					plan := plan
					plans = append(plans, databasePlanFromAPI(&plan))
				}
			}
			return plans
		}(),
	}
}

// DatabaseServiceNotification represents a Database Service notification.
type DatabaseServiceNotification struct {
	Level   string
	Message string
	Type    string
}

func databaseServiceNotificationFromAPI(n *oapi.DbaasServiceNotification) *DatabaseServiceNotification {
	return &DatabaseServiceNotification{
		Level:   string(n.Level),
		Message: n.Message,
		Type:    string(n.Type),
	}
}

// DatabaseServiceComponent represents a Database Service component.
type DatabaseServiceComponent struct {
	Name *string
	Info map[string]interface{}
}

// DatabaseService represents a Database Service.
type DatabaseService struct {
	CreatedAt             *time.Time
	DiskSize              *int64
	Name                  *string `req-for:"delete"`
	Nodes                 *int64
	NodeCPUs              *int64
	NodeMemory            *int64
	Notifications         []*DatabaseServiceNotification
	Plan                  *string
	State                 *string
	TerminationProtection *bool
	Type                  *string
	UpdatedAt             *time.Time
	Zone                  *string
}

func databaseServiceFromAPI(s *oapi.DbaasServiceCommon, zone string) *DatabaseService {
	return &DatabaseService{
		CreatedAt:  s.CreatedAt,
		DiskSize:   s.DiskSize,
		Name:       (*string)(&s.Name),
		Nodes:      s.NodeCount,
		NodeCPUs:   s.NodeCpuCount,
		NodeMemory: s.NodeMemory,
		Notifications: func() []*DatabaseServiceNotification {
			notifications := make([]*DatabaseServiceNotification, 0)
			if s.Notifications != nil {
				for _, n := range *s.Notifications {
					notifications = append(notifications, databaseServiceNotificationFromAPI(&n))
				}
			}
			return notifications
		}(),
		Plan:                  &s.Plan,
		State:                 (*string)(s.State),
		TerminationProtection: s.TerminationProtection,
		Type:                  (*string)(&s.Type),
		UpdatedAt:             s.UpdatedAt,
		Zone:                  &zone,
	}
}

// DeleteDatabaseService deletes a Database Service.
func (c *Client) DeleteDatabaseService(ctx context.Context, zone string, databaseService *DatabaseService) error {
	if err := validateOperationParams(databaseService, "delete"); err != nil {
		return err
	}

	_, err := c.DeleteDbaasServiceWithResponse(apiv2.WithZone(ctx, zone), *databaseService.Name)
	if err != nil {
		return err
	}

	return nil
}

// GetDatabaseCACertificate returns the CA certificate required to access Database Services using a TLS connection.
func (c *Client) GetDatabaseCACertificate(ctx context.Context, zone string) (string, error) {
	resp, err := c.GetDbaasCaCertificateWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return "", err
	}

	return *resp.JSON200.Certificate, nil
}

// GetDatabaseServiceType returns the Database Service type corresponding to the specified name.
func (c *Client) GetDatabaseServiceType(ctx context.Context, zone, name string) (*DatabaseServiceType, error) {
	resp, err := c.GetDbaasServiceTypeWithResponse(apiv2.WithZone(ctx, zone), name)
	if err != nil {
		return nil, err
	}

	return databaseServiceTypeFromAPI(resp.JSON200), nil
}

// ListDatabaseServiceTypes returns the list of existing Database Service types.
func (c *Client) ListDatabaseServiceTypes(ctx context.Context, zone string) ([]*DatabaseServiceType, error) {
	list := make([]*DatabaseServiceType, 0)

	resp, err := c.ListDbaasServiceTypesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DbaasServiceTypes != nil {
		for i := range *resp.JSON200.DbaasServiceTypes {
			list = append(list, databaseServiceTypeFromAPI(&(*resp.JSON200.DbaasServiceTypes)[i]))
		}
	}

	return list, nil
}

// ListDatabaseServices returns the list of Database Services.
func (c *Client) ListDatabaseServices(ctx context.Context, zone string) ([]*DatabaseService, error) {
	list := make([]*DatabaseService, 0)

	resp, err := c.ListDbaasServicesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DbaasServices != nil {
		for i := range *resp.JSON200.DbaasServices {
			list = append(list, databaseServiceFromAPI(&(*resp.JSON200.DbaasServices)[i], zone))
		}
	}

	return list, nil
}

type DatabaseMigrationStatusDetailsStatus string

const (
	DatabaseMigrationStatusDone    DatabaseMigrationStatusDetailsStatus = "done"
	DatabaseMigrationStatusFailed  DatabaseMigrationStatusDetailsStatus = "failed"
	DatabaseMigrationStatusRunning DatabaseMigrationStatusDetailsStatus = "running"
	DatabaseMigrationStatusSyncing DatabaseMigrationStatusDetailsStatus = "syncing"
)

type DatabaseMigrationRedisMasterLinkStatus string

// Defines values for MasterLinkStatus.
const (
	MasterLinkStatusDown DatabaseMigrationRedisMasterLinkStatus = "down"

	MasterLinkStatusUp DatabaseMigrationRedisMasterLinkStatus = "up"
)

type DatabaseMigrationStatusDetails struct {
	// Migrated db name (PG) or number (Redis)
	DBName *string

	// Error message in case that migration has failed
	Error *string

	// Migration method
	Method *string
	Status *DatabaseMigrationStatusDetailsStatus
}

type DatabaseMigrationStatus struct {
	// Migration status per database
	Details []DatabaseMigrationStatusDetails

	// Error message in case that migration has failed
	Error *string

	// Redis only: how many seconds since last I/O with redis master
	MasterLastIOSecondsAgo *int64
	MasterLinkStatus       *DatabaseMigrationRedisMasterLinkStatus

	// Migration method. Empty in case of multiple methods or error
	Method *string

	// Migration status
	Status *string
}

func databaseMigrationStatusFromAPI(in *oapi.DbaasMigrationStatus) *DatabaseMigrationStatus {
	if in == nil {
		return nil
	}

	out := &DatabaseMigrationStatus{
		Details:                []DatabaseMigrationStatusDetails{},
		Error:                  in.Error,
		MasterLastIOSecondsAgo: in.MasterLastIoSecondsAgo,
		MasterLinkStatus:       (*DatabaseMigrationRedisMasterLinkStatus)(in.MasterLinkStatus),
		Method:                 in.Method,
		Status:                 in.Status,
	}

	if in.Details != nil {
		for _, d := range *in.Details {
			out.Details = append(out.Details, DatabaseMigrationStatusDetails{
				DBName: d.Dbname,
				Error:  d.Error,
				Method: d.Method,
				Status: (*DatabaseMigrationStatusDetailsStatus)(d.Status),
			})
		}
	}

	return out
}

func (c *Client) GetDatabaseMigrationStatus(ctx context.Context, zone string, name string) (*DatabaseMigrationStatus, error) {
	resp, err := c.GetDbaasMigrationStatusWithResponse(apiv2.WithZone(ctx, zone), oapi.DbaasServiceName(name))
	if err != nil {
		return nil, err
	}

	return databaseMigrationStatusFromAPI(resp.JSON200), nil
}
