<<<<<<<< HEAD:cluster-autoscaler/vendor/github.com/prometheus/procfs/cmdline.go
// Copyright 2021 The Prometheus Authors
========
// Copyright 2019 The etcd Authors
//
>>>>>>>> cluster-autoscaler-release-1.22:cluster-autoscaler/vendor/go.etcd.io/etcd/client/pkg/v3/logutil/log_level.go
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

<<<<<<<< HEAD:cluster-autoscaler/vendor/github.com/prometheus/procfs/cmdline.go
package procfs

import (
	"strings"

	"github.com/prometheus/procfs/internal/util"
)

// CmdLine returns the command line of the kernel.
func (fs FS) CmdLine() ([]string, error) {
	data, err := util.ReadFileNoStat(fs.proc.Path("cmdline"))
	if err != nil {
		return nil, err
	}

	return strings.Fields(string(data)), nil
========
package logutil

import (
	"go.uber.org/zap/zapcore"
)

var DefaultLogLevel = "info"

// ConvertToZapLevel converts log level string to zapcore.Level.
func ConvertToZapLevel(lvl string) zapcore.Level {
	var level zapcore.Level
	if err := level.Set(lvl); err != nil {
		panic(err)
	}
	return level
>>>>>>>> cluster-autoscaler-release-1.22:cluster-autoscaler/vendor/go.etcd.io/etcd/client/pkg/v3/logutil/log_level.go
}
