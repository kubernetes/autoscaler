/*
Copyright 2024 The Kubernetes Authors.

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

package log

import (
	"time"

	"github.com/spf13/pflag"
)

type Format string

const (
	JSON Format = "json"
	Text Format = "text"
)

const (
	flagLogFormat         = "log-format"
	flagLogFlushFrequency = "log-flush-frequency"

	DefaultFormat         = Text
	DefaultFlushFrequency = 5 * time.Second
)

func BindCLIFlags(fs *pflag.FlagSet) {
	fs.String(flagLogFormat, "text", "The log format to use. One of: text, json.")
	// log-flush-frequency is registered in kubernetes component-base.
}

type Options struct {
	Format         Format
	FlushFrequency time.Duration
}

func OptionsFromCLIFlags(fs *pflag.FlagSet) *Options {
	var o Options

	o.Format = Format(fs.Lookup(flagLogFormat).Value.String())
	if o.Format != JSON && o.Format != Text {
		SetupError("Invalid log format", flagLogFormat, o.Format)
		o.Format = DefaultFormat
	}

	freq, err := fs.GetDuration(flagLogFlushFrequency)
	if err != nil {
		SetupError("Invalid log flush frequency", flagLogFlushFrequency, err)
		o.FlushFrequency = DefaultFlushFrequency
	}
	o.FlushFrequency = freq

	return &o
}
