/*
Copyright 2023 The Kubernetes Authors.

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

package fnutil

func Map[T any, R any](f func(T) R, xs []T) []R {
	rv := make([]R, len(xs))
	for i, x := range xs {
		rv[i] = f(x)
	}
	return rv
}

func Filter[T any](f func(T) bool, xs []T) []T {
	var rv []T
	for _, x := range xs {
		if f(x) {
			rv = append(rv, x)
		}
	}
	return rv
}

func RemoveIf[T any](f func(T) bool, xs []T) []T {
	var rv []T
	for _, x := range xs {
		if !f(x) {
			rv = append(rv, x)
		}
	}
	return rv
}

func IsAll[T any](f func(T) bool, xs []T) bool {
	for _, x := range xs {
		if !f(x) {
			return false
		}
	}
	return true
}

func PlanHashCode[D comparable](data D) D { return data }

func IndexSet[D comparable](xs []D) *IndexSetWithComparableIndex[D, D] {
	return NewIndexSetWithComparableIndex(PlanHashCode, xs)
}

type IndexSetWithComparableIndex[I comparable, D any] struct {
	hashCode func(data D) I
	data     map[I]D
}

func NewIndexSetWithComparableIndex[I comparable, D any](hashCode func(data D) I, xs []D) *IndexSetWithComparableIndex[I, D] {
	if hashCode == nil {
		panic("hashCode must not be nil")
	}
	rv := make(map[I]D, len(xs))
	for _, x := range xs {
		rv[hashCode(x)] = x
	}
	return &IndexSetWithComparableIndex[I, D]{
		data:     rv,
		hashCode: hashCode,
	}
}
func (xs *IndexSetWithComparableIndex[I, D]) Contains(data D) bool {
	_, ok := xs.data[xs.hashCode(data)]
	return ok
}

func (xs *IndexSetWithComparableIndex[I, D]) Intersection(ys []D) []D {
	var rv []D
	for _, y := range ys {
		if xs.Contains(y) {
			rv = append(rv, y)
		}
	}
	return rv
}

func (xs *IndexSetWithComparableIndex[I, D]) SubtractedBy(ys []D) []D {
	var rv []D
	for _, y := range ys {
		if !xs.Contains(y) {
			rv = append(rv, y)
		}
	}
	return rv
}

func Intersection[D comparable](xs, ys []D) []D {
	return IndexSet(xs).Intersection(ys)
}
