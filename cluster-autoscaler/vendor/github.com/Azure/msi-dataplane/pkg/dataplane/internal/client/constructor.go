package client

import "github.com/Azure/azure-sdk-for-go/sdk/azcore"

// NewManagedIdentityDataPlaneAPIClient creates a new MSI data-plane client.
func NewManagedIdentityDataPlaneAPIClient(delegate *azcore.Client) *ManagedIdentityDataPlaneAPIClient {
	return &ManagedIdentityDataPlaneAPIClient{internal: delegate}
}
