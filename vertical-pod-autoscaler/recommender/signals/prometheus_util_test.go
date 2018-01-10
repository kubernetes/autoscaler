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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSingleTimeseries(t *testing.T) {
	s := `{"status":"success",
	       "data":{
		 "resultType":"matrix",
		 "result":[{
		   "metric":{
	             "__name__":"up",
		     "x":"y"},
		   "values":[[1515422500.45,"0."],
			    [1515422560.453,"1."],
			    [1515422620.45,"0"]]}]}}"`
	res, err := decodeTimeseriesFromResponse(strings.NewReader(s))
	assert.Nil(t, err)
	assert.Equal(t, res, []Timeseries{Timeseries{
		Labels: map[string]string{"__name__": "up", "x": "y"},
		Samples: []Sample{
			Sample{Value: 0, Timestamp: time.Unix(1515422500, 0)},
			Sample{Value: 1, Timestamp: time.Unix(1515422560, 0)},
			Sample{Value: 0, Timestamp: time.Unix(1515422620, 0)},
		}}})
}

func TestEmptyResponse(t *testing.T) {
	s := `{"status":"success", "data":{"resultType":"matrix", "result":[]}}`
	res, err := decodeTimeseriesFromResponse(strings.NewReader(s))
	assert.Nil(t, err)
	assert.Equal(t, res, []Timeseries{})
}

func TestResponseError(t *testing.T) {
	s := `{"status":"error", "error":"my bad", "errorType":"some"}`
	res, err := decodeTimeseriesFromResponse(strings.NewReader(s))
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestParseError(t *testing.T) {
	s := `{"status":"success", "other-key":[unparsable], "data":{"resultType":"matrix", "result":[]}}`
	res, err := decodeTimeseriesFromResponse(strings.NewReader(s))
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestTwoTimeseries(t *testing.T) {
	s := `{"status":"success",
	       "data":{
		 "resultType":"matrix",
		 "result":[
		   {"metric":{"x":"y"},
		    "values":[[1515422620,"15"]]},
		   {"metric":{"x":"z"},
		    "values":[]}]}}"`
	res, err := decodeTimeseriesFromResponse(strings.NewReader(s))
	assert.Nil(t, err)
	assert.Equal(t, res, []Timeseries{
		Timeseries{
			Labels: map[string]string{"x": "y"},
			Samples: []Sample{Sample{Value: 15, Timestamp: time.Unix(1515422620, 0)}}},
		Timeseries{
			Labels: map[string]string{"x":"z"},
			Samples: []Sample{}}})
}
