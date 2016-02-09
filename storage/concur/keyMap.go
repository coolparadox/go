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
import "fmt"
import "os"
import "unicode"
import "unicode/utf8"

const tableLenMin = 1

type formatRange struct {
	component uint16
	character rune
	length    uint16
}

var formatMap []formatRange

func init() {

	// Initialize key component character mapping.
	formatMap = make([]formatRange, 0)
	formatMap = append(
		formatMap,
		formatRange{
			component: 0,
			character: '0',
			length:    10,
		},
	)
	formatMap = append(
		formatMap,
		formatRange{
			component: 10,
			character: 'A',
			length:    26,
		},
	)
	var comp uint32 = 36
	var char rune = 'Z' + 1
	var prevChar rune = unicode.MaxRune
	var fr formatRange
	for ; char < unicode.MaxRune; char++ {
		if !unicode.IsPrint(char) {
			continue
		}
		if comp < MaxBase && char == prevChar+1 {
			fr.length++
		} else {
			if fr.length != 0 {
				formatMap = append(formatMap, fr)
				//fmt.Printf("formatMap append %v '%c' (%U) %v\n", fr.component, fr.character, fr.character, fr.length)
			}
			if comp >= MaxBase {
				break
			}
			fr.component = uint16(comp)
			fr.character = char
			fr.length = 1
		}
		prevChar = char
		comp++
	}
	if comp < MaxBase {
		panic("unicode character exaustion")
	}

}

// formatChar converts a key component to its character representation in the
// filesystem.
func formatChar(kc uint32) rune {
	if kc > 0xFFFF {
		panic("key component out of range")
	}
	for _, fr := range formatMap {
		if kc < uint32(fr.component)+uint32(fr.length) {
			return fr.character + rune(kc-uint32(fr.component))
		}
	}
	panic("format character exaustion")
}

// parseChar converts a character to its key component value.
func parseChar(r rune) (uint32, error) {
	for _, fr := range formatMap {
		if r < fr.character+rune(fr.length) {
			return uint32(fr.component) + uint32(r-fr.character), nil
		}
	}
	return 0, fmt.Errorf("unknown format character")
}

// listKeyComponentsInDir returns all key components found in a subdirectory,
// sorted in ascending order.
func listKeyComponentsInDir(dir string, keyBase uint32) ([]uint32, error) {
	answer := make([]uint32, 0, keyBase)
	// Iterate through all names in directory.
	var err error
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return answer, nil
		}
		return nil, fmt.Errorf("cannot open directory '%s': %s", dir, err)
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory '%s': %s", dir, err)
	}
	for _, name := range names {
		// If name is a key character, store its component value for answer.
		char, n := utf8.DecodeRuneInString(name)
		if char == utf8.RuneError {
			continue
		}
		if n < len(name) {
			continue
		}
		component, err := parseChar(char)
		if err != nil {
			continue
		}
		if component >= keyBase {
			continue
		}
		answer = append(answer, component)
	}
	// Sort answer slice before returning it.
	uint32slice.SortUint32s(answer)
	return answer, nil
}
