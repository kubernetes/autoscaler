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
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// EC2Metadata is an abstraction over the AWS metadata service.
type EC2Metadata interface {
	// Query the EC2 metadata service (used to discover instance-id etc)
	GetMetadata(path string) (string, error)
}

// ImdsEC2MetadataAdapter wraps the AWS SDK V2's IMDS client to implement EC2Metadata.
// The API for GetMetadata changed between V1 and V2, so this adapter allows the use of the same interface.
type ImdsEC2MetadataAdapter struct {
	imdsClient *imds.Client
}

var _ EC2Metadata = &ImdsEC2MetadataAdapter{}

func (i *ImdsEC2MetadataAdapter) GetMetadata(path string) (string, error) {
	input := imds.GetMetadataInput{Path: path}
	output, err := i.imdsClient.GetMetadata(context.Background(), &input)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(output.Content); err != nil {
		return "", err
	}
	return buffer.String(), nil
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
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}
		c.metadataClientCache = &ImdsEC2MetadataAdapter{imdsClient: imds.NewFromConfig(cfg)}
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
