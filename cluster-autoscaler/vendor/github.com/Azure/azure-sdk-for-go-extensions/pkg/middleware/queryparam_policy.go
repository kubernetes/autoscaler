package middleware

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type QueryParameterPolicy struct {
	Name    string
	Value   string
	Replace bool
}

func (p *QueryParameterPolicy) Do(req *policy.Request) (*http.Response, error) {
	if !p.Replace {
		// append behavior
		if req.Raw().URL.RawQuery != "" {
			req.Raw().URL.RawQuery += "&"
		}

		req.Raw().URL.RawQuery += url.QueryEscape(p.Name) + "=" + url.QueryEscape(p.Value)
		return req.Next()
	}

	// replace behavior
	originalQueryParams, err := url.ParseQuery(req.Raw().URL.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("cannot replace url query parameter due to parsing err: %w", err)
	}

	originalQueryParams.Set(p.Name, p.Value)
	req.Raw().URL.RawQuery = originalQueryParams.Encode()
	return req.Next()
}
