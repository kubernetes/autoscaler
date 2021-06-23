package gobrightbox

import (
	"time"
)

// DatabaseServer represents a database server.
// https://api.gb1.brightbox.com/1.0/#database_server
type DatabaseServer struct {
	Id                      string
	Name                    string
	Description             string
	Status                  string
	Account                 Account
	DatabaseEngine          string             `json:"database_engine"`
	DatabaseVersion         string             `json:"database_version"`
	AdminUsername           string             `json:"admin_username"`
	AdminPassword           string             `json:"admin_password"`
	CreatedAt               *time.Time         `json:"created_at"`
	UpdatedAt               *time.Time         `json:"updated_at"`
	DeletedAt               *time.Time         `json:"deleted_at"`
	SnapshotsScheduleNextAt *time.Time         `json:"snapshots_schedule_next_at"`
	AllowAccess             []string           `json:"allow_access"`
	MaintenanceWeekday      int                `json:"maintenance_weekday"`
	MaintenanceHour         int                `json:"maintenance_hour"`
	SnapshotsSchedule       string             `json:"snapshots_schedule"`
	CloudIPs                []CloudIP          `json:"cloud_ips"`
	DatabaseServerType      DatabaseServerType `json:"database_server_type"`
	Locked                  bool
	Zone                    Zone
}

// DatabaseServerOptions is used in conjunction with CreateDatabaseServer and
// UpdateDatabaseServer to create and update database servers.
type DatabaseServerOptions struct {
	Id                 string   `json:"-"`
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	Engine             string   `json:"engine,omitempty"`
	Version            string   `json:"version,omitempty"`
	AllowAccess        []string `json:"allow_access,omitempty"`
	Snapshot           string   `json:"snapshot,omitempty"`
	Zone               string   `json:"zone,omitempty"`
	DatabaseType       string   `json:"database_type,omitempty"`
	MaintenanceWeekday *int     `json:"maintenance_weekday,omitempty"`
	MaintenanceHour    *int     `json:"maintenance_hour,omitempty"`
	SnapshotsSchedule  *string  `json:"snapshots_schedule,omitempty"`
}

// DatabaseServers retrieves a list of all database servers
func (c *Client) DatabaseServers() ([]DatabaseServer, error) {
	var dbs []DatabaseServer
	_, err := c.MakeApiRequest("GET", "/1.0/database_servers", nil, &dbs)
	if err != nil {
		return nil, err
	}
	return dbs, err
}

// DatabaseServer retrieves a detailed view of one database server
func (c *Client) DatabaseServer(identifier string) (*DatabaseServer, error) {
	dbs := new(DatabaseServer)
	_, err := c.MakeApiRequest("GET", "/1.0/database_servers/"+identifier, nil, dbs)
	if err != nil {
		return nil, err
	}
	return dbs, err
}

// CreateDatabaseServer creates a new database server.
//
// It takes a DatabaseServerOptions struct for specifying name and other
// attributes. Not all attributes can be specified at create time
// (such as Id, which is allocated for you)
func (c *Client) CreateDatabaseServer(options *DatabaseServerOptions) (*DatabaseServer, error) {
	dbs := new(DatabaseServer)
	_, err := c.MakeApiRequest("POST", "/1.0/database_servers", options, &dbs)
	if err != nil {
		return nil, err
	}
	return dbs, nil
}

// UpdateDatabaseServer updates an existing database server.
//
// It takes a DatabaseServerOptions struct for specifying Id, name and other
// attributes. Not all attributes can be specified at update time.
func (c *Client) UpdateDatabaseServer(options *DatabaseServerOptions) (*DatabaseServer, error) {
	dbs := new(DatabaseServer)
	_, err := c.MakeApiRequest("PUT", "/1.0/database_servers/"+options.Id, options, &dbs)
	if err != nil {
		return nil, err
	}
	return dbs, nil
}

// DestroyDatabaseServer issues a request to deletes an existing database server
func (c *Client) DestroyDatabaseServer(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/database_servers/"+identifier, nil, nil)
	return err
}

// SnapshotDatabaseServer requests a snapshot of an existing database server.
func (c *Client) SnapshotDatabaseServer(identifier string) (*DatabaseSnapshot, error) {
	dbs := new(DatabaseServer)
	res, err := c.MakeApiRequest("POST", "/1.0/database_servers/"+identifier+"/snapshot", nil, &dbs)
	if err != nil {
		return nil, err
	}
	snapID := getLinkRel(res.Header.Get("Link"), "dbi", "snapshot")
	if snapID != nil {
		snap := new(DatabaseSnapshot)
		snap.Id = *snapID
		return snap, nil
	}
	return nil, nil
}

// ResetPasswordForDatabaseServer requests a snapshot of an existing database server.
func (c *Client) ResetPasswordForDatabaseServer(identifier string) (*DatabaseServer, error) {
	dbs := new(DatabaseServer)
	_, err := c.MakeApiRequest("POST", "/1.0/database_servers/"+identifier+"/reset_password", nil, &dbs)
	if err != nil {
		return nil, err
	}
	return dbs, nil
}
