package dataplane

import (
	"context"

	"github.com/Azure/msi-dataplane/pkg/dataplane/internal/client"
)

// Client wraps the generated code to smooth over the rough edges from generation, namely:
// - the generated clients incorrectly expose the API version as a parameter, even though there's only one option
// - the generated clients incorrectly expose all sorts of internal logic like the request body, etc, when we just want a clean client
// Ideally we wouldn't need this wrapper, but it's much easier to implement this here than update the generator.

// Client exposes the API for the MSI data plane.
type Client interface {
	// DeleteSystemAssignedIdentity deletes the system-assigned identity for a proxy resource.
	DeleteSystemAssignedIdentity(ctx context.Context) error

	// GetSystemAssignedIdentityCredentials retrieves the credentials for the system-assigned identity associated with the proxy resource.
	GetSystemAssignedIdentityCredentials(ctx context.Context) (*ManagedIdentityCredentials, error)

	// GetUserAssignedIdentitiesCredentials retrieves the credentials for any user-assigned identities associated with the proxy resource.
	GetUserAssignedIdentitiesCredentials(ctx context.Context, request UserAssignedIdentitiesRequest) (*ManagedIdentityCredentials, error)

	// MoveIdentity moves the identity from one resource group into another.
	MoveIdentity(ctx context.Context, request MoveIdentityRequest) (*MoveIdentityResponse, error)
}
type clientAdapter struct {
	hostPath string
	delegate *client.ManagedIdentityDataPlaneAPIClient
}

var _ Client = (*clientAdapter)(nil)

func (c *clientAdapter) DeleteSystemAssignedIdentity(ctx context.Context) error {
	_, err := c.delegate.Deleteidentity(ctx, c.hostPath, nil)
	return err
}

func (c *clientAdapter) GetSystemAssignedIdentityCredentials(ctx context.Context) (*ManagedIdentityCredentials, error) {
	resp, err := c.delegate.Getcred(ctx, c.hostPath, nil)
	return &resp.ManagedIdentityCredentials, err
}

func (c *clientAdapter) GetUserAssignedIdentitiesCredentials(ctx context.Context, request UserAssignedIdentitiesRequest) (*ManagedIdentityCredentials, error) {
	resp, err := c.delegate.Getcreds(ctx, c.hostPath, request, nil)
	return &resp.ManagedIdentityCredentials, err
}

func (c *clientAdapter) MoveIdentity(ctx context.Context, request MoveIdentityRequest) (*MoveIdentityResponse, error) {
	resp, err := c.delegate.Moveidentity(ctx, c.hostPath, request, nil)
	return &resp.MoveIdentityResponse, err
}
