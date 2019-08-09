package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKubernetesClusters_ListClusters(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := []*KubernetesCluster{
		&KubernetesCluster{
			ID:            "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
			Name:          "blablabla",
			RegionSlug:    "nyc1",
			VersionSlug:   "1.10.0-gen0",
			ClusterSubnet: "10.244.0.0/16",
			ServiceSubnet: "10.245.0.0/16",
			IPv4:          "",
			Tags:          []string(nil),
			VPCUUID:       "880b7f98-f062-404d-b33c-458d545696f6",
			Status: &KubernetesClusterStatus{
				State: KubernetesClusterStatusRunning,
			},
			NodePools: []*KubernetesNodePool{
				{
					ID:    "1a17a012-cb31-4886-a787-deadbeef1191",
					Name:  "blablabla-1",
					Size:  "s-1vcpu-2gb",
					Count: 2,
					Nodes: []*KubernetesNode{
						{
							ID:        "",
							Name:      "",
							Status:    &KubernetesNodeStatus{},
							CreatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
							UpdatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
						},
						{
							ID:        "",
							Name:      "",
							Status:    &KubernetesNodeStatus{},
							CreatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
							UpdatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
						},
					},
				},
			},
			CreatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
			UpdatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
		},
		&KubernetesCluster{
			ID:            "deadbeef-dead-4aa5-beef-deadbeef347d",
			Name:          "antoine",
			RegionSlug:    "nyc1",
			VersionSlug:   "1.10.0-gen0",
			ClusterSubnet: "10.244.0.0/16",
			ServiceSubnet: "10.245.0.0/16",
			IPv4:          "1.2.3.4",
			VPCUUID:       "880b7f98-f062-404d-b33c-458d545696f7",
			Status: &KubernetesClusterStatus{
				State: KubernetesClusterStatusRunning,
			},
			NodePools: []*KubernetesNodePool{
				{
					ID:    "deadbeef-dead-beef-dead-deadbeefb4b3",
					Name:  "antoine-1",
					Size:  "s-1vcpu-2gb",
					Count: 5,
					Nodes: []*KubernetesNode{
						{
							ID:        "deadbeef-dead-beef-dead-deadbeefb4b1",
							Name:      "worker-393",
							Status:    &KubernetesNodeStatus{State: "running"},
							CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
							UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
						},
						{
							ID:        "deadbeef-dead-beef-dead-deadbeefb4b2",
							Name:      "worker-394",
							Status:    &KubernetesNodeStatus{State: "running"},
							CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
							UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
						},
					},
				},
			},
			CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
			UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
		},
	}
	jBlob := `
{
	"kubernetes_clusters": [
		{
			"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
			"name": "blablabla",
			"region": "nyc1",
			"version": "1.10.0-gen0",
			"cluster_subnet": "10.244.0.0/16",
			"service_subnet": "10.245.0.0/16",
			"ipv4": "",
			"tags": null,
			"vpc_uuid": "880b7f98-f062-404d-b33c-458d545696f6",
			"status": {
				"state": "running"
			},
			"node_pools": [
				{
					"id": "1a17a012-cb31-4886-a787-deadbeef1191",
					"name": "blablabla-1",
					"version": "1.10.0-gen0",
					"size": "s-1vcpu-2gb",
					"count": 2,
					"tags": null,
					"nodes": [
						{
							"id": "",
							"name": "",
							"status": {
								"state": ""
							},
							"created_at": "2018-06-21T08:44:38Z",
							"updated_at": "2018-06-21T08:44:38Z"
						},
						{
							"id": "",
							"name": "",
							"status": {
								"state": ""
							},
							"created_at": "2018-06-21T08:44:38Z",
							"updated_at": "2018-06-21T08:44:38Z"
						}
					]
				}
			],
			"created_at": "2018-06-21T08:44:38Z",
			"updated_at": "2018-06-21T08:44:38Z"
		},
		{
			"id": "deadbeef-dead-4aa5-beef-deadbeef347d",
			"name": "antoine",
			"region": "nyc1",
			"version": "1.10.0-gen0",
			"cluster_subnet": "10.244.0.0/16",
			"service_subnet": "10.245.0.0/16",
			"ipv4": "1.2.3.4",
			"tags": null,
			"status": {
				"state": "running"
			},
			"vpc_uuid": "880b7f98-f062-404d-b33c-458d545696f7",
			"node_pools": [
				{
					"id": "deadbeef-dead-beef-dead-deadbeefb4b3",
					"name": "antoine-1",
					"version": "1.10.0-gen0",
					"size": "s-1vcpu-2gb",
					"count": 5,
					"tags": null,
					"nodes": [
						{
							"id": "deadbeef-dead-beef-dead-deadbeefb4b1",
							"name": "worker-393",
							"status": {
								"state": "running"
							},
							"created_at": "2018-06-15T07:10:23Z",
							"updated_at": "2018-06-15T07:11:26Z"
						},
						{
							"id": "deadbeef-dead-beef-dead-deadbeefb4b2",
							"name": "worker-394",
								"status": {
									"state": "running"
								},
							"created_at": "2018-06-15T07:10:23Z",
							"updated_at": "2018-06-15T07:11:26Z"
						}
					]
				}
			],
			"created_at": "2018-06-15T07:10:23Z",
			"updated_at": "2018-06-15T07:11:26Z"
		}
	]
}`

	mux.HandleFunc("/v2/kubernetes/clusters", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_Get(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes
	want := &KubernetesCluster{
		ID:            "deadbeef-dead-4aa5-beef-deadbeef347d",
		Name:          "antoine",
		RegionSlug:    "nyc1",
		VersionSlug:   "1.10.0-gen0",
		ClusterSubnet: "10.244.0.0/16",
		ServiceSubnet: "10.245.0.0/16",
		IPv4:          "1.2.3.4",
		VPCUUID:       "880b7f98-f062-404d-b33c-458d545696f6",
		Status: &KubernetesClusterStatus{
			State: KubernetesClusterStatusRunning,
		},
		NodePools: []*KubernetesNodePool{
			{
				ID:    "deadbeef-dead-beef-dead-deadbeefb4b3",
				Name:  "antoine-1",
				Size:  "s-1vcpu-2gb",
				Count: 5,
				Nodes: []*KubernetesNode{
					{
						ID:        "deadbeef-dead-beef-dead-deadbeefb4b1",
						Name:      "worker-393",
						Status:    &KubernetesNodeStatus{State: "running"},
						CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
						UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
					},
					{
						ID:        "deadbeef-dead-beef-dead-deadbeefb4b2",
						Name:      "worker-394",
						Status:    &KubernetesNodeStatus{State: "running"},
						CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
						UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
					},
				},
			},
		},
		MaintenancePolicy: &KubernetesMaintenancePolicy{
			StartTime: "00:00",
			Day:       KubernetesMaintenanceDayMonday,
		},
		CreatedAt: time.Date(2018, 6, 15, 7, 10, 23, 0, time.UTC),
		UpdatedAt: time.Date(2018, 6, 15, 7, 11, 26, 0, time.UTC),
	}
	jBlob := `
{
	"kubernetes_cluster": {
		"id": "deadbeef-dead-4aa5-beef-deadbeef347d",
		"name": "antoine",
		"region": "nyc1",
		"version": "1.10.0-gen0",
		"cluster_subnet": "10.244.0.0/16",
		"service_subnet": "10.245.0.0/16",
		"ipv4": "1.2.3.4",
		"tags": null,
		"vpc_uuid": "880b7f98-f062-404d-b33c-458d545696f6",
		"status": {
			"state": "running"
		},
		"node_pools": [
			{
				"id": "deadbeef-dead-beef-dead-deadbeefb4b3",
				"name": "antoine-1",
				"version": "1.10.0-gen0",
				"size": "s-1vcpu-2gb",
				"count": 5,
				"tags": null,
				"nodes": [
					{
						"id": "deadbeef-dead-beef-dead-deadbeefb4b1",
						"name": "worker-393",
						"status": {
							"state": "running"
						},
						"created_at": "2018-06-15T07:10:23Z",
						"updated_at": "2018-06-15T07:11:26Z"
					},
					{
						"id": "deadbeef-dead-beef-dead-deadbeefb4b2",
						"name": "worker-394",
							"status": {
								"state": "running"
							},
						"created_at": "2018-06-15T07:10:23Z",
						"updated_at": "2018-06-15T07:11:26Z"
					}
				]
			}
		],
		"maintenance_policy": {
			"start_time": "00:00",
			"day": "monday"
		},
		"created_at": "2018-06-15T07:10:23Z",
		"updated_at": "2018-06-15T07:11:26Z"
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})
	got, _, err := kubeSvc.Get(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_GetKubeConfig(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes
	want := "some YAML"
	blob := []byte(want)
	mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/kubeconfig", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, want)
	})
	got, _, err := kubeSvc.GetKubeConfig(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
	require.Equal(t, blob, got.KubeconfigYAML)
}

func TestKubernetesClusters_GetUpgrades(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes
	want := []*KubernetesVersion{
		{
			Slug:              "1.12.3-do.2",
			KubernetesVersion: "1.12.3",
		},
		{
			Slug:              "1.13.1-do.1",
			KubernetesVersion: "1.13.1",
		},
	}
	jBlob := `
{
	"available_upgrade_versions": [
		{
			"slug": "1.12.3-do.2",
			"kubernetes_version": "1.12.3"
		},
		{
			"slug": "1.13.1-do.1",
			"kubernetes_version": "1.13.1"
		}
	]
}
`

	mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/upgrades", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})
	got, _, err := kubeSvc.GetUpgrades(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_Create(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesCluster{
		ID:            "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		Name:          "antoine-test-cluster",
		RegionSlug:    "s2r1",
		VersionSlug:   "1.10.0-gen0",
		ClusterSubnet: "10.244.0.0/16",
		ServiceSubnet: "10.245.0.0/16",
		Tags:          []string{"cluster-tag-1", "cluster-tag-2"},
		VPCUUID:       "880b7f98-f062-404d-b33c-458d545696f6",
		NodePools: []*KubernetesNodePool{
			&KubernetesNodePool{
				ID:    "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
				Size:  "s-1vcpu-1gb",
				Count: 2,
				Name:  "pool-a",
				Tags:  []string{"tag-1"},
			},
		},
		MaintenancePolicy: &KubernetesMaintenancePolicy{
			StartTime: "00:00",
			Day:       KubernetesMaintenanceDayMonday,
		},
	}
	createRequest := &KubernetesClusterCreateRequest{
		Name:        want.Name,
		RegionSlug:  want.RegionSlug,
		VersionSlug: want.VersionSlug,
		Tags:        want.Tags,
		VPCUUID:     want.VPCUUID,
		NodePools: []*KubernetesNodePoolCreateRequest{
			&KubernetesNodePoolCreateRequest{
				Size:  want.NodePools[0].Size,
				Count: want.NodePools[0].Count,
				Name:  want.NodePools[0].Name,
				Tags:  want.NodePools[0].Tags,
			},
		},
		MaintenancePolicy: want.MaintenancePolicy,
	}

	jBlob := `
{
	"kubernetes_cluster": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "antoine-test-cluster",
		"region": "s2r1",
		"version": "1.10.0-gen0",
		"cluster_subnet": "10.244.0.0/16",
		"service_subnet": "10.245.0.0/16",
		"tags": [
			"cluster-tag-1",
			"cluster-tag-2"
		],
		"vpc_uuid": "880b7f98-f062-404d-b33c-458d545696f6",
		"node_pools": [
			{
				"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
				"size": "s-1vcpu-1gb",
				"count": 2,
				"name": "pool-a",
				"tags": [
					"tag-1"
				]
			}
		],
		"maintenance_policy": {
			"start_time": "00:00",
			"day": "monday"
		}
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesClusterCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, createRequest)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.Create(ctx, createRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_Update(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesCluster{
		ID:            "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		Name:          "antoine-test-cluster",
		RegionSlug:    "s2r1",
		VersionSlug:   "1.10.0-gen0",
		ClusterSubnet: "10.244.0.0/16",
		ServiceSubnet: "10.245.0.0/16",
		Tags:          []string{"cluster-tag-1", "cluster-tag-2"},
		VPCUUID:       "880b7f98-f062-404d-b33c-458d545696f6",
		NodePools: []*KubernetesNodePool{
			&KubernetesNodePool{
				ID:    "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
				Size:  "s-1vcpu-1gb",
				Count: 2,
				Name:  "pool-a",
				Tags:  []string{"tag-1"},
			},
		},
		MaintenancePolicy: &KubernetesMaintenancePolicy{
			StartTime: "00:00",
			Day:       KubernetesMaintenanceDayMonday,
		},
	}
	updateRequest := &KubernetesClusterUpdateRequest{
		Name:              want.Name,
		Tags:              want.Tags,
		MaintenancePolicy: want.MaintenancePolicy,
	}

	jBlob := `
{
	"kubernetes_cluster": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "antoine-test-cluster",
		"region": "s2r1",
		"version": "1.10.0-gen0",
		"cluster_subnet": "10.244.0.0/16",
		"service_subnet": "10.245.0.0/16",
		"tags": [
			"cluster-tag-1",
			"cluster-tag-2"
		],
		"vpc_uuid": "880b7f98-f062-404d-b33c-458d545696f6",
		"node_pools": [
			{
				"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
				"size": "s-1vcpu-1gb",
				"count": 2,
				"name": "pool-a",
				"tags": [
					"tag-1"
				]
			}
		],
		"maintenance_policy": {
			"start_time": "00:00",
			"day": "monday"
		}
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesClusterUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPut)
		require.Equal(t, v, updateRequest)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.Update(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", updateRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_Upgrade(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	upgradeRequest := &KubernetesClusterUpgradeRequest{
		VersionSlug: "1.12.3-do.2",
	}

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/upgrade", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesClusterUpgradeRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, upgradeRequest)
	})

	_, err := kubeSvc.Upgrade(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", upgradeRequest)
	require.NoError(t, err)
}

func TestKubernetesClusters_Destroy(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := kubeSvc.Delete(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
}

func TestKubernetesClusters_CreateNodePool(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesNodePool{
		ID:    "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
		Size:  "s-1vcpu-1gb",
		Count: 2,
		Name:  "pool-a",
		Tags:  []string{"tag-1"},
	}
	createRequest := &KubernetesNodePoolCreateRequest{
		Size:  want.Size,
		Count: want.Count,
		Name:  want.Name,
		Tags:  want.Tags,
	}

	jBlob := `
{
	"node_pool": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
		"size": "s-1vcpu-1gb",
		"count": 2,
		"name": "pool-a",
		"tags": [
			"tag-1"
		]
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/node_pools", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesNodePoolCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, createRequest)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.CreateNodePool(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", createRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_GetNodePool(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesNodePool{
		ID:    "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
		Size:  "s-1vcpu-1gb",
		Count: 2,
		Name:  "pool-a",
		Tags:  []string{"tag-1"},
	}

	jBlob := `
{
	"node_pool": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8a",
		"size": "s-1vcpu-1gb",
		"count": 2,
		"name": "pool-a",
		"tags": [
			"tag-1"
		]
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/node_pools/8d91899c-0739-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.GetNodePool(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", "8d91899c-0739-4a1a-acc5-deadbeefbb8a")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_ListNodePools(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := []*KubernetesNodePool{
		{
			ID:    "1a17a012-cb31-4886-a787-deadbeef1191",
			Name:  "blablabla-1",
			Size:  "s-1vcpu-2gb",
			Count: 2,
			Nodes: []*KubernetesNode{
				{
					ID:        "",
					Name:      "",
					Status:    &KubernetesNodeStatus{},
					CreatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
					UpdatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
				},
				{
					ID:        "",
					Name:      "",
					Status:    &KubernetesNodeStatus{},
					CreatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
					UpdatedAt: time.Date(2018, 6, 21, 8, 44, 38, 0, time.UTC),
				},
			},
		},
	}
	jBlob := `
{
	"node_pools": [
		{
			"id": "1a17a012-cb31-4886-a787-deadbeef1191",
			"name": "blablabla-1",
			"version": "1.10.0-gen0",
			"size": "s-1vcpu-2gb",
			"count": 2,
			"tags": null,
			"nodes": [
				{
					"id": "",
					"name": "",
					"status": {
						"state": ""
					},
					"created_at": "2018-06-21T08:44:38Z",
					"updated_at": "2018-06-21T08:44:38Z"
				},
				{
					"id": "",
					"name": "",
					"status": {
						"state": ""
					},
					"created_at": "2018-06-21T08:44:38Z",
					"updated_at": "2018-06-21T08:44:38Z"
				}
			]
		}
	]
}`

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/node_pools", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.ListNodePools(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_UpdateNodePool(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesNodePool{
		ID:    "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a",
		Name:  "a better name",
		Size:  "s-1vcpu-1gb",
		Count: 4,
		Tags:  []string{"tag-1", "tag-2"},
	}
	updateRequest := &KubernetesNodePoolUpdateRequest{
		Name:  "a better name",
		Count: 4,
		Tags:  []string{"tag-1", "tag-2"},
	}

	jBlob := `
{
	"node_pool": {
		"id": "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a",
		"size": "s-1vcpu-1gb",
		"count": 4,
		"name": "a better name",
		"tags": [
			"tag-1", "tag-2"
		]
	}
}`

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesNodePoolUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPut)
		require.Equal(t, v, updateRequest)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.UpdateNodePool(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", updateRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestKubernetesClusters_DeleteNodePool(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := kubeSvc.DeleteNodePool(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a")
	require.NoError(t, err)
}

func TestKubernetesClusters_DeleteNode(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		setup()
		defer teardown()
		kubeSvc := client.Kubernetes

		mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a/nodes/8d91899c-node-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			require.Equal(t, "", r.URL.Query().Encode())
		})

		_, err := kubeSvc.DeleteNode(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", "8d91899c-node-4a1a-acc5-deadbeefbb8a", nil)
		require.NoError(t, err)
	})

	t.Run("drain", func(t *testing.T) {
		setup()
		defer teardown()
		kubeSvc := client.Kubernetes

		mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a/nodes/8d91899c-node-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			require.Equal(t, "skip_drain=1", r.URL.Query().Encode())
		})

		_, err := kubeSvc.DeleteNode(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", "8d91899c-node-4a1a-acc5-deadbeefbb8a", &KubernetesNodeDeleteRequest{
			SkipDrain: true,
		})
		require.NoError(t, err)
	})

	t.Run("replace", func(t *testing.T) {
		setup()
		defer teardown()
		kubeSvc := client.Kubernetes

		mux.HandleFunc("/v2/kubernetes/clusters/deadbeef-dead-4aa5-beef-deadbeef347d/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a/nodes/8d91899c-node-4a1a-acc5-deadbeefbb8a", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodDelete)
			require.Equal(t, "replace=1", r.URL.Query().Encode())
		})

		_, err := kubeSvc.DeleteNode(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", "8d91899c-node-4a1a-acc5-deadbeefbb8a", &KubernetesNodeDeleteRequest{
			Replace: true,
		})
		require.NoError(t, err)
	})
}

func TestKubernetesClusters_RecycleNodePoolNodes(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	recycleRequest := &KubernetesNodePoolRecycleNodesRequest{
		Nodes: []string{"node1", "node2"},
	}

	mux.HandleFunc("/v2/kubernetes/clusters/8d91899c-0739-4a1a-acc5-deadbeefbb8f/node_pools/8d91899c-nodepool-4a1a-acc5-deadbeefbb8a/recycle", func(w http.ResponseWriter, r *http.Request) {
		v := new(KubernetesNodePoolRecycleNodesRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, recycleRequest)
	})

	_, err := kubeSvc.RecycleNodePoolNodes(ctx, "8d91899c-0739-4a1a-acc5-deadbeefbb8f", "8d91899c-nodepool-4a1a-acc5-deadbeefbb8a", recycleRequest)
	require.NoError(t, err)
}

func TestKubernetesVersions_List(t *testing.T) {
	setup()
	defer teardown()

	kubeSvc := client.Kubernetes

	want := &KubernetesOptions{
		Versions: []*KubernetesVersion{
			{Slug: "1.10.0-gen0", KubernetesVersion: "1.10.0"},
		},
		Regions: []*KubernetesRegion{
			{Name: "New York 3", Slug: "nyc3"},
		},
		Sizes: []*KubernetesNodeSize{
			{Name: "c-8", Slug: "c-8"},
		},
	}
	jBlob := `
{
	"options": {
		"versions": [
			{
				"slug": "1.10.0-gen0",
				"kubernetes_version": "1.10.0"
			}
		],
		"regions": [
			{
				"name": "New York 3",
				"slug": "nyc3"
			}
		],
		"sizes": [
			{
				"name": "c-8",
				"slug": "c-8"
			}
		]
	}
}`

	mux.HandleFunc("/v2/kubernetes/options", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	got, _, err := kubeSvc.GetOptions(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

var maintenancePolicyDayTests = []struct {
	name  string
	json  string
	day   KubernetesMaintenancePolicyDay
	valid bool
}{
	{
		name:  "sunday",
		day:   KubernetesMaintenanceDaySunday,
		json:  `"sunday"`,
		valid: true,
	},

	{
		name:  "any",
		day:   KubernetesMaintenanceDayAny,
		json:  `"any"`,
		valid: true,
	},

	{
		name:  "invalid",
		day:   100, // invalid input
		json:  `"invalid weekday (100)"`,
		valid: false,
	},
}

func TestWeekday_UnmarshalJSON(t *testing.T) {
	for _, ts := range maintenancePolicyDayTests {
		t.Run(ts.name, func(t *testing.T) {
			var got KubernetesMaintenancePolicyDay
			err := json.Unmarshal([]byte(ts.json), &got)
			valid := err == nil
			if valid != ts.valid {
				t.Errorf("valid unmarshal case\n\tgot: %+v\n\twant : %+v", valid, ts.valid)
			}

			if valid && got != ts.day {
				t.Errorf("\ninput: %s\ngot : %+v\nwant  : %+v\n",
					ts.day, got, ts.day)
			}
		})
	}
}

func TestWeekday_MarshalJSON(t *testing.T) {
	for _, ts := range maintenancePolicyDayTests {
		t.Run(ts.name, func(t *testing.T) {
			out, err := json.Marshal(ts.day)
			valid := err == nil
			if valid != ts.valid {
				t.Errorf("valid marshal case\n\tgot: %+v\n\twant : %+v", valid, ts.valid)
			}

			if valid && ts.json != string(out) {
				t.Errorf("\ninput: %s\ngot : %+v\nwant  : %+v\n",
					ts.day, string(out), ts.json)
			}
		})
	}
}
