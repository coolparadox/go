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
// filesystem names and components of keys in base 36 representation.
// At least 36 characters must be provided in ascending order of representation
// value.
const formatSequence = "0123456789abcdefghijklmnopqrstuvwxyz"

// Mapping between characters and positions in formatSequence.
var (
	formatMap map[uint32]rune
	parseMap  map[rune]uint32
)

func init() {

	// Initialize key component character mapping.
	if len(formatSequence) < 36 {
		panic("missing format characters")
	}
	formatMap = make(map[uint32]rune, 36)
	parseMap = make(map[rune]uint32, 36)
	for i := 0; i < 36; i++ {
		key := rune(formatSequence[i])
		formatMap[uint32(i)] = key
		parseMap[key] = uint32(i)
	}
}

// listKeyComponentsInDir returns all key components found in a subdirectory,
// sorted in ascending order.
func listKeyComponentsInDir(dir string) ([]uint32, error) {

	answer := make([]uint32, 0, 36)

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
