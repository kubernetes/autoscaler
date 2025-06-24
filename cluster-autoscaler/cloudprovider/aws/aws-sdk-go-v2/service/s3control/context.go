package s3control

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

// stashOperationInput clones the operation input at the start of the serialize
// phase to be modified as necessary by s3control customizations, then restores
// it after that phase (at the start of build) such that auth and endpoint
// resolution can flow as normal
type stashOperationInput struct {
	origInput interface{}
}

func (*stashOperationInput) ID() string {
	return "stashOperationInput"
}

type copyable interface {
	copy() interface{}
}

func (m *stashOperationInput) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if v, ok := in.Parameters.(copyable); ok {
		m.origInput = v
		in.Parameters = v.copy()
		return next.HandleSerialize(ctx, in)
	}
	return next.HandleSerialize(ctx, in)
}

func (m *stashOperationInput) HandleBuild(
	ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler,
) (
	out middleware.BuildOutput, metadata middleware.Metadata, err error,
) {
	if m.origInput != nil {
		ctx = setOperationInput(ctx, m.origInput)
	}
	return next.HandleBuild(ctx, in)
}

func addStashOperationInput(stack *middleware.Stack) error {
	m := &stashOperationInput{}
	if _, err := stack.Serialize.Swap("setOperationInput", m); err != nil {
		return err
	}
	if err := stack.Build.Add(m, middleware.Before); err != nil {
		return err
	}
	return nil
}
