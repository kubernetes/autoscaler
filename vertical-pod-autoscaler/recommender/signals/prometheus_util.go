/*
Copyright 2017 The Kubernetes Authors.
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


package signals

import (
	"fmt"
	"io"
	"encoding/json"
	"time"
)

// Helper types used for parsing json returned by Prometheus.
// This is the top-level structure of the response.
type responseType struct {
	// Should be "success".
	Status      string   `json:status`
	Data        dataType `json:data`
	ErrorType   string   `json:errorType`
	ErrorString string   `json:error`
}

// Holds all the data returned.
type dataType struct {
	// For range vectors, this will be "matrix". Other possibilities are:
	// "vector","scalar","string".
	ResultType string          `json:resultType`
	// This has different types depending on ResultType.
	Result     json.RawMessage `json:result`
}

// dataType.Result has this type when ResultType="matrix".
type matrixType struct {
	// Labels of the timeseries.
	Metric map[string]string `json:metric`
	// List of samples. Each sample is represented as a two-item list with
	// floating point timestamp in seconds and a string holding the value
	// of the metric.
	Values [][]interface{}   `json:values`
}

// Decodes the list of samples from the format of matrixType.Values field.
func decodeSamples(input [][]interface{}) ([]Sample, error) {
	res := make([]Sample, 0)
	for _, item := range input {
		if len(item) != 2 {
			return nil, fmt.Errorf("invalid length: %d", len(item))
		}
		ts, ok := item[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid time: %v", item[0])
		}
		stringVal, ok := item[1].(string)
		if !ok {
			return nil, fmt.Errorf("invalid value: %v", item[1])
		}
		var val float64
		fmt.Sscan(stringVal, &val)
		res = append(res, Sample{
			Value:     val,
			Timestamp: time.Unix(int64(ts), 0)})
	}
	return res, nil
}

// Decodes timeseries from a Prometheus response.
func decodeTimeseriesFromResponse(input io.Reader) ([]Timeseries, error) {
	var resp responseType
	err := json.NewDecoder(input).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response: %v", err)
	}
	if resp.Status != "success" || resp.Data.ResultType != "matrix" {
		return nil, fmt.Errorf("invalid response status: %s or type: %s", resp.Status, resp.Data.ResultType)
	}
	var matrices []matrixType
	err = json.Unmarshal(resp.Data.Result, &matrices)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response matrix: %v", err)
	}
	res := make([]Timeseries, 0)
	for _, matrix := range matrices {
		samples, err := decodeSamples(matrix.Values)
		if err != nil {
			return []Timeseries{}, fmt.Errorf("error decoding samples: %v", err)
		}
		res = append(res, Timeseries{Labels: matrix.Metric, Samples: samples})
	}
	return res, nil
}
