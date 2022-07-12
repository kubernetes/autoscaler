package rest

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
)

func TestCleanPath(t *testing.T) {
	uri := &url.URL{
		Path:   "//foo//bar",
		Scheme: "https",
		Host:   "host",
	}
	cleanPath(uri)

	expected := "https://host/foo/bar"
	if a, e := uri.String(), expected; a != e {
		t.Errorf("expect %q URI, got %q", e, a)
	}
}

func TestMarshalPath(t *testing.T) {
	in := struct {
		Bucket *string `location:"uri" locationName:"bucket"`
		Key    *string `location:"uri" locationName:"key"`
	}{
		Bucket: aws.String("mybucket"),
		Key:    aws.String("my/cool+thing space/object世界"),
	}

	expectURL := `/mybucket/my/cool+thing space/object世界`
	expectEscapedURL := `/mybucket/my/cool%2Bthing%20space/object%E4%B8%96%E7%95%8C`

	req := &request.Request{
		HTTPRequest: &http.Request{
			URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/{bucket}/{key+}"},
		},
		Params: &in,
	}

	Build(req)

	if req.Error != nil {
		t.Fatalf("unexpected error, %v", req.Error)
	}

	if a, e := req.HTTPRequest.URL.Path, expectURL; a != e {
		t.Errorf("expect %q URI, got %q", e, a)
	}

	if a, e := req.HTTPRequest.URL.RawPath, expectEscapedURL; a != e {
		t.Errorf("expect %q escaped URI, got %q", e, a)
	}

	if a, e := req.HTTPRequest.URL.EscapedPath(), expectEscapedURL; a != e {
		t.Errorf("expect %q escaped URI, got %q", e, a)
	}

}

func TestListOfEnums(t *testing.T) {
	cases := []struct {
		Input    interface{}
		Expected http.Header
		WantErr  bool
	}{
		{
			Input: func() interface{} {
				type v struct {
					List []*string `type:"list" location:"header" locationName:"x-amz-test-header"`
				}
				return &v{
					List: []*string{aws.String("foo")},
				}
			}(),
			WantErr: true,
		},
		{
			Input: func() interface{} {
				type v struct {
					List []*string `type:"list" location:"uri" locationName:"someList"`
				}
				return &v{
					List: []*string{aws.String("foo")},
				}
			}(),
			WantErr: true,
		},
		{
			Input: func() interface{} {
				type v struct {
					List map[string][]*string `type:"map" location:"headers" locationName:"x-amz-metadata"`
				}
				return &v{
					List: map[string][]*string{
						"foo": {aws.String("foo")},
					},
				}
			}(),
			WantErr: true,
		},
		{
			Input: func() interface{} {
				type v struct {
					List []*string `type:"list" location:"header" locationName:"x-amz-test-header" enum:"FooBar"`
				}
				return &v{
					List: []*string{aws.String("foo")},
				}
			}(),
			Expected: http.Header{
				"X-Amz-Test-Header": {"foo"},
			},
		},
		{
			Input: func() interface{} {
				type v struct {
					List []*string `type:"list" location:"header" locationName:"x-amz-test-header" enum:"FooBar"`
				}
				return &v{
					List: []*string{aws.String("foo"), nil, aws.String(""), nil, aws.String("bar")},
				}
			}(),
			Expected: http.Header{
				"X-Amz-Test-Header": {"foo,bar"},
			},
		},
		{
			Input: func() interface{} {
				type v struct {
					List []*string `type:"list" location:"header" locationName:"x-amz-test-header" enum:"FooBar"`
				}
				return &v{
					List: []*string{aws.String("f,o,o"), nil, aws.String(""), nil, aws.String(`"bar"`)},
				}
			}(),
			Expected: http.Header{
				"X-Amz-Test-Header": {`"f,o,o","\"bar\""`},
			},
		},
	}

	for i, tt := range cases {
		req := &request.Request{
			HTTPRequest: &http.Request{
				URL: func() *url.URL {
					u, err := url.Parse("https://foo.amazonaws.com")
					if err != nil {
						panic(err)
					}
					return u
				}(),
				Header: map[string][]string{},
			},
			Params: tt.Input,
		}

		Build(req)

		if (req.Error != nil) != (tt.WantErr) {
			t.Fatalf("(%d) WantErr(%t) got %v", i, tt.WantErr, req.Error)
		}

		if tt.WantErr {
			continue
		}

		if e, a := tt.Expected, req.HTTPRequest.Header; !reflect.DeepEqual(e, a) {
			t.Errorf("(%d) expect %v, got %v", i, e, a)
		}
	}
}
