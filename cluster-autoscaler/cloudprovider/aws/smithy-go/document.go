package smithy

// Document provides access to loosely structured data in a document-like
// format.
//
// Deprecated: See the k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/document package.
type Document interface {
	UnmarshalDocument(interface{}) error
	GetValue() (interface{}, error)
}
