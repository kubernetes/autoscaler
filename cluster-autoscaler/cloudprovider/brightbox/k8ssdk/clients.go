// Copyright 2018 Brightbox Systems Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8ssdk

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// EC2Metadata is an abstraction over the AWS metadata service.
type EC2Metadata interface {
	// Query the EC2 metadata service (used to discover instance-id etc)
	GetMetadata(path string) (string, error)
}

// Cloud allows access to the Brightbox Cloud and/or any EC2 compatible metadata.
type Cloud struct {
	client              CloudAccess
	metadataClientCache EC2Metadata
}

// MetadataClient returns the EC2 Metadata client, or creates a new client
// from the default AWS config if one doesn't exist.
func (c *Cloud) MetadataClient() (EC2Metadata, error) {
	if c.metadataClientCache == nil {
		cfg, err := session.NewSession()
		if err != nil {
			return nil, err
		}
		c.metadataClientCache = ec2metadata.New(cfg)
	}

	return c.metadataClientCache, nil
}

// CloudClient returns the Brightbox Cloud client, or creates a new client from the current environment if one doesn't exist.
func (c *Cloud) CloudClient() (CloudAccess, error) {
	if c.client == nil {
		client, err := obtainCloudClient()
		if err != nil {
			return nil, err
		}
		c.client = client
	}

	return c.client, nil
}
