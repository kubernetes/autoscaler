# flock
[![TravisCI Build Status](https://img.shields.io/travis/gofrs/flock/master.svg?style=flat)](https://travis-ci.org/gofrs/flock)
[![GoDoc](https://img.shields.io/badge/godoc-flock-blue.svg?style=flat)](https://godoc.org/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock)
[![License](https://img.shields.io/badge/license-BSD_3--Clause-brightgreen.svg?style=flat)](https://k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock)](https://goreportcard.com/report/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock)

`flock` implements a thread-safe sync.Locker interface for file locking. It also
includes a non-blocking TryLock() function to allow locking without blocking execution.

## License
`flock` is released under the BSD 3-Clause License. See the `LICENSE` file for more details.

## Go Compatibility
This package makes use of the `context` package that was introduced in Go 1.7. As such, this
package has an implicit dependency on Go 1.7+.

## Installation
```
go get -u k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock
```

## Usage
```Go
import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock"

fileLock := flock.New("/var/lock/go-lock.lock")

locked, err := fileLock.TryLock()

if err != nil {
	// handle locking error
}

if locked {
	// do work
	fileLock.Unlock()
}
```

For more detailed usage information take a look at the package API docs on
[GoDoc](https://godoc.org/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/github.com/gofrs/flock).
