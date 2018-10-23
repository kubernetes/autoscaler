/*
Copyright 2018 The Kubernetes Authors.

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

package requests

import "strconv"

// Integer wrap string
type Integer string

// NewInteger returns Integer format
func NewInteger(integer int) Integer {
	return Integer(strconv.Itoa(integer))
}

// HasValue returns true if integer is not null
func (integer Integer) HasValue() bool {
	return integer != ""
}

// GetValue returns int value
func (integer Integer) GetValue() (int, error) {
	return strconv.Atoi(string(integer))
}

// NewInteger64 returns Integer format
func NewInteger64(integer int64) Integer {
	return Integer(strconv.FormatInt(integer, 10))
}

// GetValue64 returns int64 value
func (integer Integer) GetValue64() (int64, error) {
	return strconv.ParseInt(string(integer), 10, 0)
}

// Boolean wrap string
type Boolean string

// NewBoolean returns Boolean format
func NewBoolean(bool bool) Boolean {
	return Boolean(strconv.FormatBool(bool))
}

// HasValue returns true if boolean is not null
func (boolean Boolean) HasValue() bool {
	return boolean != ""
}

// GetValue returns bool format
func (boolean Boolean) GetValue() (bool, error) {
	return strconv.ParseBool(string(boolean))
}

// Float wrap string
type Float string

// NewFloat returns Float format
func NewFloat(f float64) Float {
	return Float(strconv.FormatFloat(f, 'f', 6, 64))
}

// HasValue returns true if float is not null
func (float Float) HasValue() bool {
	return float != ""
}

// GetValue returns float64 format
func (float Float) GetValue() (float64, error) {
	return strconv.ParseFloat(string(float), 64)
}
