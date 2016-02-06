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
import "unicode"
import "unicode/utf8"

const tableLenMin = 1

func FormatChar(kc uint32) rune {
	return formatChar(kc)
}

func ParseChar(r rune) (uint32, error) {
	return parseChar(r)
}

type formatRange struct {
	startComponent uint16
	rangeLen       uint16
	startChar      rune
	charStride     uint16
}

var formatMap []formatRange

func init() {

	formatMap = make([]formatRange, 0)
	formatMap = append(
		formatMap,
		formatRange{
			startComponent: 0,
			rangeLen:       10,
			startChar:      '0',
			charStride:     1,
		},
	)
	formatMap = append(
		formatMap,
		formatRange{
			startComponent: 10,
			rangeLen:       26,
			startChar:      'A',
			charStride:     1,
		},
	)

	var numberIDX int
	for i, table := range unicode.Number.R16 {
		if table.Lo <= '9' {
			continue
		}
		numberIDX = i
		break
	}
	number16 := true

	var letterIDX int
	for i, table := range unicode.Letter.R16 {
		if table.Lo <= 'Z' {
			continue
		}
		letterIDX = i
		break
	}
	letter16 := true

	var nextComponent uint32 = 36
	for nextComponent < BaseMax {

		//fmt.Printf("component %v numberIDX %v %v letterIDX %v %v\n", nextComponent, number16, numberIDX, letter16, letterIDX)
		var nextNumber uint32
		if number16 {
			if numberIDX < len(unicode.Number.R16) {
				nextNumber = uint32(unicode.Number.R16[numberIDX].Lo)
			} else {
				number16 = false
				nextNumber = unicode.Number.R32[0].Lo
				numberIDX = 0
			}
		} else {
			if numberIDX < len(unicode.Number.R32) {
				nextNumber = unicode.Number.R32[numberIDX].Lo
			} else {
				nextNumber = unicode.MaxRune
			}
		}

		var nextLetter uint32
		if letter16 {
			if letterIDX < len(unicode.Letter.R16) {
				nextLetter = uint32(unicode.Letter.R16[letterIDX].Lo)
			} else {
				letter16 = false
				nextLetter = unicode.Letter.R32[0].Lo
				letterIDX = 0
			}
		} else {
			nextLetter = unicode.Letter.R32[letterIDX].Lo
		}

		var table unicode.Range32
		if nextNumber < nextLetter {
			if number16 {
				t := unicode.Number.R16[numberIDX]
				table.Lo = uint32(t.Lo)
				table.Hi = uint32(t.Hi)
				table.Stride = uint32(t.Stride)
			} else {
				t := unicode.Number.R32[numberIDX]
				table.Lo = t.Lo
				table.Hi = t.Hi
				table.Stride = t.Stride
			}
			numberIDX++
		} else if nextLetter < nextNumber {
			if letter16 {
				t := unicode.Letter.R16[letterIDX]
				table.Lo = uint32(t.Lo)
				table.Hi = uint32(t.Hi)
				table.Stride = uint32(t.Stride)
			} else {
				t := unicode.Letter.R32[letterIDX]
				table.Lo = t.Lo
				table.Hi = t.Hi
				table.Stride = t.Stride
			}
			letterIDX++
		} else {
			panic("unicode range tables corruption")
		}
		tableLen := (table.Hi-table.Lo)/table.Stride + 1
		formatMap = append(
			formatMap,
			formatRange{
				startComponent: uint16(nextComponent),
				rangeLen:       uint16(tableLen),
				startChar:      rune(table.Lo),
				charStride:     uint16(table.Stride),
			},
		)
		fmt.Printf("formatMap append %v %v '%c' %v\n", nextComponent, tableLen, table.Lo, table.Stride)
		nextComponent += tableLen
	}

}

// formatChar converts a key component to its character representation in the
// filesystem.
func formatChar(kc uint32) rune {
	if kc > 0xFFFF {
		panic("key component out of range")
	}
	var charCount uint32

	for _, table := range unicode.Letter.R16 {
		tableLen := uint32((table.Hi-table.Lo)/table.Stride + 1)
		if tableLen < tableLenMin {
			continue
		}
		//fmt.Printf("R16 table len %v\n", tableLen)
		if tableLen > kc || charCount > kc-tableLen {
			return rune(uint32(table.Lo) + (kc-charCount)*uint32(table.Stride))
		}
		charCount += uint32(tableLen)
	}

	for _, table := range unicode.Letter.R32 {
		tableLen := uint32((table.Hi-table.Lo)/table.Stride + 1)
		if tableLen < tableLenMin {
			continue
		}
		//fmt.Printf("R32 table len %v\n", tableLen)
		if tableLen > kc || charCount > kc-tableLen {
			return rune(uint32(table.Lo) + (kc-charCount)*uint32(table.Stride))
		}
		charCount += uint32(tableLen)
	}

	panic("character exaustion")
}

// parseChar converts a character to its key component value.
func parseChar(r rune) (uint32, error) {

	c := uint32(r)
	var charCount uint32

	for _, table := range unicode.Letter.R16 {
		tableLen := uint32((table.Hi-table.Lo)/table.Stride + 1)
		if tableLen < tableLenMin {
			continue
		}
		if uint32(table.Hi) >= c {
			return charCount + (c-uint32(table.Lo))/uint32(table.Stride), nil
		}
		charCount += uint32(tableLen)
	}

	for _, table := range unicode.Letter.R32 {
		tableLen := uint32((table.Hi-table.Lo)/table.Stride + 1)
		if tableLen < tableLenMin {
			continue
		}
		if table.Hi >= c {
			return charCount + (c-table.Lo)/table.Stride, nil
		}
		charCount += uint32(tableLen)
	}

	return 0, errors.New("unknown character")

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
		return nil, errors.New(fmt.Sprintf("cannot open directory '%s': %s", dir, err))
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read directory '%s': %s", dir, err))
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
