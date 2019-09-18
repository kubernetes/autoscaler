package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var vTestObj = &VPC{
	ID:         "880b7f98-f062-404d-b33c-458d545696f6",
	Name:       "my-new-vpc",
	RegionSlug: "s2r7",
	CreatedAt:  time.Date(2019, 2, 4, 21, 48, 40, 995304079, time.UTC),
	Default:    false,
}

var vTestJSON = `
    {
      "id":"880b7f98-f062-404d-b33c-458d545696f6",
      "name":"my-new-vpc",
      "region":"s2r7",
      "created_at":"2019-02-04T21:48:40.995304079Z",
      "default":false
    }
`

func TestVPCs_Get(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_List(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := []*VPC{
		vTestObj,
	}
	links := &Links{
		Pages: &Pages{
			Last: "http://localhost/v2/vpcs?page=3&per_page=1",
			Next: "http://localhost/v2/vpcs?page=2&per_page=1",
		},
	}
	jsonBlob := `
{
  "vpcs": [
` + vTestJSON + `
  ],
  "links": {
    "pages": {
      "last": "http://localhost/v2/vpcs?page=3&per_page=1",
      "next": "http://localhost/v2/vpcs?page=2&per_page=1"
    }
  },
  "meta": {"total": 3}
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jsonBlob)
	})

	got, resp, err := svc.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
	require.Equal(t, resp.Links, links)
}

func TestVPCs_Create(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	req := &VPCCreateRequest{
		Name:       "my-new-vpc",
		RegionSlug: "s2r7",
	}
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCCreateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Create(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_Update(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	req := &VPCUpdateRequest{
		Name: "my-new-vpc",
	}
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		c := new(VPCUpdateRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPut)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Update(ctx, id, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_Set(t *testing.T) {
	setup()
	defer teardown()

	type setRequest struct {
		Name string `json:"name"`
	}

	svc := client.VPCs
	path := "/v2/vpcs"
	want := vTestObj
	id := "880b7f98-f062-404d-b33c-458d545696f6"
	name := "my-new-vpc"
	req := &setRequest{Name: name}
	jsonBlob := `
{
  "vpc":
` + vTestJSON + `
}
`

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		c := new(setRequest)
		err := json.NewDecoder(r.Body).Decode(c)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPatch)
		require.Equal(t, c, req)
		fmt.Fprint(w, jsonBlob)
	})

	got, _, err := svc.Set(ctx, id, VPCSetName(name))
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestVPCs_Delete(t *testing.T) {
	setup()
	defer teardown()

	svc := client.VPCs
	path := "/v2/vpcs"
	id := "880b7f98-f062-404d-b33c-458d545696f6"

	mux.HandleFunc(path+"/"+id, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := svc.Delete(ctx, id)
	require.NoError(t, err)
}
