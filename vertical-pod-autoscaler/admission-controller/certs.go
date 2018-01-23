/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"os"
	"path"

	"github.com/golang/glog"
)

type certsContainer struct {
	caKey, caCert, serverKey, serverCert []byte
}

func readFile(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		glog.Error(err)
		return nil
	}
	res := make([]byte, 5000)
	count, err := file.Read(res)
	if err != nil {
		glog.Error(err)
		return nil
	}
	glog.Infof("Successfuly read %d bytes from %v", count, filePath)
	return res
}

func initCerts(certsDir string) certsContainer {
	res := certsContainer{}
	res.caKey = readFile(path.Join(certsDir, "caKey.pem"))
	res.caCert = readFile(path.Join(certsDir, "caCert.pem"))
	res.serverKey = readFile(path.Join(certsDir, "serverKey.pem"))
	res.serverCert = readFile(path.Join(certsDir, "serverCert.pem"))
	return res
}
