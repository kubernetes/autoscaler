package gobrightbox

import (
	"time"
)

// FirewallPolicy represents a firewall policy.
// https://api.gb1.brightbox.com/1.0/#firewall_policy
type FirewallPolicy struct {
	Id          string
	Name        string
	Default     bool
	CreatedAt   time.Time `json:"created_at"`
	Description string
	ServerGroup *ServerGroup   `json:"server_group"`
	Rules       []FirewallRule `json:"rules"`
}

// FirewallPolicyOptions is used in conjunction with CreateFirewallPolicy and
// UpdateFirewallPolicy to create and update firewall policies.
type FirewallPolicyOptions struct {
	Id          string  `json:"-"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ServerGroup *string `json:"server_group,omitempty"`
}

// FirewallPolicies retrieves a list of all firewall policies
func (c *Client) FirewallPolicies() ([]FirewallPolicy, error) {
	var policies []FirewallPolicy
	_, err := c.MakeApiRequest("GET", "/1.0/firewall_policies", nil, &policies)
	if err != nil {
		return nil, err
	}
	return policies, err
}

// FirewallPolicy retrieves a detailed view of one firewall policy
func (c *Client) FirewallPolicy(identifier string) (*FirewallPolicy, error) {
	policy := new(FirewallPolicy)
	_, err := c.MakeApiRequest("GET", "/1.0/firewall_policies/"+identifier, nil, policy)
	if err != nil {
		return nil, err
	}
	return policy, err
}

// CreateFirewallPolicy creates a new firewall policy.
//
// It takes a FirewallPolicyOptions struct for specifying name and other
// attributes. Not all attributes can be specified at create time (such as Id,
// which is allocated for you)
func (c *Client) CreateFirewallPolicy(policyOptions *FirewallPolicyOptions) (*FirewallPolicy, error) {
	policy := new(FirewallPolicy)
	_, err := c.MakeApiRequest("POST", "/1.0/firewall_policies", policyOptions, &policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UpdateFirewallPolicy updates an existing firewall policy.
//
// It takes a FirewallPolicyOptions struct for specifying name and other
// attributes. Not all attributes can be update(such as server_group which is
// instead changed with ApplyFirewallPolicy).
//
// Specify the policy you want to update using the Id field
func (c *Client) UpdateFirewallPolicy(policyOptions *FirewallPolicyOptions) (*FirewallPolicy, error) {
	policy := new(FirewallPolicy)
	_, err := c.MakeApiRequest("PUT", "/1.0/firewall_policies/"+policyOptions.Id, policyOptions, &policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// DestroyFirewallPolicy issues a request to destroy the firewall policy
func (c *Client) DestroyFirewallPolicy(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/firewall_policies/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// ApplyFirewallPolicy issues a request to apply the given firewall policy to
// the given server group.
func (c *Client) ApplyFirewallPolicy(policyId string, serverGroupId string) (*FirewallPolicy, error) {
	policy := new(FirewallPolicy)
	_, err := c.MakeApiRequest("POST", "/1.0/firewall_policies/"+policyId+"/apply_to",
		map[string]string{"server_group": serverGroupId}, &policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// RemoveFirewallPolicy issues a request to remove the given firewall policy from
// the given server group.
func (c *Client) RemoveFirewallPolicy(policyId string, serverGroupId string) (*FirewallPolicy, error) {
	policy := new(FirewallPolicy)
	_, err := c.MakeApiRequest("POST", "/1.0/firewall_policies/"+policyId+"/remove",
		map[string]string{"server_group": serverGroupId}, &policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}
