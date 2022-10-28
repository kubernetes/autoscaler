package trusts

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/pagination"
)

type trustResult struct {
	gophercloud.Result
}

// CreateResult is the response from a Create operation. Call its Extract method
// to interpret it as a Trust.
type CreateResult struct {
	trustResult
}

// DeleteResult is the response from a Delete operation. Call its ExtractErr to
// determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// TrustPage is a single page of Region results.
type TrustPage struct {
	pagination.LinkedPageBase
}

// GetResult is the response from a Get operation. Call its Extract method
// to interpret it as a Trust.
type GetResult struct {
	trustResult
}

// IsEmpty determines whether or not a page of Trusts contains any results.
func (t TrustPage) IsEmpty() (bool, error) {
	roles, err := ExtractTrusts(t)
	return len(roles) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (t TrustPage) NextPageURL() (string, error) {
	var s struct {
		Links struct {
			Next     string `json:"next"`
			Previous string `json:"previous"`
		} `json:"links"`
	}
	err := t.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Links.Next, err
}

// ExtractProjects returns a slice of Trusts contained in a single page of
// results.
func ExtractTrusts(r pagination.Page) ([]Trust, error) {
	var s struct {
		Trusts []Trust `json:"trusts"`
	}
	err := (r.(TrustPage)).ExtractInto(&s)
	return s.Trusts, err
}

// Extract interprets any trust result as a Trust.
func (t trustResult) Extract() (*Trust, error) {
	var s struct {
		Trust *Trust `json:"trust"`
	}
	err := t.ExtractInto(&s)
	return s.Trust, err
}

// Trust represents a delegated authorization request between two
// identities.
type Trust struct {
	ID                 string    `json:"id"`
	Impersonation      bool      `json:"impersonation"`
	TrusteeUserID      string    `json:"trustee_user_id"`
	TrustorUserID      string    `json:"trustor_user_id"`
	RedelegatedTrustID string    `json:"redelegated_trust_id"`
	RedelegationCount  int       `json:"redelegation_count,omitempty"`
	AllowRedelegation  bool      `json:"allow_redelegation,omitempty"`
	ProjectID          string    `json:"project_id,omitempty"`
	RemainingUses      int       `json:"remaining_uses,omitempty"`
	Roles              []Role    `json:"roles,omitempty"`
	DeletedAt          time.Time `json:"deleted_at"`
	ExpiresAt          time.Time `json:"expires_at"`
}

// Role specifies a single role that is granted to a trustee.
type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// TokenExt represents an extension of the base token result.
type TokenExt struct {
	Trust Trust `json:"OS-TRUST:trust"`
}

// RolesPage is a single page of Trust roles results.
type RolesPage struct {
	pagination.LinkedPageBase
}

// IsEmpty determines whether or not a a Page contains any results.
func (r RolesPage) IsEmpty() (bool, error) {
	accessTokenRoles, err := ExtractRoles(r)
	return len(accessTokenRoles) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (r RolesPage) NextPageURL() (string, error) {
	var s struct {
		Links struct {
			Next     string `json:"next"`
			Previous string `json:"previous"`
		} `json:"links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Links.Next, err
}

// ExtractRoles returns a slice of Role contained in a single page of results.
func ExtractRoles(r pagination.Page) ([]Role, error) {
	var s struct {
		Roles []Role `json:"roles"`
	}
	err := (r.(RolesPage)).ExtractInto(&s)
	return s.Roles, err
}

type GetRoleResult struct {
	gophercloud.Result
}

// Extract interprets any GetRoleResult result as an Role.
func (r GetRoleResult) Extract() (*Role, error) {
	var s struct {
		Role *Role `json:"role"`
	}
	err := r.ExtractInto(&s)
	return s.Role, err
}

type CheckRoleResult struct {
	gophercloud.ErrResult
}
