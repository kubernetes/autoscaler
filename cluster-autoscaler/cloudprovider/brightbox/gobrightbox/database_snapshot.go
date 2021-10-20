package gobrightbox

import (
	"time"
)

// DatabaseSnapshot represents a snapshot of a database server.
// https://api.gb1.brightbox.com/1.0/#database_snapshot
type DatabaseSnapshot struct {
	Id              string
	Name            string
	Description     string
	Status          string
	Account         Account
	DatabaseEngine  string `json:"database_engine"`
	DatabaseVersion string `json:"database_version"`
	Size            int
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	Locked          bool
}

// DatabaseSnapshots retrieves a list of all database snapshot
func (c *Client) DatabaseSnapshots() ([]DatabaseSnapshot, error) {
	var database_snapshot []DatabaseSnapshot
	_, err := c.MakeApiRequest("GET", "/1.0/database_snapshots", nil, &database_snapshot)
	if err != nil {
		return nil, err
	}
	return database_snapshot, err
}

// DatabaseSnapshot retrieves a detailed view of one database snapshot
func (c *Client) DatabaseSnapshot(identifier string) (*DatabaseSnapshot, error) {
	database_snapshot := new(DatabaseSnapshot)
	_, err := c.MakeApiRequest("GET", "/1.0/database_snapshots/"+identifier, nil, database_snapshot)
	if err != nil {
		return nil, err
	}
	return database_snapshot, err
}

// DestroyDatabaseSnapshot issues a request to destroy the database snapshot
func (c *Client) DestroyDatabaseSnapshot(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/database_snapshots/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
