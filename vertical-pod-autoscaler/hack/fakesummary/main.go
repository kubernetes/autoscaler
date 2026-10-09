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

// fakesummary serves kubelet /stats/summary for KWOK nodes (AEP-9926 PoC harness, not for upstream).
//
// Each KWOK node's InternalIP is a distinct loopback address (127.1.x.y). The server listens once on
// 0.0.0.0:<port> and identifies the node by the local address the client dialled. Pods come from the
// API server. Per-node behaviour is scripted over a plain-HTTP control port:
//
//	POST /control {"node":"kwok-node-7","mode":"stall","latencyMs":20}
//	GET  /fetches  -> per-node fetch timestamps (for coverage / revisit measurement)
//
// modes: quiet, stall, 429, timeout, oversized, malformed, stale, future, nanoseconds, nopsi
package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type nodeCfg struct {
	Mode      string `json:"mode"`
	LatencyMs int    `json:"latencyMs"`
}

type server struct {
	mu        sync.Mutex
	ipToNode  map[string]string
	cfg       map[string]nodeCfg
	defCfg    nodeCfg
	stallBase map[string]float64 // accumulated stall seconds per container key
	lastSeen  map[string]time.Time
	fetches   map[string][]time.Time
	pods      cache.Indexer
}

func (s *server) summary(w http.ResponseWriter, r *http.Request) {
	local, _ := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	host, _, _ := net.SplitHostPort(local.String())
	s.mu.Lock()
	node := s.ipToNode[host]
	c, ok := s.cfg[node]
	if !ok {
		c = s.defCfg
	}
	s.fetches[node] = append(s.fetches[node], time.Now())
	s.mu.Unlock()
	if node == "" {
		http.Error(w, "unknown node for "+host, 404)
		return
	}
	time.Sleep(time.Duration(c.LatencyMs) * time.Millisecond)
	switch c.Mode {
	case "429":
		w.Header().Set("Retry-After", "30")
		http.Error(w, "throttled", http.StatusTooManyRequests)
		return
	case "timeout":
		time.Sleep(30 * time.Second)
	case "malformed":
		_, _ = w.Write([]byte(`{"pods":[{"podRef":`))
		return
	case "oversized":
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pods":[],"pad":"`))
		chunk := strings.Repeat("x", 1<<20)
		for range 20 {
			_, _ = w.Write([]byte(chunk))
		}
		_, _ = w.Write([]byte(`"}`))
		return
	default:
	}
	now := time.Now()
	objs, _ := s.pods.ByIndex("node", node)
	type psi struct {
		Some struct {
			Total uint64 `json:"total"`
		} `json:"some"`
		Full struct {
			Total uint64 `json:"total"`
		} `json:"full"`
	}
	type mem struct {
		Time            metav1.Time `json:"time"`
		WorkingSetBytes uint64      `json:"workingSetBytes"`
		PSI             *psi        `json:"psi,omitempty"`
	}
	type ctr struct {
		Name      string      `json:"name"`
		StartTime metav1.Time `json:"startTime"`
		Memory    mem         `json:"memory"`
	}
	type pod struct {
		PodRef struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			UID       string `json:"uid"`
		} `json:"podRef"`
		Containers []ctr `json:"containers"`
	}
	out := struct {
		Pods []pod `json:"pods"`
	}{}
	s.mu.Lock()
	for _, o := range objs {
		p := o.(*corev1.Pod)
		var pp pod
		pp.PodRef.Name, pp.PodRef.Namespace, pp.PodRef.UID = p.Name, p.Namespace, string(p.UID)
		for _, cs := range p.Spec.Containers {
			key := string(p.UID) + "/" + cs.Name
			limit := cs.Resources.Limits.Memory().Value()
			last, seen := s.lastSeen[key]
			if !seen {
				last = now
			}
			stallFrac, wsFrac := 0.0, 0.30
			if c.Mode == "stall" || c.Mode == "nanoseconds" || c.Mode == "stale" || c.Mode == "future" {
				stallFrac, wsFrac = 0.60, 0.95
			}
			s.stallBase[key] += stallFrac * now.Sub(last).Seconds()
			s.lastSeen[key] = now
			t := now
			switch c.Mode {
			case "stale":
				t = now.Add(-10 * time.Minute)
			case "future":
				t = now.Add(10 * time.Minute)
			default:
			}
			unit := 1e6 // microseconds
			if c.Mode == "nanoseconds" {
				unit = 1e9
			}
			m := mem{Time: metav1.NewTime(t), WorkingSetBytes: uint64(wsFrac * float64(limit))}
			if c.Mode != "nopsi" {
				m.PSI = &psi{}
				m.PSI.Some.Total = uint64(s.stallBase[key] * unit)
				m.PSI.Full.Total = m.PSI.Some.Total
			}
			start := p.CreationTimestamp
			pp.Containers = append(pp.Containers, ctr{Name: cs.Name, StartTime: start, Memory: m})
		}
		out.Pods = append(out.Pods, pp)
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *server) control(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Node string `json:"node"`
		nodeCfg
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	if req.Node == "*" {
		s.defCfg = req.nodeCfg
		s.cfg = map[string]nodeCfg{}
	} else {
		s.cfg[req.Node] = req.nodeCfg
	}
	s.mu.Unlock()
	_, _ = fmt.Fprintln(w, "ok")
}

func (s *server) fetchLog(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = json.NewEncoder(w).Encode(s.fetches)
}

func (s *server) reset(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	s.fetches = map[string][]time.Time{}
	s.mu.Unlock()
	_, _ = fmt.Fprintln(w, "ok")
}

// ca issues a leaf certificate for whichever local address the client dialled, so a client that verifies
// against the CA also verifies the node IP it connected to.
type ca struct {
	cert     *x509.Certificate
	key      *rsa.PrivateKey
	der      []byte
	mu       sync.Mutex
	leaf     map[string]*tls.Certificate
	nodeName func(ip string) string
}

func newCA() *ca {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "fakesummary-ca"}, IsCA: true,
		BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour)}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	c, _ := x509.ParseCertificate(der)
	return &ca{cert: c, key: key, der: der, leaf: map[string]*tls.Certificate{}}
}

// names returns the node name for an IP, so leaf certs carry it as a DNS SAN (clients verify by node name).
func (c *ca) names(ip string) []string {
	if c.nodeName == nil {
		return nil
	}
	if n := c.nodeName(ip); n != "" {
		return []string{n}
	}
	return nil
}

func (c *ca) forIP(ip string) *tls.Certificate {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.leaf[ip]; ok {
		return l
	}
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: ip},
		IPAddresses: []net.IP{net.ParseIP(ip)}, DNSNames: c.names(ip), ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageDigitalSignature, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour)}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, c.cert, &key.PublicKey, c.key)
	l := &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	c.leaf[ip] = l
	return l
}

func main() {
	kubeconfig := flag.String("kubeconfig", "", "")
	port := flag.Int("port", 10250, "")
	ctlPort := flag.Int("control-port", 18080, "")
	caOut := flag.String("ca-out", "fakesummary-ca.pem", "where to write the CA certificate")
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	cs := kubernetes.NewForConfigOrDie(cfg)
	f := informers.NewSharedInformerFactory(cs, 0)
	pi := f.Core().V1().Pods().Informer()
	_ = pi.AddIndexers(cache.Indexers{"node": func(o any) ([]string, error) { return []string{o.(*corev1.Pod).Spec.NodeName}, nil }})
	ni := f.Core().V1().Nodes().Informer()
	s := &server{ipToNode: map[string]string{}, cfg: map[string]nodeCfg{}, defCfg: nodeCfg{Mode: "quiet"},
		stallBase: map[string]float64{}, lastSeen: map[string]time.Time{}, fetches: map[string][]time.Time{}, pods: pi.GetIndexer()}
	onNode := func(o any) {
		n := o.(*corev1.Node)
		for _, a := range n.Status.Addresses {
			if a.Type == corev1.NodeInternalIP {
				s.mu.Lock()
				s.ipToNode[a.Address] = n.Name
				s.mu.Unlock()
			}
		}
	}
	_, _ = ni.AddEventHandler(cache.ResourceEventHandlerFuncs{AddFunc: onNode, UpdateFunc: func(_, o any) { onNode(o) }})
	ctx := context.Background()
	f.Start(ctx.Done())
	f.WaitForCacheSync(ctx.Done())
	mux := http.NewServeMux()
	mux.HandleFunc("/stats/summary", s.summary)
	authority := newCA()
	authority.nodeName = func(ip string) string { s.mu.Lock(); defer s.mu.Unlock(); return s.ipToNode[ip] }
	_ = os.WriteFile(*caOut, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: authority.der}), 0o644)
	srv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", *port), Handler: mux, TLSConfig: &tls.Config{
		GetCertificate: func(h *tls.ClientHelloInfo) (*tls.Certificate, error) {
			host, _, _ := net.SplitHostPort(h.Conn.LocalAddr().String())
			return authority.forIP(host), nil
		}}}
	cm := http.NewServeMux()
	cm.HandleFunc("/control", s.control)
	cm.HandleFunc("/fetches", s.fetchLog)
	cm.HandleFunc("/reset", s.reset)
	go func() { _ = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *ctlPort), cm) }()
	fmt.Println("fakesummary listening", *port, "control", *ctlPort)
	panic(srv.ListenAndServeTLS("", ""))
}
