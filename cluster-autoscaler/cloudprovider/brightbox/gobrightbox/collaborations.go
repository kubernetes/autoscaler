package gobrightbox

import (
	"time"
)

// Collaboration represents a User's links to it's Accounts
// https://api.gb1.brightbox.com/1.0/#user
type Collaboration struct {
	Id         string
	Email      string
	Role       string
	RoleLabel  string `json:"role_label"`
	Status     string
	CreatedAt  *time.Time `json:"created_at"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Account    Account
	User       User
	Inviter    User
}

// Collaborations retrieves a list of all the current user's collaborations
func (c *Client) Collaborations() ([]Collaboration, error) {
	var cl []Collaboration
	_, err := c.MakeApiRequest("GET", "/1.0/user/collaborations", nil, &cl)
	if err != nil {
		return nil, err
	}
	return cl, err
}

// Collaboration retrieves a detailed view of one of the current user's
// collaborations
func (c *Client) Collaboration(identifier string) (*Collaboration, error) {
	col := new(Collaboration)
	_, err := c.MakeApiRequest("GET", "/1.0/user/collaborations/"+identifier, nil, col)
	if err != nil {
		return nil, err
	}
	return col, err
}
