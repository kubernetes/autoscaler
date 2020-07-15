package paperspace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"time"
)

var DefaultBaseURL = "https://api.paperspace.io"

type APIBackend struct {
	BaseURL    string
	Debug      bool
	DebugBody  bool
	HTTPClient *http.Client
	RetryCount int
}

func NewAPIBackend() *APIBackend {
	return &APIBackend{
		BaseURL: DefaultBaseURL,
		Debug:   false,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		RetryCount: 0,
	}
}

func (c *APIBackend) Request(ctx context.Context, method string, url string,
	params, result interface{}, headers map[string]string) (res *http.Response, err error) {
	for i := 0; i < c.RetryCount+1; i++ {
		retryDuration := time.Duration((math.Pow(2, float64(i))-1)/2*1000) * time.Millisecond
		time.Sleep(retryDuration)

		res, err = c.request(ctx, method, url, params, result, headers)
		if res != nil && res.StatusCode == 429 {
			continue
		} else {
			break
		}
	}

	return res, err
}

func (c *APIBackend) request(ctx context.Context, method string, url string,
	params, result interface{}, headers map[string]string) (res *http.Response, err error) {
	var data []byte
	body := bytes.NewReader(make([]byte, 0))

	if params != nil {
		data, err = json.Marshal(params)
		if err != nil {
			return res, err
		}

		body = bytes.NewReader(data)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, url)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return res, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Go Paperspace Gradient 1.0")

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if c.Debug {
		requestDump, err := httputil.DumpRequest(req, c.DebugBody)
		if err != nil {
			return res, err
		}
		c.debug(string(requestDump))
	}

	res, err = c.HTTPClient.Do(req)
	if err != nil {
		return res, err
	}
	defer res.Body.Close()

	if c.Debug {
		responseDump, err := httputil.DumpResponse(res, c.DebugBody)
		if err != nil {
			return res, err
		}
		c.debug(string(responseDump))
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()
		errorBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return res, err
		}
		errorReader := bytes.NewReader(errorBody)

		paperspaceErrorResponse := PaperspaceErrorResponse{}
		paperspaceError := PaperspaceError{}
		errorResponseDecoder := json.NewDecoder(errorReader)
		if err := errorResponseDecoder.Decode(&paperspaceErrorResponse); err != nil {
			c.debug(string(err.Error()))
			return res, errors.New("There was a server error, please try your request again")
		}

		if paperspaceErrorResponse.Error == nil {
			errorDecoder := json.NewDecoder(errorReader)
			if err := errorDecoder.Decode(&paperspaceErrorResponse); err != nil {
				c.debug(string(err.Error()))
				return res, errors.New("There was a server error, please try your request again")
			}

			return res, error(paperspaceError)
		}

		return res, error(paperspaceErrorResponse.Error)
	}

	if result != nil {
		decoder := json.NewDecoder(res.Body)
		if err = decoder.Decode(result); err != nil {
			return res, err
		}
	}

	return res, nil
}

func (c *APIBackend) debug(message string) {
	if c.Debug {
		log.Println(message)
	}
}
