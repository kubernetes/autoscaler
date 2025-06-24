package customizations

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

type setDefaultAccountID func(input interface{}, accountID string) interface{}

// AddDefaultAccountIDMiddleware adds the DefaultAccountID to the stack using
// the given options.
func AddDefaultAccountIDMiddleware(stack *middleware.Stack, setDefaultAccountID setDefaultAccountID) error {
	return stack.Initialize.Add(&DefaultAccountID{
		setDefaultAccountID: setDefaultAccountID,
	}, middleware.Before)
}

// DefaultAccountID sets the account ID to "-" if it isn't already set
type DefaultAccountID struct {
	setDefaultAccountID setDefaultAccountID
}

// ID returns the id of the middleware
func (*DefaultAccountID) ID() string {
	return "Glacier:DefaultAccountID"
}

// HandleInitialize implements the InitializeMiddleware interface
func (m *DefaultAccountID) HandleInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	in.Parameters = m.setDefaultAccountID(in.Parameters, "-")
	return next.HandleInitialize(ctx, in)
}
