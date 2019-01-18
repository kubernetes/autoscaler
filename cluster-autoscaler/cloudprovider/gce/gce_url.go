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

package gce

import (
	"fmt"
	"strings"
)

const (
	gceUrlSchema        = "https"
	gceDomainSuffix     = "googleapis.com/compute/v1/projects/"
	gcePrefix           = gceUrlSchema + "://content." + gceDomainSuffix
	instanceUrlTemplate = gcePrefix + "%s/zones/%s/instances/%s"
	migUrlTemplate      = gcePrefix + "%s/zones/%s/instanceGroups/%s"
)

// ParseMigUrl expects url in format:
// https://content.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instanceGroups/<name>
func ParseMigUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instanceGroups")
}

// ParseIgmUrl expects url in format:
// https://content.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instanceGroupManagers/<name>
func ParseIgmUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instanceGroupManagers")
}

// ParseInstanceUrl expects url in format:
// https://content.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instances/<name>
func ParseInstanceUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instances")
}

// ParseInstanceUrlRef expects url in format:
// https://content.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instances/<name>
// and returns a GceRef struct for it.
func ParseInstanceUrlRef(url string) (GceRef, error) {
	project, zone, name, err := parseGceUrl(url, "instances")
	if err != nil {
		return GceRef{}, err
	}
	return GceRef{
		Project: project,
		Zone:    zone,
		Name:    name,
	}, nil
}

// GenerateInstanceUrl generates url for instance.
func GenerateInstanceUrl(ref GceRef) string {
	return fmt.Sprintf(instanceUrlTemplate, ref.Project, ref.Zone, ref.Name)
}

// GenerateMigUrl generates url for instance.
func GenerateMigUrl(ref GceRef) string {
	return fmt.Sprintf(migUrlTemplate, ref.Project, ref.Zone, ref.Name)
}

func parseGceUrl(url, expectedResource string) (project string, zone string, name string, err error) {
	errMsg := fmt.Errorf("Wrong url: expected format https://content.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/%s/<name>, got %s", expectedResource, url)
	if !strings.Contains(url, gceDomainSuffix) {
		return "", "", "", errMsg
	}
	if !strings.HasPrefix(url, gceUrlSchema) {
		return "", "", "", errMsg
	}
	splitted := strings.Split(strings.Split(url, gceDomainSuffix)[1], "/")
	if len(splitted) != 5 || splitted[1] != "zones" {
		return "", "", "", errMsg
	}
	if splitted[3] != expectedResource {
		return "", "", "", fmt.Errorf("Wrong resource in url: expected %s, got %s", expectedResource, splitted[3])
	}
	project = splitted[0]
	zone = splitted[2]
	name = splitted[4]
	return project, zone, name, nil
}
