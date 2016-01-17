// Copyright 2015 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of runeslice, an implementation of sort.Interface
// for slices of runes.
//
// runeslice is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// runeslice is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with runeslice. If not, see <http://www.gnu.org/licenses/>.

/*
Package runeslice is an implementation of sort.Interface for slices of runes.

*/
package runeslice

import "sort"

// RuneSlice attaches the methods of sort.Interface to []rune,
// sorting in increasing order.
type RuneSlice []rune

// Len is the number of elements in the collection.
func (r RuneSlice) Len() int { return len(r) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (r RuneSlice) Less(i, j int) bool { return r[i] < r[j] }

// Swap swaps the elements with indexes i and j.
func (r RuneSlice) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// SearchRunes searches for x in a sorted slice of runes and returns the index
// as specified by sort.Search. The return value is the index to insert x if x
// is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchRunes(a []rune, x rune) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Search returns the result of applying SearchRunes to the receiver and x.
func (r RuneSlice) Search(x rune) int { return SearchRunes(r, x) }

// Sort is a convenience method for applying sort.Sort to the receiver.
func (r RuneSlice) Sort() { sort.Sort(r) }

// SortRunes sorts a slice of runes in increasing order.
func SortRunes (s []rune) { sort.Sort(RuneSlice(s)) }
