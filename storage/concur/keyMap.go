// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Concur, a generic value storage library
// for the Go language.
//
// Concur is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Concur is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Concur. If not, see <http://www.gnu.org/licenses/>.

package concur

import "github.com/coolparadox/go/sort/uint32slice"
import "errors"
import "fmt"
import "os"

// formatSequence contains characters to be used for mapping between
// filesystem names and components of keys.
const formatSequence = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Mapping between characters and positions in formatSequence.
var (
	keyBase   uint32
	keyDepth  int
	formatMap map[uint32]rune
	parseMap  map[rune]uint32
)

func init() {
	for i, c1 := range formatSequence {
		for j, c2 := range formatSequence {
			if j <= i {
				continue
			}
			if c1 == c2 {
				panic(fmt.Sprintf("non unique character in format sequence: '%c'", c1))
			}
		}
	}
	keyBase = uint32(len(formatSequence))
	var k uint32
	for k = KeyMax; k > 0; k /= keyBase {
		keyDepth++
	}
	formatMap = make(map[uint32]rune, keyBase)
	parseMap = make(map[rune]uint32, keyBase)
	for k = 0; k < keyBase; k++ {
		key := rune(formatSequence[k])
		formatMap[uint32(k)] = key
		parseMap[key] = uint32(k)
	}
}

// listKeyComponentsInDir returns all key components found in a subdirectory,
// sorted in ascending order.
func listKeyComponentsInDir(dir string) ([]uint32, error) {
	answer := make([]uint32, 0, keyBase)
	// Iterate through all names in directory.
	var err error
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return answer, nil
		}
		return nil, errors.New(fmt.Sprintf("cannot open directory '%s': %s", dir, err))
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read directory '%s': %s", dir, err))
	}
	for _, name := range names {
		// If name is a key character, store its component value for answer.
		if len(name) > 1 {
			continue
		}
		char := rune(name[0])
		component, ok := parseMap[char]
		if !ok {
			continue
		}
		answer = append(answer, component)
	}
	// Sort answer slice before returning it.
	uint32slice.SortUint32s(answer)
	return answer, nil
}
