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

/*
Functions to handle Kamatera API calls
Copied from the Kamatera terraform provider:
https://github.com/Kamatera/terraform-provider-kamatera/blob/master/kamatera/request.go
*/

package kamatera

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
	"time"
)

// ProviderConfig is the configuration for the Kamatera cloud provider
type ProviderConfig struct {
	ApiUrl      string
	ApiClientID string
	ApiSecret   string
}

func request(ctx context.Context, provider ProviderConfig, method string, path string, body interface{}, numRetries int, secondsBetweenRetries int) (interface{}, error) {
	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}
	path = strings.TrimPrefix(path, "/")
	url := fmt.Sprintf("%s/%s", provider.ApiUrl, path)
	var result interface{}
	var err error
	for attempt := 0; attempt < numRetries; attempt++ {
		klog.V(2).Infof("kamatera request: %s %s %s", method, url, buf.String())
		if attempt > 0 {
			klog.V(2).Infof("kamatera request retry %d", attempt)
			time.Sleep(time.Duration(secondsBetweenRetries<<attempt) * time.Second)
		}
		req, e := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", provider.ApiUrl, path), buf)
		if e != nil {
			err = e
			continue
		}
		req.Header.Add("AuthClientId", provider.ApiClientID)
		req.Header.Add("AuthSecret", provider.ApiSecret)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		res, e := http.DefaultClient.Do(req)
		if e != nil {
			err = e
			continue
		}
		defer res.Body.Close()
		e = json.NewDecoder(res.Body).Decode(&result)
		if e != nil {
			if res.StatusCode != 200 {
				err = fmt.Errorf("bad status code from Kamatera API: %d", res.StatusCode)
			} else {
				err = fmt.Errorf("invalid response from Kamatera API: %+v", result)
			}
			continue
		}
		if res.StatusCode != 200 {
			err = fmt.Errorf("error response from Kamatera API (%d): %+v", res.StatusCode, result)
			continue
		}
		break
	}
	return result, err
}

func waitCommand(ctx context.Context, provider ProviderConfig, commandID string, numRetries int, secondsBetweenRetries int) (map[string]interface{}, error) {
	startTime := time.Now()
	time.Sleep(2 * time.Second)

	for {
		if startTime.Add(40*time.Minute).Sub(time.Now()) < 0 {
			return nil, errors.New("timeout waiting for Kamatera command to complete")
		}

		time.Sleep(2 * time.Second)

		result, e := request(ctx, provider, "GET", fmt.Sprintf("/service/queue?id=%s", commandID), nil, numRetries, secondsBetweenRetries)
		if e != nil {
			return nil, e
		}

		commands := result.([]interface{})
		if len(commands) != 1 {
			return nil, errors.New("invalid response from Kamatera queue API: invalid number of command responses")
		}

		command := commands[0].(map[string]interface{})
		status, hasStatus := command["status"]
		if hasStatus {
			switch status.(string) {
			case "complete":
				return command, nil
			case "error":
				log, hasLog := command["log"]
				if hasLog {
					return nil, fmt.Errorf("kamatera command failed: %s", log)
				}
				return nil, fmt.Errorf("kamatera command failed: %v", command)
			}
		}
	}
}

func waitCommands(ctx context.Context, provider ProviderConfig, commandIds map[string]string, numRetries int, secondsBetweenRetries int) (map[string]interface{}, error) {
	startTime := time.Now()
	time.Sleep(2 * time.Second)

	commandIdsResults := make(map[string]interface{})
	for id := range commandIds {
		commandIdsResults[id] = nil
	}

	for {
		if startTime.Add((40)*time.Minute).Sub(time.Now()) < 0 {
			return nil, errors.New("timeout waiting for Kamatera commands to complete")
		}

		time.Sleep(2 * time.Second)

		for id, result := range commandIdsResults {
			if result == nil {
				commandId := commandIds[id]
				result, e := request(ctx, provider, "GET", fmt.Sprintf("/service/queue?id=%s", commandId), nil, numRetries, secondsBetweenRetries)
				if e != nil {
					return nil, e
				}
				commands := result.([]interface{})
				if len(commands) != 1 {
					return nil, errors.New("invalid response from Kamatera queue API: invalid number of command responses")
				}
				command := commands[0].(map[string]interface{})
				status, hasStatus := command["status"]
				if hasStatus {
					switch status.(string) {
					case "complete":
						commandIdsResults[id] = command
						break
					case "error":
						log, hasLog := command["log"]
						if hasLog {
							return nil, fmt.Errorf("kamatera command failed: %s", log)
						}
						return nil, fmt.Errorf("kamatera command failed: %v", command)
					}
				}
			}
		}

		numComplete := 0
		for _, result := range commandIdsResults {
			if result != nil {
				numComplete++
			}
		}
		if numComplete == len(commandIds) {
			return commandIdsResults, nil
		}
	}
}
