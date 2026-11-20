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

package sakuracloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeAPI is a minimal fake of the SAKURA cloud IaaS API.
type fakeAPI struct {
	t      *testing.T
	server *httptest.Server

	mu       sync.Mutex
	lastUser string
	lastPass string

	// scripted behavior
	diskStates    []string // sequence of availability states returned by GET /disk/:id
	diskStateIdx  int
	powerOffCode  int // status for DELETE /server/:id/power (0 = 200)
	serverStatus  string
	createdDisks  []map[string]interface{}
	createdServer map[string]interface{}
	deletedServer bool
}

func newFakeAPI(t *testing.T) *fakeAPI {
	f := &fakeAPI{t: t, diskStates: []string{"available"}, serverStatus: "down"}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	return f
}

func (f *fakeAPI) close() { f.server.Close() }

func (f *fakeAPI) manager() *sakuracloudManager {
	return &sakuracloudManager{
		token:      "token",
		secret:     "secret",
		zone:       "is1a",
		apiBase:    f.server.URL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		nodeGroups: map[string]*sakuracloudNodeGroup{},
	}
}

func (f *fakeAPI) handle(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastUser, f.lastPass, _ = r.BasicAuth()

	var body map[string]interface{}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}

	path := r.URL.Path
	switch {
	case r.Method == http.MethodGet && path == "/server":
		json.NewEncoder(w).Encode(map[string]interface{}{"Servers": []interface{}{}})
	case r.Method == http.MethodPost && path == "/disk":
		f.createdDisks = append(f.createdDisks, body)
		json.NewEncoder(w).Encode(map[string]interface{}{"Disk": map[string]interface{}{"ID": "201"}})
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/disk/"):
		state := f.diskStates[min(f.diskStateIdx, len(f.diskStates)-1)]
		f.diskStateIdx++
		json.NewEncoder(w).Encode(map[string]interface{}{"Disk": map[string]interface{}{"Availability": state}})
	case r.Method == http.MethodPost && path == "/server":
		f.createdServer = body
		json.NewEncoder(w).Encode(map[string]interface{}{"Server": map[string]interface{}{"ID": "301"}})
	case r.Method == http.MethodPut && strings.Contains(path, "/to/server/"):
		json.NewEncoder(w).Encode(map[string]interface{}{"Success": true})
	case r.Method == http.MethodPut && strings.HasSuffix(path, "/config"):
		json.NewEncoder(w).Encode(map[string]interface{}{"Success": true})
	case r.Method == http.MethodPut && strings.HasSuffix(path, "/power"):
		json.NewEncoder(w).Encode(map[string]interface{}{"Success": true})
	case r.Method == http.MethodDelete && strings.HasSuffix(path, "/power"):
		if f.powerOffCode != 0 {
			w.WriteHeader(f.powerOffCode)
			fmt.Fprint(w, `{"error_code":"conflict"}`)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"Success": true})
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/server/"):
		json.NewEncoder(w).Encode(map[string]interface{}{"Server": map[string]interface{}{
			"ID": "301", "Name": "pool-abc12", "Instance": map[string]interface{}{"Status": f.serverStatus},
		}})
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/server/"):
		f.deletedServer = true
		json.NewEncoder(w).Encode(map[string]interface{}{"Success": true})
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error_code":"not_found"}`)
	}
}

func testGroupConfig() *nodeGroupConfig {
	return &nodeGroupConfig{
		MinSize: 0, MaxSize: 2, Core: 2, MemoryGB: 4, DiskGB: 20,
		SourceArchiveID: "111", StartupNoteID: "112",
	}
}

func TestCreateServer(t *testing.T) {
	f := newFakeAPI(t)
	defer f.close()
	m := f.manager()

	err := m.createServer("pool", testGroupConfig())
	assert.NoError(t, err)

	// The server plan must be requested by CPU/memory spec, not by plan ID:
	// the plan-ID form is rejected with 400 by the live API.
	serverBody := f.createdServer["Server"].(map[string]interface{})
	plan := serverBody["ServerPlan"].(map[string]interface{})
	assert.Equal(t, float64(2), plan["CPU"])
	assert.Equal(t, float64(4096), plan["MemoryMB"])
	_, hasID := plan["ID"]
	assert.False(t, hasID)
	assert.Contains(t, serverBody["Tags"], groupTagPrefix+"pool")

	diskBody := f.createdDisks[0]["Disk"].(map[string]interface{})
	assert.Equal(t, float64(20*1024), diskBody["SizeMB"])
}

func TestCreateServerWaitsForDiskAfterConfig(t *testing.T) {
	f := newFakeAPI(t)
	defer f.close()
	// First wait (after create): available. Second wait (after config): first
	// migrating, then available. Powering on while the disk is still
	// modifying fails with 409 on the live API.
	f.diskStates = []string{"available", "migrating", "available"}
	m := f.manager()

	err := m.createServer("pool", testGroupConfig())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, f.diskStateIdx, 3)
}

func TestDeleteServerForcesPowerOffAndToleratesConflict(t *testing.T) {
	f := newFakeAPI(t)
	defer f.close()
	// The server list response does not reliably include the power state, so
	// delete always force-powers-off first and must tolerate 409 (already
	// down).
	f.powerOffCode = http.StatusConflict
	f.serverStatus = "down"
	m := f.manager()

	err := m.deleteServer(&sakuraServer{ID: "301", Name: "pool-abc12"})
	assert.NoError(t, err)
	assert.True(t, f.deletedServer)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
