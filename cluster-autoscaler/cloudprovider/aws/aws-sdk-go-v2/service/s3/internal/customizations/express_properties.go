package customizations

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go"

// GetPropertiesBackend returns a resolved endpoint backend from the property
// set.
func GetPropertiesBackend(p *smithy.Properties) string {
	v, _ := p.Get("backend").(string)
	return v
}

// GetIdentityPropertiesBucket returns the S3 bucket from identity properties.
func GetIdentityPropertiesBucket(ip *smithy.Properties) (string, bool) {
	v, ok := ip.Get(bucketKey{}).(string)
	return v, ok
}

// SetIdentityPropertiesBucket sets the S3 bucket to identity properties.
func SetIdentityPropertiesBucket(ip *smithy.Properties, bucket string) {
	ip.Set(bucketKey{}, bucket)
}
