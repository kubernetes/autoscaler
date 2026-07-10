package dataplane

import "github.com/Azure/msi-dataplane/pkg/dataplane/internal/client"

type ManagedIdentityCredentials = client.ManagedIdentityCredentials
type CustomClaims = client.CustomClaims
type DelegatedResource = client.DelegatedResource
type UserAssignedIdentityCredentials = client.UserAssignedIdentityCredentials

type UserAssignedIdentitiesRequest = client.CredRequestDefinition

type MoveIdentityRequest = client.MoveRequestBodyDefinition
type MoveIdentityResponse = client.MoveIdentityResponse
