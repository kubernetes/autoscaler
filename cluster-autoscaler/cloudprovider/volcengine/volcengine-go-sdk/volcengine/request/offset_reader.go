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

package request

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"io"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/internal/sdkio"
)

// offsetReader is a thread-safe io.ReadCloser to prevent racing
// with retrying requests
type offsetReader struct {
	buf    io.ReadSeeker
	lock   sync.Mutex
	closed bool
}

func newOffsetReader(buf io.ReadSeeker, offset int64) (*offsetReader, error) {
	reader := &offsetReader{}
	_, err := buf.Seek(offset, sdkio.SeekStart)
	if err != nil {
		return nil, err
	}

	reader.buf = buf
	return reader, nil
}

// Close will close the instance of the offset reader's access to
// the underlying io.ReadSeeker.
func (o *offsetReader) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.closed = true
	return nil
}

// Read is a thread-safe read of the underlying io.ReadSeeker
func (o *offsetReader) Read(p []byte) (int, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.closed {
		return 0, io.EOF
	}

	return o.buf.Read(p)
}

// Seek is a thread-safe seeking operation.
func (o *offsetReader) Seek(offset int64, whence int) (int64, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	return o.buf.Seek(offset, whence)
}

// CloseAndCopy will return a new offsetReader with a copy of the old buffer
// and close the old buffer.
func (o *offsetReader) CloseAndCopy(offset int64) (*offsetReader, error) {
	if err := o.Close(); err != nil {
		return nil, err
	}
	return newOffsetReader(o.buf, offset)
}
