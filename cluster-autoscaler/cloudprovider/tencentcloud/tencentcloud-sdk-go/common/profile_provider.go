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

package common

import (
	"os"
	"path/filepath"
	"runtime"

	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/errors"
)

const (
	EnvCredentialFile = "TENCENTCLOUD_CREDENTIALS_FILE"
)

type ProfileProvider struct {
}

// DefaultProfileProvider return a default Profile  provider
// profile path :
//  1. The value of the environment variable TENCENTCLOUD_CREDENTIALS_FILE
//  2. linux: ~/.tencentcloud/credentials
//     windows: \c:\Users\NAME\.tencentcloud\credentials
func DefaultProfileProvider() *ProfileProvider {
	return &ProfileProvider{}
}

// getHomePath return home directory according to the system.
// if the environmental variables does not exist, it will return empty string
func getHomePath() string {
	// Windows
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	// *nix
	return os.Getenv("HOME")
}

func getCredentialsFilePath() string {
	homePath := getHomePath()
	if homePath == "" {
		return homePath
	}
	return filepath.Join(homePath, ".tencentcloud", "credentials")
}

func checkDefaultFile() (path string, err error) {
	path = getCredentialsFilePath()
	if path == "" {
		return path, nil
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return path, nil
}

func (p *ProfileProvider) GetCredential() (CredentialIface, error) {
	path, ok := os.LookupEnv(EnvCredentialFile)
	// if not set custom file path, will use the default path
	if !ok {
		var err error
		path, err = checkDefaultFile()
		// only when the file exist but failed read it the err is not nil
		if err != nil {
			return nil, tcerr.NewTencentCloudSDKError(creErr, "Failed to find profile file,"+err.Error(), "")
		}
		// when the path is "" means the file dose not exist
		if path == "" {
			return nil, fileDoseNotExist
		}
		// if the EnvCredentialFile is set to "", will return an error
	} else if path == "" {
		return nil, tcerr.NewTencentCloudSDKError(creErr, "Environment variable '"+EnvCredentialFile+"' cannot be empty", "")
	}

	cfg, err := parse(path)
	if err != nil {
		return nil, err
	}

	sId := cfg.section("default").key("secret_id").string()
	sKey := cfg.section("default").key("secret_key").string()
	// if sId and sKey is "", but the credential file exist, means an error
	if sId == "" || sKey == "" {
		return nil, tcerr.NewTencentCloudSDKError(creErr, "Failed to parse profile file,please confirm whether it contains \"secret_id\" and \"secret_key\" in section: \"default\" ", "")
	}
	return &Credential{
		SecretId:  sId,
		SecretKey: sKey,
		Token:     "",
	}, nil
}
