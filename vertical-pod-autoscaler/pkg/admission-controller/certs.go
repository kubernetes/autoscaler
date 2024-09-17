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
	"crypto/tls"
	"os"
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

type certsConfig struct {
	clientCaFile, tlsCertFile, tlsPrivateKey *string
	reload                                   *bool
}

func readFile(filePath string) []byte {
	res, err := os.ReadFile(filePath)
	if err != nil {
		klog.ErrorS(err, "Error reading certificate file", "file", filePath)
		return nil
	}
	klog.V(3).InfoS("Successfully read bytes from file", "bytes", len(res), "file", filePath)
	return res
}

type certReloader struct {
	tlsCertPath string
	tlsKeyPath  string
	cert        *tls.Certificate
	mu          sync.RWMutex
}

func (cr *certReloader) start(stop <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err = watcher.Add(path.Dir(cr.tlsCertPath)); err != nil {
		return err
	}
	if err = watcher.Add(path.Dir(cr.tlsKeyPath)); err != nil {
		return err
	}
	go func() {
		defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					klog.V(2).InfoS("New certificate found, reloading")
					if err := cr.load(); err != nil {
						klog.ErrorS(err, "Failed to reload certificate")
					}
				}
			case err := <-watcher.Errors:
				klog.Warningf("Error watching certificate files: %s", err)
			case <-stop:
				return
			}
		}
	}()
	return nil
}

func (cr *certReloader) load() error {
	cert, err := tls.LoadX509KeyPair(cr.tlsCertPath, cr.tlsKeyPath)
	if err != nil {
		return err
	}
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.cert = &cert
	return nil
}

func (cr *certReloader) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.cert, nil
}
