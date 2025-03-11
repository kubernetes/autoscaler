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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

func generateCerts(t *testing.T, org string, caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, []byte) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			Organization: []string{org},
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
	certKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error(err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certKey.PublicKey, caKey)
	if err != nil {
		t.Error(err)
	}

	var certPem bytes.Buffer
	err = pem.Encode(&certPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		t.Error(err)
	}

	var certKeyPem bytes.Buffer
	err = pem.Encode(&certKeyPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certKey),
	})
	if err != nil {
		t.Error(err)
	}
	return certPem.Bytes(), certKeyPem.Bytes()
}

func TestKeypairReloader(t *testing.T) {
	tempDir := t.TempDir()
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			Organization: []string{"ca"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Error(err)
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		t.Error(err)
	}
	caPath := path.Join(tempDir, "ca.crt")
	caFile, err := os.Create(caPath)
	if err != nil {
		t.Error(err)
	}
	err = pem.Encode(caFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		t.Error(err)
	}

	pub, privateKey := generateCerts(t, "first", caCert, caKey)
	certPath := path.Join(tempDir, "cert.crt")
	if err = os.WriteFile(certPath, pub, 0666); err != nil {
		t.Error(err)
	}
	keyPath := path.Join(tempDir, "cert.key")
	if err = os.WriteFile(keyPath, privateKey, 0666); err != nil {
		t.Error(err)
	}

	reloader := certReloader{
		tlsCertPath: certPath,
		tlsKeyPath:  keyPath,
	}
	stop := make(chan struct{})
	defer close(stop)
	if err = reloader.start(stop); err != nil {
		t.Error(err)
	}

	pub, privateKey = generateCerts(t, "second", caCert, caKey)
	if err = os.WriteFile(certPath, pub, 0666); err != nil {
		t.Error(err)
	}
	if err = os.WriteFile(keyPath, privateKey, 0666); err != nil {
		t.Error(err)
	}
	for {
		tlsCert, err := reloader.getCertificate(nil)
		if err != nil {
			t.Error(err)
		}
		if tlsCert == nil {
			continue
		}
		pubDER, _ := pem.Decode(pub)
		if string(tlsCert.Certificate[0]) == string(pubDER.Bytes) {
			return
		}
	}
}
