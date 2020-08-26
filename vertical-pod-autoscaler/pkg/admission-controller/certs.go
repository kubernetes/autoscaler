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
	"io/ioutil"
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog"
)

// KeypairReloader structs holds cert path and certs
type KeypairReloader struct {
	certMu   sync.RWMutex
	cert     *tls.Certificate
	caCert   []byte
	certPath string
	keyPath  string
	caPath   string
}

type certsContainer struct {
	caCert, serverKey, serverCert []byte
}

type certsConfig struct {
	clientCaFile, tlsCertFile, tlsPrivateKey *string
}

func readFile(filePath string) []byte {
	res, err := ioutil.ReadFile(filePath)
	if err != nil {
		klog.Errorf("Error reading certificate file at %s: %v", filePath, err)
		return nil
	}

	klog.V(3).Infof("Successfully read %d bytes from %v", len(res), filePath)
	return res
}

// NewKeypairReloader will load certs on first run and trigger a goroutine for fsnotify watcher
func NewKeypairReloader(config certsConfig) (*KeypairReloader, error) {
	result := &KeypairReloader{
		certPath: *config.tlsCertFile,
		keyPath:  *config.tlsPrivateKey,
		caCert:   readFile(*config.clientCaFile),
	}
	cert, err := tls.LoadX509KeyPair(*config.tlsCertFile, *config.tlsPrivateKey)
	if err != nil {
		return nil, err
	}
	result.cert = &cert

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			watcher.Close()
		}
	}()

	if err := watcher.Add("/etc/tls-certs"); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				// fsnotify.create events will tell us if there are new certs
				if event.Op&fsnotify.Create == fsnotify.Create {
					klog.Info("Reloading certs")
					if err := result.reload(); err != nil {
						klog.Infof("Could not load new certs: %v", err)
					}
				}

				// watch for errors
			case err := <-watcher.Errors:
				klog.Infof("error: %v", err)
			}
		}
	}()

	return result, nil
}

// reload loads updated cert and key whenever they are updated
func (kpr *KeypairReloader) reload() error {
	newCert, err := tls.LoadX509KeyPair(kpr.certPath, kpr.keyPath)
	if err != nil {
		return err
	}
	kpr.certMu.Lock()
	defer kpr.certMu.Unlock()
	kpr.cert = &newCert
	return nil
}

// GetCertificateFunc will return function which will be used as tls.Config.GetCertificate
func (kpr *KeypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		kpr.certMu.RLock()
		defer kpr.certMu.RUnlock()
		return kpr.cert, nil
	}
}
