/*
Copyright 2021 The Kubernetes Authors.

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

package rancher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterByID(t *testing.T) {
	clusterID := "c-lkf2d"
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"id": "c-lkf2d",
				"name": "c-lkf2d"
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		cluster, err := cli.ClusterByID(clusterID)
		assert.NoError(t, err, "unexpected error")
		assert.Equal(t, cluster.ID, clusterID)
	})

	t.Run("failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.ClusterByID(clusterID)
		assert.Error(t, err)
	})
}

func TestResizeNodePool(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"id": "c-wxas29:master",
				"quantity": 2
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		qty := 2
		np, err := cli.ResizeNodePool("asd", qty)
		assert.NoError(t, err)
		assert.Equal(t, np.Quantity, qty)
	})

	t.Run("failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.ResizeNodePool("asd", 2)
		assert.Error(t, err)
	})
}

func TestNodePoolsByCluster(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"data": [{"id":"nodepool1"},{"id":"nodepool2"}]
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		nps, err := cli.NodePoolsByCluster("cluster1")
		assert.NoError(t, err)
		assert.Equal(t, len(nps), 2)
	})

	t.Run("failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"data": []
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		nps, err := cli.NodePoolsByCluster("cluster1")
		assert.Error(t, err)
		assert.Equal(t, len(nps), 0)
	})
}

func TestNodePoolByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"id": "nodepool1"
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		np, err := cli.NodePoolByID("nodepool1")
		assert.NoError(t, err)
		assert.Equal(t, np.ID, "nodepool1")
	})

	t.Run("failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.NodePoolByID("nodepool1")
		assert.Error(t, err)
	})
}

func TestNodeByProviderID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"data": [{"id": "node1"}]
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		n, err := cli.NodeByProviderID("gce_uuid")
		assert.NoError(t, err)
		assert.Equal(t, n.ID, "node1")
	})

	t.Run("failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `
			{
				"data": []
			}
			`
			fmt.Fprintln(w, response)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.NodeByProviderID("gce_uuid")
		assert.Error(t, err)
	})
}

func Test_doRequest(t *testing.T) {
	t.Run("failed with 401", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.doRequest("GET", ts.URL, nil, nil)
		assert.Error(t, err)
	})

	t.Run("failed with 403", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.doRequest("GET", ts.URL, nil, nil)
		assert.Error(t, err)
	})

	t.Run("failed with 404", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		defer ts.Close()
		cli := New(ts.URL, "")
		_, err := cli.doRequest("GET", ts.URL, nil, nil)
		assert.Error(t, err)
	})
}
