package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/client"
)

type Client interface {
	// Get performs a GET request to the specified path and returns the response body.
	Get(ctx context.Context, path string) ([]byte, error)
	// GetStream performs a GET request to the specified path and returns the response body reader.
	GetStream(ctx context.Context, path string) (io.ReadCloser, error)
	// Post performs a POST request to the specified path and returns the response body.
	Post(ctx context.Context, path string, body []byte) ([]byte, error)
	// Put performs a PUT request to the specified path and returns the response body.
	Put(ctx context.Context, path string, body []byte) ([]byte, error)
	// Patch performs a PATCH request to the specified path and returns the response body.
	Patch(ctx context.Context, path string, body []byte) ([]byte, error)
	// Delete performs a DELETE request to the specified path and returns the response body.
	Delete(ctx context.Context, path string) ([]byte, error)
	// Do performs an HTTP request using custom request object and returns the response body.
	Do(r *http.Request) ([]byte, error)
	// DoStream performs an HTTP request using custom request object and returns the response body reader.
	DoStream(r *http.Request) (io.ReadCloser, error)
}

type requestable interface {
	RequestURL() string
}

type service interface {
	Cloud
	Account
	Firewall
	Host
	IPAddress
	LoadBalancer
	ServerGroup
	Network
	Tag
	Server
	Storage
	ObjectStorage
	ManagedDatabaseServiceManager
	ManagedDatabaseUserManager
	ManagedDatabaseLogicalDatabaseManager
	Permission
	Kubernetes
	ManagedObjectStorage
	Gateway
	Partner
	AuditLog
	FileStorage
}

var _ service = (*Service)(nil)

// Service represents the API service with context support. The specified client is used to communicate with the API
type Service struct {
	client Client
}

// Get performs a GET request to the specified location with context and stores the result in the value pointed to by v.
func (s *Service) get(ctx context.Context, location string, v any) error {
	res, err := s.client.Get(ctx, location)
	if err != nil {
		return parseJSONServiceError(err)
	}

	if v == nil {
		return nil
	}

	err = json.Unmarshal(res, v)
	if err == nil {
		return nil
	}

	if strings.HasPrefix(err.Error(), "json: cannot unmarshal array") {
		return errors.Join(err, errors.New("get: request parameters might be incorrect, ensure that required fields, such as UUID, are set to valid values"))
	}

	return err
}

// Create performs a POST request to the specified location with context and stores the response in the value pointed to by v.
func (s *Service) create(ctx context.Context, r requestable, v any) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	res, err := s.client.Post(ctx, r.RequestURL(), payload)
	if err != nil {
		return parseJSONServiceError(err)
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(res, v)
}

// Modify performs a PATCH request to the specified location with context and stores the response in the value pointed to by v.
func (s *Service) modify(ctx context.Context, r requestable, v any) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	res, err := s.client.Patch(ctx, r.RequestURL(), payload)
	if err != nil {
		return parseJSONServiceError(err)
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(res, v)
}

// Modify performs a PUT request to the specified location with context and stores the response in the value pointed to by v.
func (s *Service) replace(ctx context.Context, r requestable, v any) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	res, err := s.client.Put(ctx, r.RequestURL(), payload)
	if err != nil {
		return parseJSONServiceError(err)
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(res, v)
}

// Delete performs a DELETE request to the specified location with context
func (s *Service) delete(ctx context.Context, r requestable) error {
	_, err := s.client.Delete(ctx, r.RequestURL())
	if err != nil {
		return parseJSONServiceError(err)
	}
	return nil
}

func New(client Client) *Service {
	return &Service{client}
}

// Parses an error returned from the client into corresponding error type
func parseJSONServiceError(err error) error {
	if clientError, ok := err.(*client.Error); ok {
		prob := &upcloud.Problem{}

		switch clientError.Type {
		case client.ErrorTypeProblem:
			if parseErr := json.Unmarshal(clientError.ResponseBody, prob); parseErr != nil {
				return fmt.Errorf("received malformed client error (%w): %w", parseErr, err)
			}
			return prob
		default:
			ucError := &legacyError{}
			if parseErr := json.Unmarshal(clientError.ResponseBody, ucError); parseErr != nil {
				return fmt.Errorf("received malformed client error (%w): %w", parseErr, err)
			}

			prob.Type = ucError.ErrorCode
			prob.Title = ucError.ErrorMessage
			prob.Status = clientError.ErrorCode
			return prob
		}
	}
	return err
}

// Error represents a legacy error object
// It is still returned by UpCloud API, but it is deprecated and in the future all API endpoint should return json+problem conforming errors
type legacyError struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`

	// HTTP Status code
	Status int `json:"-"`
}

// UnmarshalJSON is a custom unmarshaller that deals with
// deeply embedded values.
func (e *legacyError) UnmarshalJSON(b []byte) error {
	type localError legacyError
	v := struct {
		Error localError `json:"error"`
	}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	*e = legacyError(v.Error)

	return nil
}
