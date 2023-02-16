/*
Copyright 2023 The Kubernetes Authors.

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

package shareddefaults

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"os"
	"path/filepath"
	"runtime"
)

// SharedCredentialsFilename returns the SDK's default file path
// for the shared credentials file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.volcengine/credentials
//   - Windows: %USERPROFILE%\.volcengine\credentials
func SharedCredentialsFilename() string {
	return filepath.Join(UserHomeDir(), ".volcengine", "credentials")
}

// SharedConfigFilename returns the SDK's default file path for
// the shared config file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.volcengine/config
//   - Windows: %USERPROFILE%\.volcengine\config
func SharedConfigFilename() string {
	return filepath.Join(UserHomeDir(), ".volcengine", "config")
}

// UserHomeDir returns the home directory for the user the process is
// running under.
func UserHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
