// This file is part of gobizfly
//
// Copyright (C) 2020  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

package gobizfly

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

var _ ServiceInterface = (*service)(nil)

const serviceUrl = "/api/auth/service"

type Service struct {
	Name          string `json:"name"`
	Code          string `json:"code"`
	CanonicalName string `json:"canonical_name"`
	Id            int    `json:"id"`
	Region        string `json:"region"`
	Icon          string `json:"icon"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
	ServiceUrl    string `json:"service_url"`
}

type ServiceList struct {
	Services []*Service `json:"services"`
}

type service struct {
	client *Client
}

type ServiceInterface interface {
	List(ctx context.Context) ([]*Service, error)
	//GetEndpoint(ctx context.Context, name string, region string) (string, error)
}

func (s *service) List(ctx context.Context) ([]*Service, error) {
	u, err := s.client.apiURL.Parse(serviceUrl)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)

	req, err := http.NewRequest("GET", u.String(), buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var services ServiceList

	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, err
	}
	return services.Services, nil
}
