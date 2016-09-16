// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of uint32slice, a generic value storage library
// for the Go language.
//
// uint32slice is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// uint32slice is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with uint32slice. If not, see <http://www.gnu.org/licenses/>.

/*
Package uint32slice is an implementation of sort.Interface for slices of
uint32s.
*/
package uint32slice

import "sort"

// Uint32Slice attaches the methods of sort.Interface to []uint32,
// sorting in increasing order.
type Uint32Slice []uint32

// Len is the number of elements in the collection.
func (r Uint32Slice) Len() int { return len(r) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (r Uint32Slice) Less(i, j int) bool { return r[i] < r[j] }

// Swap swaps the elements with indexes i and j.
func (r Uint32Slice) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// SearchUint32s searches for x in a sorted slice of uint32s and returns the
// index as specified by sort.Search. The return value is the index to insert
// x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint32s(a []uint32, x uint32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Search returns the result of applying SearchUint32s to the receiver and x.
func (r Uint32Slice) Search(x uint32) int { return SearchUint32s(r, x) }

// Sort is a convenience method for applying sort.Sort to the receiver.
func (r Uint32Slice) Sort() { sort.Sort(r) }

// SortUint32s sorts a slice of uint32 in increasing order.
func SortUint32s(s []uint32) { sort.Sort(Uint32Slice(s)) }

// ReversedSortUint32s sorts a slice of uint32 in decreasing order.
func ReversedSortUint32s(s []uint32) { sort.Sort(sort.Reverse(Uint32Slice(s))) }
