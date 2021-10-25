package gobrightbox

import (
	"time"
)

// FirewallRule represents a firewall rule.
// https://api.gb1.brightbox.com/1.0/#firewall_rule
type FirewallRule struct {
	Id              string
	Source          string         `json:"source"`
	SourcePort      string         `json:"source_port"`
	Destination     string         `json:"destination"`
	DestinationPort string         `json:"destination_port"`
	Protocol        string         `json:"protocol"`
	IcmpTypeName    string         `json:"icmp_type_name"`
	CreatedAt       time.Time      `json:"created_at"`
	Description     string         `json:"description"`
	FirewallPolicy  FirewallPolicy `json:"firewall_policy"`
}

// FirewallRuleOptions is used in conjunction with CreateFirewallRule and
// UpdateFirewallRule to create and update firewall rules.
type FirewallRuleOptions struct {
	Id              string  `json:"-"`
	FirewallPolicy  string  `json:"firewall_policy,omitempty"`
	Protocol        *string `json:"protocol,omitempty"`
	Source          *string `json:"source,omitempty"`
	SourcePort      *string `json:"source_port,omitempty"`
	Destination     *string `json:"destination,omitempty"`
	DestinationPort *string `json:"destination_port,omitempty"`
	IcmpTypeName    *string `json:"icmp_type_name,omitempty"`
	Description     *string `json:"description,omitempty"`
}

// FirewallRule retrieves a detailed view of one firewall rule
func (c *Client) FirewallRule(identifier string) (*FirewallRule, error) {
	rule := new(FirewallRule)
	_, err := c.MakeApiRequest("GET", "/1.0/firewall_rules/"+identifier, nil, rule)
	if err != nil {
		return nil, err
	}
	return rule, err
}

// CreateFirewallRule creates a new firewall rule.
//
// It takes a FirewallRuleOptions struct for specifying name and other
// attributes. Not all attributes can be specified at create time
// (such as Id, which is allocated for you)
func (c *Client) CreateFirewallRule(ruleOptions *FirewallRuleOptions) (*FirewallRule, error) {
	rule := new(FirewallRule)
	_, err := c.MakeApiRequest("POST", "/1.0/firewall_rules", ruleOptions, &rule)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// UpdateFirewallRule updates an existing firewall rule.
//
// It takes a FirewallRuleOptions struct for specifying the attributes. Not all
// attributes can be updated (such as firewall_policy)
func (c *Client) UpdateFirewallRule(ruleOptions *FirewallRuleOptions) (*FirewallRule, error) {
	rule := new(FirewallRule)
	_, err := c.MakeApiRequest("PUT", "/1.0/firewall_rules/"+ruleOptions.Id, ruleOptions, &rule)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// DestroyFirewallRule destroys an existing firewall rule
func (c *Client) DestroyFirewallRule(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/firewall_rules/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
