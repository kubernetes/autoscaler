/*
Copyright 2019 The Kubernetes Authors.

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

// ServerConfig struct for Kamatera server config
type ServerConfig struct {
	NamePrefix     string
	Password       string
	SshKey         string
	Datacenter     string
	Image          string
	Cpu            string
	Ram            string
	Disks          []string
	Dailybackup    bool
	Managed        bool
	Networks       []string
	BillingCycle   string
	MonthlyPackage string
	ScriptFile     string
	UserdataFile   string
	Tags           []string
}
