/*
Copyright 2021 The Kubernetes Authors.

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
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	gce "google.golang.org/api/compute/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

var (
	errFetchMig             = errors.New("fetch migs error")
	errFetchMigTargetSize   = errors.New("fetch mig target size error")
	errFetchMigBaseName     = errors.New("fetch mig basename error")
	errFetchMigTemplateName = errors.New("fetch mig template name error")
	errFetchMigTemplate     = errors.New("fetch mig template error")

	gceRef = GceRef{
		Project: "project",
		Zone:    "us-test1",
		Name:    "mig",
	}
	mig = &gceMig{
		gceRef: gceRef,
	}
)

type mockAutoscalingGceClient struct {
	fetchMigs            func(string) ([]*gce.InstanceGroupManager, error)
	fetchMigTargetSize   func(GceRef) (int64, error)
	fetchMigBasename     func(GceRef) (string, error)
	fetchMigTemplateName func(GceRef) (string, error)
	fetchMigTemplate     func(GceRef, string) (*gce.InstanceTemplate, error)
}

func (client *mockAutoscalingGceClient) FetchMachineType(_, _ string) (*gce.MachineType, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchMachineTypes(_ string) ([]*gce.MachineType, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchAllMigs(zone string) ([]*gce.InstanceGroupManager, error) {
	return client.fetchMigs(zone)
}

func (client *mockAutoscalingGceClient) FetchMigTargetSize(migRef GceRef) (int64, error) {
	return client.fetchMigTargetSize(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigBasename(migRef GceRef) (string, error) {
	return client.fetchMigBasename(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigInstances(_ GceRef) ([]cloudprovider.Instance, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchMigTemplateName(migRef GceRef) (string, error) {
	return client.fetchMigTemplateName(migRef)
}

func (client *mockAutoscalingGceClient) FetchMigTemplate(migRef GceRef, templateName string) (*gce.InstanceTemplate, error) {
	return client.fetchMigTemplate(migRef, templateName)
}

func (client *mockAutoscalingGceClient) FetchMigsWithName(_ string, _ *regexp.Regexp) ([]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchZones(_ string) ([]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) FetchAvailableCpuPlatforms() (map[string][]string, error) {
	return nil, nil
}

func (client *mockAutoscalingGceClient) ResizeMig(_ GceRef, _ int64) error {
	return nil
}

func (client *mockAutoscalingGceClient) DeleteInstances(_ GceRef, _ []GceRef) error {
	return nil
}

func (client *mockAutoscalingGceClient) CreateInstances(_ GceRef, _ string, _ int64, _ []string) error {
	return nil
}

func TestGetMigTargetSize(t *testing.T) {
	targetSize := int64(42)
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:       gceRef.Zone,
		Name:       gceRef.Name,
		TargetSize: targetSize,
	}

	testCases := []struct {
		name               string
		cache              *GceCache
		fetchMigs          func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTargetSize func(GceRef) (int64, error)
		expectedTargetSize int64
		expectedErr        error
	}{
		{
			name: "target size in cache",
			cache: &GceCache{
				migs:               map[GceRef]Mig{gceRef: mig},
				migTargetSizeCache: map[GceRef]int64{gceRef: targetSize},
			},
			expectedTargetSize: targetSize,
		},
		{
			name:               "target size from cache fill",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill without mig, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTargetSize: fetchMigTargetSizeConst(targetSize),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill failure, fallback success",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsFail,
			fetchMigTargetSize: fetchMigTargetSizeConst(targetSize),
			expectedTargetSize: targetSize,
		},
		{
			name:               "cache fill without mig, fallback failure",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTargetSize: fetchMigTargetSizeFail,
			expectedErr:        errFetchMigTargetSize,
		},
		{
			name:               "cache fill failure, fallback failure",
			cache:              emptyCache(),
			fetchMigs:          fetchMigsFail,
			fetchMigTargetSize: fetchMigTargetSizeFail,
			expectedErr:        errFetchMigTargetSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:          tc.fetchMigs,
				fetchMigTargetSize: tc.fetchMigTargetSize,
			}
			provider := NewCachingMigInfoProvider(tc.cache, client, gceRef.Project)

			targetSize, err := provider.GetMigTargetSize(gceRef)
			cachedTargetSize, found := tc.cache.GetMigTargetSize(gceRef)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTargetSize, targetSize)
				assert.Equal(t, tc.expectedTargetSize, cachedTargetSize)
			}
		})
	}
}

func TestGetMigBasename(t *testing.T) {
	basename := "base-instance-name"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:             gceRef.Zone,
		Name:             gceRef.Name,
		BaseInstanceName: basename,
	}

	testCases := []struct {
		name             string
		cache            *GceCache
		fetchMigs        func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigBasename func(GceRef) (string, error)
		expectedBasename string
		expectedErr      error
	}{
		{
			name: "basename in cache",
			cache: &GceCache{
				migs:             map[GceRef]Mig{gceRef: mig},
				migBaseNameCache: map[GceRef]string{gceRef: basename},
			},
			expectedBasename: basename,
		},
		{
			name:             "target size from cache fill",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedBasename: basename,
		},
		{
			name:             "cache fill without mig, fallback success",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigBasename: fetchMigBasenameConst(basename),
			expectedBasename: basename,
		},
		{
			name:             "cache fill failure, fallback success",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsFail,
			fetchMigBasename: fetchMigBasenameConst(basename),
			expectedBasename: basename,
		},
		{
			name:             "cache fill without mig, fallback failure",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigBasename: fetchMigBasenameFail,
			expectedErr:      errFetchMigBaseName,
		},
		{
			name:             "cache fill failure, fallback failure",
			cache:            emptyCache(),
			fetchMigs:        fetchMigsFail,
			fetchMigBasename: fetchMigBasenameFail,
			expectedErr:      errFetchMigBaseName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:        tc.fetchMigs,
				fetchMigBasename: tc.fetchMigBasename,
			}
			provider := NewCachingMigInfoProvider(tc.cache, client, gceRef.Project)

			basename, err := provider.GetMigBasename(gceRef)
			cachedBasename, found := tc.cache.GetMigBasename(gceRef)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedBasename, basename)
				assert.Equal(t, tc.expectedBasename, cachedBasename)
			}
		})
	}
}

func TestGetMigInstanceTemplateName(t *testing.T) {
	templateName := "template-name"
	instanceGroupManager := &gce.InstanceGroupManager{
		Zone:             gceRef.Zone,
		Name:             gceRef.Name,
		InstanceTemplate: templateName,
	}

	testCases := []struct {
		name                 string
		cache                *GceCache
		fetchMigs            func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName func(GceRef) (string, error)
		expectedTemplateName string
		expectedErr          error
	}{
		{
			name: "template name in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
			},
			expectedTemplateName: templateName,
		},
		{
			name:                 "target size from cache fill",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{instanceGroupManager}),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill without mig, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTemplateName: fetchMigTemplateNameConst(templateName),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill failure, fallback success",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameConst(templateName),
			expectedTemplateName: templateName,
		},
		{
			name:                 "cache fill without mig, fallback failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsConst([]*gce.InstanceGroupManager{}),
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
		{
			name:                 "cache fill failure, fallback failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:            tc.fetchMigs,
				fetchMigTemplateName: tc.fetchMigTemplateName,
			}
			provider := NewCachingMigInfoProvider(tc.cache, client, gceRef.Project)

			templateName, err := provider.GetMigInstanceTemplateName(gceRef)
			cachedTemplateName, found := tc.cache.GetMigInstanceTemplateName(gceRef)

			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedErr == nil, found)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTemplateName, templateName)
				assert.Equal(t, tc.expectedTemplateName, cachedTemplateName)
			}
		})
	}
}

func TestGetMigInstanceTemplate(t *testing.T) {
	templateName := "template-name"
	template := &gce.InstanceTemplate{
		Name:        templateName,
		Description: "instance template",
	}
	oldTemplate := &gce.InstanceTemplate{
		Name:        "old-template-name",
		Description: "old instance template",
	}

	testCases := []struct {
		name                   string
		cache                  *GceCache
		fetchMigs              func(string) ([]*gce.InstanceGroupManager, error)
		fetchMigTemplateName   func(GceRef) (string, error)
		fetchMigTemplate       func(GceRef, string) (*gce.InstanceTemplate, error)
		expectedTemplate       *gce.InstanceTemplate
		expectedCachedTemplate *gce.InstanceTemplate
		expectedErr            error
	}{
		{
			name: "template in cache",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{gceRef: template},
			},
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
			},
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache with old template, fetch success",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{gceRef: oldTemplate},
			},
			fetchMigTemplate:       fetchMigTemplateConst(template),
			expectedTemplate:       template,
			expectedCachedTemplate: template,
		},
		{
			name: "cache without template, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
				instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
			},
			fetchMigTemplate: fetchMigTemplateFail,
			expectedErr:      errFetchMigTemplate,
		},
		{
			name: "cache with old template, fetch failure",
			cache: &GceCache{
				migs:                      map[GceRef]Mig{gceRef: mig},
				instanceTemplateNameCache: map[GceRef]string{gceRef: templateName},
				instanceTemplatesCache:    map[GceRef]*gce.InstanceTemplate{gceRef: oldTemplate},
			},
			fetchMigTemplate:       fetchMigTemplateFail,
			expectedCachedTemplate: oldTemplate,
			expectedErr:            errFetchMigTemplate,
		},
		{
			name:                 "template name fetch failure",
			cache:                emptyCache(),
			fetchMigs:            fetchMigsFail,
			fetchMigTemplateName: fetchMigTemplateNameFail,
			expectedErr:          errFetchMigTemplateName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockAutoscalingGceClient{
				fetchMigs:            tc.fetchMigs,
				fetchMigTemplateName: tc.fetchMigTemplateName,
				fetchMigTemplate:     tc.fetchMigTemplate,
			}
			provider := NewCachingMigInfoProvider(tc.cache, client, gceRef.Project)

			template, err := provider.GetMigInstanceTemplate(gceRef)
			cachedTemplate, found := tc.cache.GetMigInstanceTemplate(gceRef)

			assert.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedTemplate, template)
			}

			assert.Equal(t, tc.expectedCachedTemplate != nil, found)
			if tc.expectedCachedTemplate != nil {
				assert.Equal(t, tc.expectedCachedTemplate, cachedTemplate)
			}
		})
	}
}

func emptyCache() *GceCache {
	return &GceCache{
		migs:                      map[GceRef]Mig{gceRef: mig},
		migTargetSizeCache:        make(map[GceRef]int64),
		migBaseNameCache:          make(map[GceRef]string),
		instanceTemplateNameCache: make(map[GceRef]string),
		instanceTemplatesCache:    make(map[GceRef]*gce.InstanceTemplate),
	}
}

func fetchMigsFail(_ string) ([]*gce.InstanceGroupManager, error) {
	return nil, errFetchMig
}

func fetchMigsConst(migs []*gce.InstanceGroupManager) func(string) ([]*gce.InstanceGroupManager, error) {
	return func(string) ([]*gce.InstanceGroupManager, error) {
		return migs, nil
	}
}

func fetchMigTargetSizeFail(_ GceRef) (int64, error) {
	return 0, errFetchMigTargetSize
}

func fetchMigTargetSizeConst(targetSize int64) func(GceRef) (int64, error) {
	return func(GceRef) (int64, error) {
		return targetSize, nil
	}
}

func fetchMigBasenameFail(_ GceRef) (string, error) {
	return "", errFetchMigBaseName
}

func fetchMigBasenameConst(basename string) func(GceRef) (string, error) {
	return func(GceRef) (string, error) {
		return basename, nil
	}
}

func fetchMigTemplateNameFail(_ GceRef) (string, error) {
	return "", errFetchMigTemplateName
}

func fetchMigTemplateNameConst(templateName string) func(GceRef) (string, error) {
	return func(GceRef) (string, error) {
		return templateName, nil
	}
}

func fetchMigTemplateFail(_ GceRef, _ string) (*gce.InstanceTemplate, error) {
	return nil, errFetchMigTemplate
}

func fetchMigTemplateConst(template *gce.InstanceTemplate) func(GceRef, string) (*gce.InstanceTemplate, error) {
	return func(GceRef, string) (*gce.InstanceTemplate, error) {
		return template, nil
	}
}
