/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import "time"

// Certificate defines the schema of an certificate.
type Certificate struct {
	ID             int               `json:"id"`
	Name           string            `json:"name"`
	Labels         map[string]string `json:"labels"`
	Certificate    string            `json:"certificate"`
	Created        time.Time         `json:"created"`
	NotValidBefore time.Time         `json:"not_valid_before"`
	NotValidAfter  time.Time         `json:"not_valid_after"`
	DomainNames    []string          `json:"domain_names"`
	Fingerprint    string            `json:"fingerprint"`
}

// CertificateListResponse defines the schema of the response when
// listing Certificates.
type CertificateListResponse struct {
	Certificates []Certificate `json:"certificates"`
}

// CertificateGetResponse defines the schema of the response when
// retrieving a single Certificate.
type CertificateGetResponse struct {
	Certificate Certificate `json:"certificate"`
}

// CertificateCreateRequest defines the schema of the request to create a certificate.
type CertificateCreateRequest struct {
	Name        string             `json:"name"`
	Certificate string             `json:"certificate"`
	PrivateKey  string             `json:"private_key"`
	Labels      *map[string]string `json:"labels,omitempty"`
}

// CertificateCreateResponse defines the schema of the response when creating a certificate.
type CertificateCreateResponse struct {
	Certificate Certificate `json:"certificate"`
}

// CertificateUpdateRequest defines the schema of the request to update a certificate.
type CertificateUpdateRequest struct {
	Name   *string            `json:"name,omitempty"`
	Labels *map[string]string `json:"labels,omitempty"`
}

// CertificateUpdateResponse defines the schema of the response when updating a certificate.
type CertificateUpdateResponse struct {
	Certificate Certificate `json:"certificate"`
}
