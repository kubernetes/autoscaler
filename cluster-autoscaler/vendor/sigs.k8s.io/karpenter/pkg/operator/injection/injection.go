/*
Copyright The Kubernetes Authors.

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

package injection

import (
	"context"
	"flag"
	"os"

	"github.com/samber/lo"

	"sigs.k8s.io/karpenter/pkg/operator/options"
)

type controllerNameKeyType struct{}

var controllerNameKey = controllerNameKeyType{}

func WithControllerName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, controllerNameKey, name)
}

func GetControllerName(ctx context.Context) string {
	name := ctx.Value(controllerNameKey)
	if name == nil {
		return ""
	}
	return name.(string)
}

func WithOptionsOrDie(ctx context.Context, opts ...options.Injectable) context.Context {
	fs := &options.FlagSet{
		FlagSet: flag.NewFlagSet("karpenter", flag.ContinueOnError),
	}
	for _, opt := range opts {
		opt.AddFlags(fs)
	}
	for _, opt := range opts {
		lo.Must0(opt.Parse(fs, os.Args[1:]...))
	}
	for _, opt := range opts {
		ctx = opt.ToContext(ctx)
	}
	return ctx
}
