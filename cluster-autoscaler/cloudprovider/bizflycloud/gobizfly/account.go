// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	regionsPath   = "/regions"
	usersInfoPath = "/users/info"
)

type accountService struct {
	client *Client
}

var _ AccountService = (*accountService)(nil)

type AccountService interface {
	ListRegion(ctx context.Context) (*Regions, error)
	GetRegion(ctx context.Context, regionName string) (*Region, error)
	GetUserInfo(ctx context.Context) (*User, error)
}

type Region struct {
	Active     bool               `json:"active"`
	Icon       string             `json:"icon"`
	Name       string             `json:"name"`
	Order      int                `json:"order"`
	RegionName string             `json:"region_name"`
	ShortName  string             `json:"short_name"`
	Zones      []AvailabilityZone `json:"zones"`
}

type AvailabilityZone struct {
	Active    bool   `json:"active"`
	Icon      string `json:"icon"`
	Name      string `json:"name"`
	Order     int    `json:"order"`
	ShortName string `json:"short_name"`
}

type Regions struct {
	HN  Region `json:"HN"`
	HCM Region `json:"HCM"`
}

type User struct {
	Service           string             `json:"service"`
	URLType           string             `json:"url_type"`
	Origin            string             `json:"origin"`
	ClientType        string             `json:"client_type"`
	BillingBalance    int                `json:"billing_balance"`
	Balances          map[string]float32 `json:"balances"`
	PaymentMethod     string             `json:"payment_method"`
	BillingAccID      string             `json:"billing_acc_id"`
	Debit             bool               `json:"debit"`
	Email             string             `json:"email"`
	Phone             string             `json:"phone"`
	FullName          string             `json:"full_name"`
	VerifiedEmail     bool               `json:"verified_email"`
	VerifiedPhone     bool               `json:"verified_phone"`
	LoginAlert        bool               `json:"login_alert"`
	VerifiedPayment   bool               `json:"verified_payment"`
	LastRegion        string             `json:"last_region"`
	LastProject       string             `json:"last_project"`
	Type              string             `json:"type"`
	OTP               bool               `json:"otp"`
	Services          []Service          `json:"services"`
	Whitelist         []string           `json:"whitelist"`
	Expires           string             `json:"expires"`
	TenantID          string             `json:"tenant_id"`
	TenantName        string             `json:"tenant_name"`
	KsUserID          string             `json:"ks_user_id"`
	IAM               IAM                `json:"iam"`
	Domains           []string           `json:"domains"`
	PaymentType       string             `json:"payment_type"`
	DOB               string             `json:"dob"`
	Gender            string             `json:"_gender"`
	Trial             Trial              `json:"trial"`
	HasExpiredInvoice bool               `json:"has_expired_invoice"`
	NegativeBalance   bool               `json:"negative_balance"`
	Promotion         []string           `json:"promotion"`
}

type IAM struct {
	Expire          string `json:"expire"`
	TenantID        string `json:"tenant_id"`
	TenantName      string `json:"tenant_name"`
	VerifiedPhone   bool   `json:"verified_phone"`
	VerifiedEmail   bool   `json:"verified_email"`
	VerifiedPayment bool   `json:"verified_payment"`
}

type Trial struct {
	StartedAt    string `json:"started_at"`
	ExpiredAt    string `json:"expired_at"`
	Active       bool   `json:"active"`
	Enable       bool   `json:"enable"`
	ServiceLevel int    `json:"service_level"`
}

func (a accountService) resourceRegionPath() string {
	return regionsPath
}

func (a accountService) resourceUserInfo() string {
	return usersInfoPath
}

func (a accountService) itemPath(name string) string {
	return strings.Join([]string{regionsPath, name}, "/")
}

func (a accountService) ListRegion(ctx context.Context) (*Regions, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, accountName, a.resourceRegionPath(), nil)
	if err != nil {
		return nil, err
	}
	var regions *Regions
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&regions); err != nil {
		return nil, err
	}
	return regions, nil
}

func (a accountService) GetRegion(ctx context.Context, regionName string) (*Region, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, accountName, a.itemPath(regionName), nil)
	if err != nil {
		return nil, err
	}
	var region *Region
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&region); err != nil {
		return nil, err
	}
	return region, nil
}

func (a accountService) GetUserInfo(ctx context.Context) (*User, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, accountName, a.resourceUserInfo(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var user *User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}
