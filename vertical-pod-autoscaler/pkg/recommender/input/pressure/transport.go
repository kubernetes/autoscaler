/*
Copyright The Kubernetes Authors.

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

package pressure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
)

// UseDirectTransport makes the observer fetch /stats/summary from each node's kubelet endpoint.
// restCfg supplies credentials (token file refresh or client certificates); caFile verifies the kubelet
// serving certificate. There is no insecure mode and redirects are rejected. The URL host is the node
// name, so the certificate is verified against the identity in the Node object, while the dialer
// connects to the node's InternalIP (Hostname as fallback).
func (o *KubeletObserver) UseDirectTransport(restCfg *rest.Config, caFile string, nodes listersv1.NodeLister) error {
	kc := kubeletRESTConfig(restCfg, caFile)
	dialer := &net.Dialer{Timeout: o.cfg.RequestTimeout}
	kc.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		n, err := nodes.Get(host)
		if err != nil {
			return nil, err
		}
		for _, want := range []corev1.NodeAddressType{corev1.NodeInternalIP, corev1.NodeHostName} {
			for _, a := range n.Status.Addresses {
				if a.Type == want {
					return dialer.DialContext(ctx, network, net.JoinHostPort(a.Address, port))
				}
			}
		}
		return nil, fmt.Errorf("node %s has no usable address", host)
	}
	rt, err := rest.TransportFor(kc)
	if err != nil {
		return err
	}
	o.httpClient = &http.Client{Transport: rt, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("redirects are not followed")
	}}
	o.nodeAddr = func(name string) (string, error) {
		n, err := nodes.Get(name)
		if err != nil {
			return "", err
		}
		port := n.Status.DaemonEndpoints.KubeletEndpoint.Port
		if port == 0 {
			port = 10250
		}
		return net.JoinHostPort(name, strconv.Itoa(int(port))), nil
	}
	return nil
}

// kubeletRESTConfig copies restCfg for kubelet requests. caFile, when set, replaces the API server CA;
// otherwise the API server CA is kept. Insecure mode is always off.
func kubeletRESTConfig(restCfg *rest.Config, caFile string) *rest.Config {
	kc := rest.CopyConfig(restCfg)
	if caFile != "" {
		kc.CAFile, kc.CAData = caFile, nil
	}
	kc.Insecure = false
	return kc
}

// openSummary returns the node's Summary response body.
func (o *KubeletObserver) openSummary(ctx context.Context, node string) (io.ReadCloser, error) {
	if o.cfg.Transport == TransportAPIServerProxy {
		return o.client.Get().AbsPath("/api/v1/nodes", node, "proxy", "stats", "summary").
			Param("only_cpu_and_memory", "true").Stream(ctx)
	}
	addr, err := o.nodeAddr(node)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+addr+"/stats/summary?only_cpu_and_memory=true", nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
	e := &apierrors.StatusError{ErrStatus: metav1.Status{Code: int32(resp.StatusCode), Message: resp.Status}}
	if resp.StatusCode == http.StatusTooManyRequests {
		if ra, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
			e.ErrStatus.Reason = metav1.StatusReasonTooManyRequests
			e.ErrStatus.Details = &metav1.StatusDetails{RetryAfterSeconds: int32(ra)}
		}
	}
	return nil, e
}
