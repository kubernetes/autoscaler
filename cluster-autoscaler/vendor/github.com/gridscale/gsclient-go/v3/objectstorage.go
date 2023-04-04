package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"
)

// ObjectStorageOperator provides an interface for operations on object storages.
type ObjectStorageOperator interface {
	GetObjectStorageAccessKeyList(ctx context.Context) ([]ObjectStorageAccessKey, error)
	GetObjectStorageAccessKey(ctx context.Context, id string) (ObjectStorageAccessKey, error)
	CreateObjectStorageAccessKey(ctx context.Context) (ObjectStorageAccessKeyCreateResponse, error)
	DeleteObjectStorageAccessKey(ctx context.Context, id string) error
	GetObjectStorageBucketList(ctx context.Context) ([]ObjectStorageBucket, error)
}

// ObjectStorageAccessKeyList holds a list of object storage access keys.
type ObjectStorageAccessKeyList struct {
	// Array of Object Storages' access keys.
	List []ObjectStorageAccessKeyProperties `json:"access_keys"`
}

// ObjectStorageAccessKey represents a single object storage access key.
type ObjectStorageAccessKey struct {
	// Properties of an object storage access key.
	Properties ObjectStorageAccessKeyProperties `json:"access_key"`
}

// ObjectStorageAccessKeyProperties holds properties of an object storage access key.
type ObjectStorageAccessKeyProperties struct {
	// The object storage secret_key.
	SecretKey string `json:"secret_key"`

	// The object storage access_key.
	AccessKey string `json:"access_key"`

	// Account this credentials belong to.
	User string `json:"user"`
}

// ObjectStorageAccessKeyCreateResponse represents a response for creating an object storage access key.
type ObjectStorageAccessKeyCreateResponse struct {
	AccessKey struct {
		////The object storage secret_key.
		SecretKey string `json:"secret_key"`

		// The object storage secret_key.
		AccessKey string `json:"access_key"`
	} `json:"access_key"`

	// UUID of the request.
	RequestUUID string `json:"request_uuid"`
}

// ObjectStorageBucketList holds a list of buckets.
type ObjectStorageBucketList struct {
	// Array of Buckets
	List []ObjectStorageBucketProperties `json:"buckets"`
}

// ObjectStorageBucket represents a single bucket.
type ObjectStorageBucket struct {
	// Properties of a bucket.
	Properties ObjectStorageBucketProperties `json:"bucket"`
}

// ObjectStorageBucketProperties holds properties of a bucket.
type ObjectStorageBucketProperties struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The current usage of the bucket.
	Usage struct {
		// The size of the the bucket (in kb).
		SizeKb int `json:"size_kb"`

		// The number of files in the bucket.
		NumObjects int `json:"num_objects"`
	} `json:"usage"`
}

// GetObjectStorageAccessKeyList gets a list of available object storage access keys.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getAccessKeys
func (c *Client) GetObjectStorageAccessKeyList(ctx context.Context) ([]ObjectStorageAccessKey, error) {
	r := gsRequest{
		uri:                 path.Join(apiObjectStorageBase, "access_keys"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ObjectStorageAccessKeyList
	var accessKeys []ObjectStorageAccessKey
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		accessKeys = append(accessKeys, ObjectStorageAccessKey{Properties: properties})
	}
	return accessKeys, err
}

// GetObjectStorageAccessKey gets a specific object storage access key based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getAccessKey
func (c *Client) GetObjectStorageAccessKey(ctx context.Context, id string) (ObjectStorageAccessKey, error) {
	if strings.TrimSpace(id) == "" {
		return ObjectStorageAccessKey{}, errors.New("'id' is required")
	}
	r := gsRequest{
		uri:                 path.Join(apiObjectStorageBase, "access_keys", id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ObjectStorageAccessKey
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateObjectStorageAccessKey creates an object storage access key.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createAccessKey
func (c *Client) CreateObjectStorageAccessKey(ctx context.Context) (ObjectStorageAccessKeyCreateResponse, error) {
	r := gsRequest{
		uri:    path.Join(apiObjectStorageBase, "access_keys"),
		method: http.MethodPost,
	}
	var response ObjectStorageAccessKeyCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// DeleteObjectStorageAccessKey removed a specific object storage access key based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteAccessKey
func (c *Client) DeleteObjectStorageAccessKey(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("'id' is required")
	}
	r := gsRequest{
		uri:    path.Join(apiObjectStorageBase, "access_keys", id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetObjectStorageBucketList gets a list of object storage buckets.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getBuckets
func (c *Client) GetObjectStorageBucketList(ctx context.Context) ([]ObjectStorageBucket, error) {
	r := gsRequest{
		uri:                 path.Join(apiObjectStorageBase, "buckets"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ObjectStorageBucketList
	var buckets []ObjectStorageBucket
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		buckets = append(buckets, ObjectStorageBucket{Properties: properties})
	}
	return buckets, err
}
