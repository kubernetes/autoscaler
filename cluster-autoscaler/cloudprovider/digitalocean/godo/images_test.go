package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestImages_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.List(ctx, nil)
	if err != nil {
		t.Errorf("Images.List returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.List returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListDistribution(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "distribution"
		actual := r.URL.Query().Get("type")
		if actual != expected {
			t.Errorf("'type' query = %v, expected %v", actual, expected)
		}
		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.ListDistribution(ctx, nil)
	if err != nil {
		t.Errorf("Images.ListDistribution returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.ListDistribution returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListApplication(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "application"
		actual := r.URL.Query().Get("type")
		if actual != expected {
			t.Errorf("'type' query = %v, expected %v", actual, expected)
		}
		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.ListApplication(ctx, nil)
	if err != nil {
		t.Errorf("Images.ListApplication returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.ListApplication returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "true"
		actual := r.URL.Query().Get("private")
		if actual != expected {
			t.Errorf("'private' query = %v, expected %v", actual, expected)
		}

		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.ListUser(ctx, nil)
	if err != nil {
		t.Errorf("Images.ListUser returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.ListUser returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListByTag(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "foo"
		actual := r.URL.Query().Get("tag_name")
		if actual != expected {
			t.Errorf("'tag_name' query = %v, expected %v", actual, expected)
		}

		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.ListByTag(ctx, "foo", nil)
	if err != nil {
		t.Errorf("Images.ListByTag returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.ListByTag returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListImagesMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"images": [{"id":1},{"id":2}], "links":{"pages":{"next":"http://example.com/v2/images/?page=2"}}}`)
	})

	_, resp, err := client.Images.List(ctx, &ListOptions{Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	checkCurrentPage(t, resp, 1)
}

func TestImages_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"images": [{"id":1},{"id":2}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/images/?page=3",
				"prev":"http://example.com/v2/images/?page=1",
				"last":"http://example.com/v2/images/?page=3",
				"first":"http://example.com/v2/images/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.Images.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestImages_GetImageByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"image":{"id":12345}}`)
	})

	images, _, err := client.Images.GetByID(ctx, 12345)
	if err != nil {
		t.Errorf("Image.GetByID returned error: %v", err)
	}

	expected := &Image{ID: 12345}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.GetByID returned %+v, expected %+v", images, expected)
	}
}

func TestImages_GetImageBySlug(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images/ubuntu", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"image":{"id":12345}}`)
	})

	images, _, err := client.Images.GetBySlug(ctx, "ubuntu")
	if err != nil {
		t.Errorf("Image.GetBySlug returned error: %v", err)
	}

	expected := &Image{ID: 12345}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.Get returned %+v, expected %+v", images, expected)
	}
}

func TestImages_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &CustomImageCreateRequest{
		Name:         "my-new-image",
		Url:          "http://example.com/distro-amd64.img",
		Region:       "nyc3",
		Distribution: "Ubuntu",
		Description:  "My new custom image",
		Tags:         []string{"foo", "bar"},
	}

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		expected := map[string]interface{}{
			"name":         "my-new-image",
			"url":          "http://example.com/distro-amd64.img",
			"region":       "nyc3",
			"distribution": "Ubuntu",
			"description":  "My new custom image",
			"tags":         []interface{}{"foo", "bar"},
		}

		var v map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		if !reflect.DeepEqual(v, expected) {
			t.Errorf("Request body\n got=%#v\nwant=%#v", v, expected)
		}

		fmt.Fprintf(w, `{"image": {"id": 1,"created_at": "2018-09-20T19:28:00Z","description": "A custom image","distribution": "Ubuntu","error_message": "","regions": [],"type": "custom","tags":["foo","bar"],"status": "NEW"}}`)
	})

	image, _, err := client.Images.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("Images.Create returned error: %v", err)
	}

	if id := image.ID; id != 1 {
		t.Errorf("expected id '%d', received '%d'", 1, id)
	}
}

func TestImages_Update(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &ImageUpdateRequest{
		Name: "name",
	}

	mux.HandleFunc("/v2/images/12345", func(w http.ResponseWriter, r *http.Request) {
		expected := map[string]interface{}{
			"name": "name",
		}

		var v map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		if !reflect.DeepEqual(v, expected) {
			t.Errorf("Request body = %#v, expected %#v", v, expected)
		}

		fmt.Fprintf(w, `{"image":{"id":1}}`)
	})

	image, _, err := client.Images.Update(ctx, 12345, updateRequest)
	if err != nil {
		t.Errorf("Images.Update returned error: %v", err)
	} else {
		if id := image.ID; id != 1 {
			t.Errorf("expected id '%d', received '%d'", 1, id)
		}
	}
}

func TestImages_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Images.Delete(ctx, 12345)
	if err != nil {
		t.Errorf("Image.Delete returned error: %v", err)
	}
}

func TestImage_String(t *testing.T) {
	image := &Image{
		ID:            1,
		Name:          "Image",
		Type:          "snapshot",
		Distribution:  "Ubuntu",
		Slug:          "image",
		Public:        true,
		Regions:       []string{"one", "two"},
		MinDiskSize:   20,
		SizeGigaBytes: 2.36,
		Created:       "2013-11-27T09:24:55Z",
	}

	stringified := image.String()
	expected := `godo.Image{ID:1, Name:"Image", Type:"snapshot", Distribution:"Ubuntu", Slug:"image", Public:true, Regions:["one" "two"], MinDiskSize:20, SizeGigaBytes:2.36, Created:"2013-11-27T09:24:55Z", Description:"", Status:"", ErrorMessage:""}`
	if expected != stringified {
		t.Errorf("Image.String returned %+v, expected %+v", stringified, expected)
	}
}
