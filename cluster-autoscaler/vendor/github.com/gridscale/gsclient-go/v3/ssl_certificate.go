package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// SSLCertificateOperator provides an interface for operations on SSL certificates.
type SSLCertificateOperator interface {
	GetSSLCertificateList(ctx context.Context) ([]SSLCertificate, error)
	GetSSLCertificate(ctx context.Context, id string) (SSLCertificate, error)
	CreateSSLCertificate(ctx context.Context, body SSLCertificateCreateRequest) (CreateResponse, error)
	DeleteSSLCertificate(ctx context.Context, id string) error
}

// SSLCertificateList holds a list of SSL certificates.
type SSLCertificateList struct {
	// Array of SSL certificates.
	List map[string]SSLCertificateProperties `json:"certificates"`
}

// SSLCertificate represents a single SSL certificate.
type SSLCertificate struct {
	// Properties of a SSL certificate.
	Properties SSLCertificateProperties `json:"certificate"`
}

// SSLCertificateProperties holds properties of a SSL certificate.
// A SSL certificate can be retrieved and linked to a loadbalancer.
type SSLCertificateProperties struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The common domain name of the SSL certificate.
	CommonName string `json:"common_name"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// Defines the date after which the certificate does not valid.
	NotValidAfter GSTime `json:"not_valid_after"`

	// Defines a list of unique identifiers generated from the MD5, SHA-1, and SHA-256 fingerprints of the certificate.
	Fingerprints FingerprintProperties `json:"fingerprints"`

	// List of labels.
	Labels []string `json:"labels"`
}

// FingerprintProperties holds properties of a list unique identifiers generated from the MD5, SHA-1, and SHA-256 fingerprints of the certificate.
type FingerprintProperties struct {
	// A unique identifier generated from the MD5 fingerprint of the certificate.
	MD5 string `json:"md5"`

	// A unique identifier generated from the SHA1 fingerprint of the certificate.
	SHA1 string `json:"sha1"`

	// A unique identifier generated from the SHA256 fingerprint of the certificate.
	SHA256 string `json:"sha256"`
}

// SSLCertificateCreateRequest represent a payload of a request for creating a SSL certificate.
type SSLCertificateCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The PEM-formatted private-key of the SSL certificate.
	PrivateKey string `json:"private_key"`

	// The PEM-formatted public SSL of the SSL certificate.
	LeafCertificate string `json:"leaf_certificate"`

	// The PEM-formatted full-chain between the certificate authority and the domain's SSL certificate.
	CertificateChain string `json:"certificate_chain,omitempty"`

	// List of labels.
	Labels []string `json:"labels,omitempty"`
}

// GetSSLCertificateList gets the list of available SSL certificates in the project.
//
// See: https://gridscale.io/en/api-documentation/index.html#operation/getCertificates
func (c *Client) GetSSLCertificateList(ctx context.Context) ([]SSLCertificate, error) {
	r := gsRequest{
		uri:                 apiSSLCertificateBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}

	var response SSLCertificateList
	var sslCerts []SSLCertificate
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		sslCerts = append(sslCerts, SSLCertificate{Properties: properties})
	}
	return sslCerts, err
}

// GetSSLCertificate gets a single SSL certificate.
//
// See: https://gridscale.io/en/api-documentation/index.html#operation/getCertificate
func (c *Client) GetSSLCertificate(ctx context.Context, id string) (SSLCertificate, error) {
	if !isValidUUID(id) {
		return SSLCertificate{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiSSLCertificateBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response SSLCertificate
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateSSLCertificate creates a new SSL certificate.
//
// See: https://gridscale.io/en/api-documentation/index.html#operation/createCertificate
func (c *Client) CreateSSLCertificate(ctx context.Context, body SSLCertificateCreateRequest) (CreateResponse, error) {
	r := gsRequest{
		uri:    apiSSLCertificateBase,
		method: http.MethodPost,
		body:   body,
	}
	var response CreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// DeleteSSLCertificate removes a single SSL certificate.
//
// See: https://gridscale.io/en/api-documentation/index.html#operation/deleteCertificate
func (c *Client) DeleteSSLCertificate(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiSSLCertificateBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}
