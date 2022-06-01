// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	registryPath = "/_"
)

type containerRegistry struct {
	client *Client
}

var _ ContainerRegistryService = (*containerRegistry)(nil)

type ContainerRegistryService interface {
	List(ctx context.Context, opts *ListOptions) ([]*Repository, error)
	Create(ctx context.Context, crpl *createRepositoryPayload) error
	Delete(ctx context.Context, repositoryName string) error
	GetTags(ctx context.Context, repositoryName string) (*TagRepository, error)
	EditRepo(ctx context.Context, repositoryName string, erpl *editRepositoryPayload) error
	DeleteTag(ctx context.Context, tagName string, repositoryName string) error
	GetTag(ctx context.Context, repositoryName string, tagName string, vulnerabilities string) (*Image, error)
}

type Repository struct {
	Name      string `json:"name"`
	LastPush  string `json:"last_push"`
	Pulls     int    `json:"pulls"`
	Public    bool   `json:"public"`
	CreatedAt string `json:"created_at"`
}

type Repositories struct {
	Repositories []Repository `json:"repositories"`
}

type createRepositoryPayload struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type RepositoryTag struct {
	Name            string `json:"name"`
	Author          string `json:"author"`
	LastUpdated     string `json:"last_updated"`
	CreatedAt       string `json:"created_at"`
	LastScan        string `json:"last_scan"`
	ScanStatus      string `json:"scan_status"`
	Vulnerabilities int    `json:"vulnerabilities"`
	Fixes           int    `json:"fixes"`
}

type editRepositoryPayload struct {
	Public bool `json:"public"`
}

type Vulnerability struct {
	Package     string `json:"package"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Severity    string `json:"severity"`
	FixedBy     string `json:"fixed_by"`
}

type Image struct {
	Repository      Repository      `json:"repository"`
	Tag             RepositoryTag   `json:"tag"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

type TagRepository struct {
	Repository Repository      `json:"repository"`
	Tags       []RepositoryTag `json:"tags"`
}

func (c *containerRegistry) resourcePath() string {
	return registryPath
}

func (c *containerRegistry) itemPath(id string) string {
	return strings.Join([]string{registryPath, id}, "/")
}

func (c *containerRegistry) List(ctx context.Context, opts *ListOptions) ([]*Repository, error) {
	req, err := c.client.NewRequest(ctx, http.MethodGet, containerRegistryName, c.resourcePath(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Repositories []*Repository `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Repositories, nil
}

func (c *containerRegistry) Create(ctx context.Context, crpl *createRepositoryPayload) error {
	req, err := c.client.NewRequest(ctx, http.MethodPost, containerRegistryName, c.resourcePath(), &crpl)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *containerRegistry) Delete(ctx context.Context, repositoryName string) error {
	req, err := c.client.NewRequest(ctx, http.MethodDelete, containerRegistryName, c.itemPath(repositoryName), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	return resp.Body.Close()
}

func (c *containerRegistry) GetTags(ctx context.Context, repositoryName string) (*TagRepository, error) {
	var data *TagRepository
	req, err := c.client.NewRequest(ctx, http.MethodGet, containerRegistryName, strings.Join([]string{registryPath, repositoryName}, "/"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *containerRegistry) EditRepo(ctx context.Context, repositoryName string, erpl *editRepositoryPayload) error {
	req, err := c.client.NewRequest(ctx, http.MethodPatch, containerRegistryName, strings.Join([]string{registryPath, repositoryName}, "/"), erpl)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *containerRegistry) DeleteTag(ctx context.Context, repositoryName string, tagName string) error {
	req, err := c.client.NewRequest(ctx, http.MethodDelete, containerRegistryName, strings.Join([]string{registryPath, repositoryName, "tag", tagName}, "/"), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *containerRegistry) GetTag(ctx context.Context, repositoryName string, tagName string, vulnerabilities string) (*Image, error) {
	var data *Image
	u, _ := url.Parse(strings.Join([]string{registryPath, repositoryName, "tag", tagName}, "/"))
	if vulnerabilities != "" {
		query := url.Values{
			"vulnerabilities": {vulnerabilities},
		}
		u.RawQuery = query.Encode()

	}
	req, err := c.client.NewRequest(ctx, http.MethodGet, containerRegistryName, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
