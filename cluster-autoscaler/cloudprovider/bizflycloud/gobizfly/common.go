// This file is part of gobizfly
//
// Copyright (C) 2020  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

package gobizfly

// SliceContains - Check data in slice
func SliceContains(slice interface{}, val interface{}) (int, bool) {
	switch v := slice.(type) {
	case string:
		if slice == val {
			return 1, true
		}
		return -1, false
	default:
		for i, item := range v.([]string) {
			if item == val {
				return i, true
			}
		}
		return -1, false
	}
}
