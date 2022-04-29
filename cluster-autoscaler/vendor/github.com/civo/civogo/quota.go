package civogo

import (
	"bytes"
	"encoding/json"
)

// Quota represents the available limits and usage for an account's Civo quota
type Quota struct {
	ID                      string `json:"id"`
	DefaultUserID           string `json:"default_user_id"`
	DefaultUserEmailAddress string `json:"default_user_email_address"`
	InstanceCountLimit      int    `json:"instance_count_limit"`
	InstanceCountUsage      int    `json:"instance_count_usage"`
	CPUCoreLimit            int    `json:"cpu_core_limit"`
	CPUCoreUsage            int    `json:"cpu_core_usage"`
	RAMMegabytesLimit       int    `json:"ram_mb_limit"`
	RAMMegabytesUsage       int    `json:"ram_mb_usage"`
	DiskGigabytesLimit      int    `json:"disk_gb_limit"`
	DiskGigabytesUsage      int    `json:"disk_gb_usage"`
	DiskVolumeCountLimit    int    `json:"disk_volume_count_limit"`
	DiskVolumeCountUsage    int    `json:"disk_volume_count_usage"`
	DiskSnapshotCountLimit  int    `json:"disk_snapshot_count_limit"`
	DiskSnapshotCountUsage  int    `json:"disk_snapshot_count_usage"`
	PublicIPAddressLimit    int    `json:"public_ip_address_limit"`
	PublicIPAddressUsage    int    `json:"public_ip_address_usage"`
	SubnetCountLimit        int    `json:"subnet_count_limit"`
	SubnetCountUsage        int    `json:"subnet_count_usage"`
	NetworkCountLimit       int    `json:"network_count_limit"`
	NetworkCountUsage       int    `json:"network_count_usage"`
	SecurityGroupLimit      int    `json:"security_group_limit"`
	SecurityGroupUsage      int    `json:"security_group_usage"`
	SecurityGroupRuleLimit  int    `json:"security_group_rule_limit"`
	SecurityGroupRuleUsage  int    `json:"security_group_rule_usage"`
	PortCountLimit          int    `json:"port_count_limit"`
	PortCountUsage          int    `json:"port_count_usage"`
	LoadBalancerCountLimit  int    `json:"loadbalancer_count_limit"`
	LoadBalancerCountUsage  int    `json:"loadbalancer_count_usage"`
}

// GetQuota returns all load balancers owned by the calling API account
func (c *Client) GetQuota() (*Quota, error) {
	resp, err := c.SendGetRequest("/v2/quota")
	if err != nil {
		return nil, decodeError(err)
	}

	var quota Quota
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&quota); err != nil {
		return nil, err
	}

	return &quota, nil
}
