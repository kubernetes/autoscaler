/*
Copyright 2024 The Kubernetes Authors.

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

package iputil

// setBitAt sets the bit at the i-th position in the byte slice to the given value.
// Panics if the index is out of bounds.
// For example,
// - setBitAt([0x00, 0x00], 8, 1) returns [0x00, 0b1000_0000].
// - setBitAt([0xff, 0xff], 0, 0) returns [0b0111_1111, 0xff].
func setBitAt(bytes []byte, i int, bit uint8) {
	if bit == 1 {
		bytes[i/8] |= 1 << (7 - i%8)
	} else {
		bytes[i/8] &^= 1 << (7 - i%8)
	}
}

// bitAt returns the bit at the i-th position in the byte slice.
// The return value is either 0 or 1 as uint8.
// Panics if the index is out of bounds.
func bitAt(bytes []byte, i int) uint8 {
	return bytes[i/8] >> (7 - i%8) & 1
}
