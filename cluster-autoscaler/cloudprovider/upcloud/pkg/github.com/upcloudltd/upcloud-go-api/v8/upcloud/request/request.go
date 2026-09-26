package request

import "net/url"

// Request is the interface for request objects
type Request interface {
	// RequestURL returns the relative API URL for the request, excluding the API version.
	RequestURL() string
}

type QueryFilter interface {
	ToQueryParam() string
}

func encodeQueryFilters(f []QueryFilter) string {
	u := url.Values{}
	for _, v := range f {
		p, err := url.ParseQuery(v.ToQueryParam())
		if err == nil {
			for key := range p {
				u.Add(key, p.Get(key))
			}
		}
	}
	return u.Encode()
}
