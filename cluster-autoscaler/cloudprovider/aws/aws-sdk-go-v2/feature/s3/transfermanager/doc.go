// Package transfermanager implements the Amazon S3 Transfer Manager, a
// high-level S3 client library.
//
// Package transfermanager is the new iteration of the original
// [feature/s3/manager] module implemented for the AWS SDK Go v2.
//
// # Beta
//
// This module is currently in a BETA release state. It is not subject to the
// same backwards-compatibility guarantees provided by the generally-available
// (GA) AWS SDK for Go v2. Features may be added or removed without warning,
// and APIs may break.
//
// For the current GA transfer manager for AWS SDK Go v2, see
// [feature/s3/manager].
//
// # Features
//
// Package transfermanager implements a high-level S3 client with support for the
// following:
//   - [Client.PutObject] - enhanced object write support w/ automatic
//     multipart upload for large objects
//
// The package also exposes several opt-in hooks that configure an
// http.Transport that may convey performance/reliability enhancements in
// certain user environments:
//   - round-robin DNS ([WithRoundRobinDNS])
//   - multi-NIC dialer ([WithRotoDialer])
//
// [feature/s3/manager]: https://pkg.go.dev/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/feature/s3/manager
package transfermanager
