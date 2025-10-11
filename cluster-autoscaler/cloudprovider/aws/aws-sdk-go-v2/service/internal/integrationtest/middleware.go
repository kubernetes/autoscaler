package integrationtest

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"

// RemoveOperationInputValidationMiddleware removes the validation middleware
// from the stack.
func RemoveOperationInputValidationMiddleware(stack *middleware.Stack) error {
	stack.Initialize.Remove("OperationInputValidation")
	return nil
}
