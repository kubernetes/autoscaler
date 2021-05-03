// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	sshKeyBasePath = "/keypairs"
)

var _ SSHKeyService = (*sshkey)(nil)

type sshkey struct {
	client *Client
}

// SSHKeyService is an interface to interact with BizFly API SSH Key
type SSHKeyService interface {
	List(ctx context.Context, opts *ListOptions) ([]*KeyPair, error)
	Create(ctx context.Context, scr *SSHKeyCreateRequest) (*SSHKeyCreateResponse, error)
	Delete(ctx context.Context, keyname string) (*SSHKeyDeleteResponse, error)
}

type SSHKey struct {
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	FingerPrint string `json:"fingerprint"`
}

type KeyPair struct {
	SSHKeyPair SSHKey `json:"keypair"`
}

type SSHKeyCreateResponse struct {
	SSHKey
	UserID string `json:"user_id"`
}

type SSHKeyCreateRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type SSHKeyDeleteResponse struct {
	Message string `json:"message"`
}

func (s *sshkey) List(ctx context.Context, opts *ListOptions) ([]*KeyPair, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, serverServiceName, sshKeyBasePath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var sshKeys []*KeyPair

	if err := json.NewDecoder(resp.Body).Decode(&sshKeys); err != nil {
		return nil, err
	}
	return sshKeys, nil
}

func (s *sshkey) Create(ctx context.Context, scr *SSHKeyCreateRequest) (*SSHKeyCreateResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, serverServiceName, sshKeyBasePath, &scr)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var sshKey *SSHKeyCreateResponse

	if err := json.NewDecoder(resp.Body).Decode(&sshKey); err != nil {
		return nil, err
	}
	return sshKey, nil
}

func (s *sshkey) Delete(ctx context.Context, keyname string) (*SSHKeyDeleteResponse, error) {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, serverServiceName, sshKeyBasePath+"/"+keyname, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var response *SSHKeyDeleteResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}
