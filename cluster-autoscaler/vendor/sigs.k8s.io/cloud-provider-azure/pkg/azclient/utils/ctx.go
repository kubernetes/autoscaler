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

package utils

import "context"

var (
	ctxKeyClientName        = struct{}{}
	ctxKeyMethodRequest     = struct{}{}
	ctxKeyResourceGroupName = struct{}{}
	ctxKeySubscriptionID    = struct{}{}
)

func ContextWithClientName(ctx context.Context, clientName string) context.Context {
	return context.WithValue(ctx, ctxKeyClientName, clientName)
}

func ClientNameFromContext(ctx context.Context) (string, bool) {
	rv, ok := ctx.Value(ctxKeyClientName).(string)
	return rv, ok
}

func ContextWithRequestMethod(ctx context.Context, methodRequest string) context.Context {
	return context.WithValue(ctx, ctxKeyMethodRequest, methodRequest)
}

func RequestMethodFromContext(ctx context.Context) (string, bool) {
	rv, ok := ctx.Value(ctxKeyMethodRequest).(string)
	return rv, ok
}

func ContextWithResourceGroupName(ctx context.Context, resourceGroupName string) context.Context {
	return context.WithValue(ctx, ctxKeyResourceGroupName, resourceGroupName)
}

func ResourceGroupNameFromContext(ctx context.Context) (string, bool) {
	rv, ok := ctx.Value(ctxKeyResourceGroupName).(string)
	return rv, ok
}

func ContextWithSubscriptionID(ctx context.Context, subscriptionID string) context.Context {
	return context.WithValue(ctx, ctxKeySubscriptionID, subscriptionID)
}

func SubscriptionIDFromContext(ctx context.Context) (string, bool) {
	rv, ok := ctx.Value(ctxKeySubscriptionID).(string)
	return rv, ok
}
