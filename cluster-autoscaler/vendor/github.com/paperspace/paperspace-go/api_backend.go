package paperspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
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
	apiBackend := APIBackend{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		RetryCount: 0,
	}

	baseURL := os.Getenv("PAPERSPACE_BASEURL")
	if baseURL != "" {
		apiBackend.BaseURL = baseURL
	}

	debug := os.Getenv("PAPERSPACE_DEBUG")
	if debug != "" {
		apiBackend.Debug = true
	}
	
	debugBody := os.Getenv("PAPERSPACE_DEBUG_BODY")
	if debugBody != "" {
		apiBackend.DebugBody = true
	}


	return &apiBackend
}

func (c *APIBackend) Request(method string, url string,
	params, result interface{}, requestParams RequestParams) (res *http.Response, err error) {
	for i := 0; i < c.RetryCount+1; i++ {
		retryDuration := time.Duration((math.Pow(2, float64(i))-1)/2*1000) * time.Millisecond
		time.Sleep(retryDuration)

		res, err = c.request(method, url, params, result, requestParams)
		if res != nil && res.StatusCode == 429 {
			continue
		} else {
			break
		}
	}

	return res, err
}

func (c *APIBackend) request(method string, url string,
	params, result interface{}, requestParams RequestParams) (res *http.Response, err error) {
	var data []byte
	var req *http.Request
	body := bytes.NewReader(make([]byte, 0))

	if params != nil {
		data, err = json.Marshal(params)
		if err != nil {
			return res, err
		}

		body = bytes.NewReader(data)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, url)

	req, err = http.NewRequest(method, fullURL, body)
	if err != nil {
		return res, err
	}

	if requestParams.Context != nil {
		req = req.WithContext(requestParams.Context)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Go Paperspace Gradient 1.0")

	for key, value := range requestParams.Headers {
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
