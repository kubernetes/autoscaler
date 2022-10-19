/*
Copyright 2016 The Kubernetes Authors.

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

package kamatera

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/klog/v2"
	"strings"
)

const (
	userAgent = "kubernetes/cluster-autoscaler/" + version.ClusterAutoscalerVersion
)

// NewKamateraApiClientRest factory to create new Rest API Client struct
func NewKamateraApiClientRest(clientId string, secret string, url string) (client KamateraApiClientRest) {
	return KamateraApiClientRest{
		userAgent: userAgent,
		clientId:  clientId,
		secret:    secret,
		url:       url,
	}
}

// KamateraServerPostRequest struct for Kamatera server post request
type KamateraServerPostRequest struct {
	ServerName string `json:"name"`
}

// KamateraServerTerminatePostRequest struct for Kamatera server terminate post request
type KamateraServerTerminatePostRequest struct {
	ServerName string `json:"name"`
	Force      bool   `json:"force"`
}

// KamateraServerCreatePostRequest struct for Kamatera server create post request
type KamateraServerCreatePostRequest struct {
	Name               string `json:"name"`
	Password           string `json:"password"`
	PasswordValidate   string `json:"passwordValidate"`
	SshKey             string `json:"ssh-key"`
	Datacenter         string `json:"datacenter"`
	Image              string `json:"image"`
	Cpu                string `json:"cpu"`
	Ram                string `json:"ram"`
	Disk               string `json:"disk"`
	Dailybackup        string `json:"dailybackup"`
	Managed            string `json:"managed"`
	Network            string `json:"network"`
	Quantity           int    `json:"quantity"`
	BillingCycle       string `json:"billingCycle"`
	MonthlyPackage     string `json:"monthlypackage"`
	Poweronaftercreate string `json:"poweronaftercreate"`
	ScriptFile         string `json:"script-file"`
	UserdataFile       string `json:"userdata-file"`
	Tag                string `json:"tag"`
}

// KamateraApiClientRest is the struct to perform API calls
type KamateraApiClientRest struct {
	userAgent string
	clientId  string
	secret    string
	url       string
}

// ListServers returns a list of all servers in the relevant account and fetches their tags
func (c *KamateraApiClientRest) ListServers(ctx context.Context, instances map[string]*Instance) ([]Server, error) {
	res, err := request(
		ctx,
		ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
		"GET",
		"/service/servers",
		nil,
	)
	if err != nil {
		return nil, err
	}
	var servers []Server
	for _, server := range res.([]interface{}) {
		server := server.(map[string]interface{})
		serverName := server["name"].(string)
		serverPowerOn := server["power"].(string) == "on"
		serverTags, err := c.getServerTags(ctx, serverName, instances)
		if err != nil {
			return nil, err
		}
		servers = append(servers, Server{
			Name:    serverName,
			Tags:    serverTags,
			PowerOn: serverPowerOn,
		})
	}
	return servers, nil
}

// DeleteServer deletes a server according to the given name
func (c *KamateraApiClientRest) DeleteServer(ctx context.Context, name string) error {
	res, err := request(
		ctx,
		ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
		"POST",
		"/service/server/poweroff",
		KamateraServerPostRequest{ServerName: name},
	)
	if err == nil {
		commandId := res.([]interface{})[0].(string)
		_, err = waitCommand(
			ctx,
			ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
			commandId,
		)
		if err != nil {
			return err
		}
	} else {
		klog.V(1).Infof("Failed to power off server but will attempt to terminate anyway %s: %v", name, err)
	}
	_, err = request(
		ctx,
		ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
		"POST",
		"/service/server/terminate",
		KamateraServerTerminatePostRequest{ServerName: name, Force: true},
	)
	if err != nil {
		return err
	}
	return nil
}

// CreateServers creates new servers according to the given configuration
func (c *KamateraApiClientRest) CreateServers(ctx context.Context, count int, config ServerConfig) ([]Server, error) {
	var tags []string
	for _, tag := range config.Tags {
		tags = append(tags, tag)
	}
	Tag, err := kamateraRequestArray(tags)
	if err != nil {
		return nil, err
	}
	Network, err := kamateraRequestArray(config.Networks)
	if err != nil {
		return nil, err
	}
	Disk, err := kamateraRequestArray(config.Disks)
	if err != nil {
		return nil, err
	}
	serverNameCommandIds := make(map[string]string)
	for i := 0; i < count; i++ {
		serverName := kamateraServerName(config.NamePrefix)
		res, err := request(
			ctx,
			ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
			"POST",
			"/service/server",
			KamateraServerCreatePostRequest{
				Name:               serverName,
				Password:           config.Password,
				PasswordValidate:   config.Password,
				SshKey:             config.SshKey,
				Datacenter:         config.Datacenter,
				Image:              config.Image,
				Cpu:                config.Cpu,
				Ram:                config.Ram,
				Disk:               Disk,
				Dailybackup:        kamateraRequestBool(config.Dailybackup),
				Managed:            kamateraRequestBool(config.Managed),
				Network:            Network,
				Quantity:           1,
				BillingCycle:       config.BillingCycle,
				MonthlyPackage:     config.MonthlyPackage,
				Poweronaftercreate: "yes",
				ScriptFile:         config.ScriptFile,
				UserdataFile:       config.UserdataFile,
				Tag:                Tag,
			},
		)
		if err != nil {
			return nil, err
		}
		if config.Password == "__generate__" {
			resData := res.(map[string]interface{})
			klog.V(2).Infof("Generated password for server %s: %s", serverName, resData["password"].(string))
			serverNameCommandIds[serverName] = resData["commandIds"].([]interface{})[0].(string)
		} else {
			serverNameCommandIds[serverName] = res.([]interface{})[0].(string)
		}
	}
	results, err := waitCommands(
		ctx,
		ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
		serverNameCommandIds,
	)
	if err != nil {
		return nil, err
	}
	var servers []Server
	for serverName := range results {
		servers = append(servers, Server{
			Name: serverName,
			Tags: tags,
		})
	}
	return servers, nil
}

func (c *KamateraApiClientRest) getServerTags(ctx context.Context, serverName string, instances map[string]*Instance) ([]string, error) {
	if instances[serverName] == nil {
		res, err := request(
			ctx,
			ProviderConfig{ApiUrl: c.url, ApiClientID: c.clientId, ApiSecret: c.secret},
			"POST",
			"/server/tags",
			KamateraServerPostRequest{ServerName: serverName},
		)
		if err != nil {
			return nil, err
		}
		var tags []string
		for _, row := range res.([]interface{}) {
			row := row.(map[string]interface{})
			tags = append(tags, row["tag name"].(string))
		}
		return tags, nil
	}
	return instances[serverName].Tags, nil
}

func kamateraRequestBool(val bool) string {
	if val {
		return "yes"
	}
	return "no"
}

func kamateraRequestArray(val []string) (string, error) {
	for _, v := range val {
		if strings.Contains(v, " ") {
			return "", fmt.Errorf("invalid Kamatera server configuration array value (\"%s\"): ,must not contain spaces", v)
		}
	}
	return strings.Join(val, " "), nil
}

func kamateraServerName(namePrefix string) string {
	if len(namePrefix) > 0 {
		namePrefix = fmt.Sprintf("%s-", namePrefix)
	}
	return fmt.Sprintf("%s%s", namePrefix, strings.ReplaceAll(uuid.New().String(), "-", ""))
}
