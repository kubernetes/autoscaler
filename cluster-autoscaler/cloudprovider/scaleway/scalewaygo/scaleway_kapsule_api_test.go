/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaygo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testClusterID = "a451feb1-e0bc-48ff-b31e-a09bbdacdb1d"
	testSecretKey = "4e713ead-7f76-4a8a-8774-0ac9b8fffb5c"
	testRegion    = "fr-par"
)

// createTestPool creates a test Pool with sensible defaults
func createTestPool(overrides ...func(*Pool)) Pool {
	now := time.Now()
	pool := Pool{
		ID:        "pool-1",
		ClusterID: testClusterID,
		Name:      "default",
		Size:      3,
		MinSize:   1,
		MaxSize:   10,
		Status:    PoolStatusReady,
		CreatedAt: &now,
	}
	for _, override := range overrides {
		override(&pool)
	}
	return pool
}

// createTestNode creates a test Node with sensible defaults
func createTestNode(overrides ...func(*Node)) Node {
	now := time.Now()
	node := Node{
		ID:         "node-1",
		PoolID:     "pool-1",
		ClusterID:  testClusterID,
		ProviderID: "instance-1",
		Name:       "node-1",
		Status:     NodeStatusReady,
		CreatedAt:  &now,
	}
	for _, override := range overrides {
		override(&node)
	}
	return node
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "valid config",
			cfg: Config{
				ClusterID: testClusterID,
				SecretKey: testSecretKey,
				Region:    testRegion,
			},
			wantErr: nil,
		},
		{
			name: "missing cluster ID",
			cfg: Config{
				SecretKey: testSecretKey,
				Region:    testRegion,
			},
			wantErr: ErrMissingClusterID,
		},
		{
			name: "missing secret key",
			cfg: Config{
				ClusterID: testClusterID,
				Region:    testRegion,
			},
			wantErr: ErrMissingSecretKey,
		},
		{
			name: "missing region",
			cfg: Config{
				ClusterID: testClusterID,
				SecretKey: testSecretKey,
			},
			wantErr: ErrMissingRegion,
		},
		{
			name: "custom API URL",
			cfg: Config{
				ClusterID: testClusterID,
				SecretKey: testSecretKey,
				Region:    testRegion,
				ApiUrl:    "https://custom.api.com",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)

			// Verify default API URL is set when not provided
			if tt.cfg.ApiUrl == "" {
				assert.Equal(t, defaultApiURL, client.ApiURL())
			} else {
				assert.Equal(t, tt.cfg.ApiUrl, client.ApiURL())
			}

			// Verify token and region are set correctly
			assert.Equal(t, tt.cfg.SecretKey, client.Token())
			assert.Equal(t, tt.cfg.Region, client.Region())
		})
	}
}

func TestScalewayRequest_getURL(t *testing.T) {
	tests := []struct {
		name    string
		req     *scalewayRequest
		apiURL  string
		want    string
		wantErr bool
	}{
		{
			name: "simple path",
			req: &scalewayRequest{
				Path: "/k8s/v1/regions/fr-par/clusters",
			},
			apiURL:  "https://api.scaleway.com",
			want:    "https://api.scaleway.com/k8s/v1/regions/fr-par/clusters",
			wantErr: false,
		},
		{
			name: "path with query parameters",
			req: &scalewayRequest{
				Path: "/k8s/v1/regions/fr-par/clusters",
				Query: url.Values{
					"page":      []string{"1"},
					"page_size": []string{"100"},
				},
			},
			apiURL:  "https://api.scaleway.com",
			want:    "https://api.scaleway.com/k8s/v1/regions/fr-par/clusters?page=1&page_size=100",
			wantErr: false,
		},
		{
			name: "invalid base URL",
			req: &scalewayRequest{
				Path: "/test",
			},
			apiURL:  "://invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.req.getURL(tt.apiURL)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestClient_ListPools(t *testing.T) {
	t.Run("successful single page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.NotEmpty(t, r.Header.Get("X-Auth-Token"))

			pool := createTestPool()
			resp := ListPoolsResponse{
				Pools:      []Pool{pool},
				TotalCount: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "max-age=30")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		cacheControl, pools, err := client.ListPools(context.Background(), testClusterID)

		require.NoError(t, err)
		assert.Len(t, pools, 1)
		assert.Equal(t, "pool-1", pools[0].ID)
		assert.Equal(t, 3, pools[0].Size)
		assert.Equal(t, PoolStatusReady, pools[0].Status)
		assert.Equal(t, 30*time.Second, cacheControl)
	})

	t.Run("multiple pages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			var pools []Pool

			if page == "1" {
				// Return full page to trigger next page fetch
				for i := range pageSizeListPools {
					pools = append(pools, createTestPool(func(p *Pool) {
						p.ID = fmt.Sprintf("pool-%d", i)
						p.Name = fmt.Sprintf("pool-%d", i)
					}))
				}
			}

			resp := ListPoolsResponse{
				Pools:      pools,
				TotalCount: uint64(len(pools)),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, pools, err := client.ListPools(context.Background(), testClusterID)

		require.NoError(t, err)
		assert.Equal(t, pageSizeListPools, len(pools))
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, _, err := client.ListPools(context.Background(), testClusterID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrServerSide))
	})
}

func TestClient_UpdatePool(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)

			var req UpdatePoolRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, 5, req.Size)

			now := time.Now()
			resp := createTestPool(func(p *Pool) {
				p.Size = 5
				p.Status = PoolStatusScaling
				p.UpdatedAt = &now
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		pool, err := client.UpdatePool(context.Background(), "pool-1", 5)

		require.NoError(t, err)
		assert.Equal(t, "pool-1", pool.ID)
		assert.Equal(t, 5, pool.Size)
		assert.Equal(t, PoolStatusScaling, pool.Status)
	})

	t.Run("empty pool ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not call server with empty pool ID")
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, err := client.UpdatePool(context.Background(), "", 5)

		assert.Error(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, err := client.UpdatePool(context.Background(), "pool-1", 5)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrClientSide))
	})
}

func TestClient_ListNodes(t *testing.T) {
	t.Run("successful single page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)

			node := createTestNode()
			resp := ListNodesResponse{
				Nodes:      []Node{node},
				TotalCount: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "max-age=15")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		cacheControl, nodes, err := client.ListNodes(context.Background(), testClusterID)

		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		assert.Equal(t, "node-1", nodes[0].ID)
		assert.Equal(t, NodeStatusReady, nodes[0].Status)
		assert.Equal(t, 15*time.Second, cacheControl)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, _, err := client.ListNodes(context.Background(), testClusterID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrServerSide))
	})
}

func TestClient_DeleteNode(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)

			now := time.Now()
			resp := createTestNode(func(n *Node) {
				n.Status = NodeStatusDeleting
				n.UpdatedAt = &now
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		node, err := client.DeleteNode(context.Background(), "node-1")

		require.NoError(t, err)
		assert.Equal(t, "node-1", node.ID)
		assert.Equal(t, NodeStatusDeleting, node.Status)
	})

	t.Run("empty node ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not call server with empty node ID")
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, err := client.DeleteNode(context.Background(), "")

		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, err := client.DeleteNode(context.Background(), "node-nonexistent")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrClientSide))
	})
}

func TestClient_cacheControl(t *testing.T) {
	tests := []struct {
		name             string
		headerValue      string
		defaultCache     time.Duration
		expectedDuration time.Duration
	}{
		{
			name:             "valid max-age",
			headerValue:      "max-age=30",
			defaultCache:     10 * time.Second,
			expectedDuration: 30 * time.Second,
		},
		{
			name:             "max-age with other directives",
			headerValue:      "public, max-age=60, must-revalidate",
			defaultCache:     10 * time.Second,
			expectedDuration: 60 * time.Second,
		},
		{
			name:             "no max-age",
			headerValue:      "no-cache",
			defaultCache:     15 * time.Second,
			expectedDuration: 15 * time.Second,
		},
		{
			name:             "empty header",
			headerValue:      "",
			defaultCache:     20 * time.Second,
			expectedDuration: 20 * time.Second,
		},
		{
			name:             "invalid max-age format",
			headerValue:      "max-age=",
			defaultCache:     25 * time.Second,
			expectedDuration: 25 * time.Second,
		},
		{
			name:             "invalid max-age value",
			headerValue:      "max-age=invalid",
			defaultCache:     30 * time.Second,
			expectedDuration: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				defaultCacheControl: tt.defaultCache,
			}

			header := http.Header{}
			if tt.headerValue != "" {
				header.Set("Cache-Control", tt.headerValue)
			}

			got := c.cacheControl(header)
			assert.Equal(t, tt.expectedDuration, got)
		})
	}
}

func TestClient_do(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not call server with nil request")
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		var resp map[string]interface{}
		_, err := client.do(context.Background(), nil, &resp)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request must be non-nil")
	})

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		var resp map[string]interface{}
		_, err := client.do(context.Background(), &scalewayRequest{
			Method: "GET",
			Path:   "/test",
		}, &resp)

		require.NoError(t, err)
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("wrong content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("plain text"))
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		var resp map[string]interface{}
		_, err := client.do(context.Background(), &scalewayRequest{
			Method: "GET",
			Path:   "/test",
		}, &resp)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected content-type")
	})

	t.Run("client error 4xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		var resp map[string]interface{}
		_, err := client.do(context.Background(), &scalewayRequest{
			Method: "GET",
			Path:   "/test",
		}, &resp)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrClientSide))
	})

	t.Run("server error 5xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		var resp map[string]interface{}
		_, err := client.do(context.Background(), &scalewayRequest{
			Method: "GET",
			Path:   "/test",
		}, &resp)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrServerSide))
	})
}

func TestListPoolsPaginated(t *testing.T) {
	t.Run("empty cluster ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not call server with empty cluster ID")
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, _, err := client.listPoolsPaginated(context.Background(), "", 1, 100)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clusterID cannot be empty")
	})
}

func TestListNodesPaginated(t *testing.T) {
	t.Run("empty cluster ID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not call server with empty cluster ID")
		}))
		defer server.Close()

		client := createTestClient(t, server.URL)
		_, _, err := client.listNodesPaginated(context.Background(), "", 1, 100)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clusterID cannot be empty")
	})
}

// createTestClient is a helper function to create a test client
func createTestClient(t *testing.T, apiURL string) *client {
	t.Helper()

	client, err := NewClient(Config{
		ClusterID: testClusterID,
		SecretKey: testSecretKey,
		Region:    testRegion,
		ApiUrl:    apiURL,
	})
	require.NoError(t, err)
	return client
}
