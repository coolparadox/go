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
import "path"
import "errors"
import "fmt"
import "os"
import "io/ioutil"

// concurMarkLabel is the file checked for existence of a concur database in a
// directory.
const concurMarkLabel string = ".concur"

// concurLabelExists answers if there is a concur label file at the top level
// of the directory pointed by an initialized collection.
func (r Concur) concurLabelExists() error {
	if !r.initialized {
		return errors.New("unitialized concur.Concur")
	}
	concurMarkFile := path.Join(r.dir, concurMarkLabel)
	_, err := os.Stat(concurMarkFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("cannot check for database label file: %s", err))
		}
		return errors.New("missing database label file")
	}
	return nil
}

// KeyMax is the maximum value of a key.
const KeyMax = 0xFFFFFFFF

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

// brokenKey is a representation of a (uint32) key in base 36 components.
type brokenKey [7]uint32

// decomposeKey converts a key to its components.
func decomposeKey(key uint32) brokenKey {
	answer := new(brokenKey)
	updateBrokenKey(answer, 0, key)
	return *answer
}

// updateBrokenKey updates components of a key starting from a depth level
// and moving up.
func updateBrokenKey(br *brokenKey, level int, key uint32) {
	(*br)[level] = key % 36
	if level >= 6 {
		return
	}
	updateBrokenKey(br, level+1, key/36)
}

// composeKey converts key components to a key.
func composeKey(br *brokenKey) (uint32, error) {

	// Detect if components would overflow uint32.
	var of bool
	of = br[6] > 1
	of = of || (br[6] == 1 && br[5] > 35)
	of = of || (br[6] == 1 && br[5] == 35 && br[4] > 1)
	of = of || (br[6] == 1 && br[5] == 35 && br[4] == 1 && br[3] > 4)
	of = of || (br[6] == 1 && br[5] == 35 && br[4] == 1 && br[3] == 4 && br[2] > 1)
	of = of || (br[6] == 1 && br[5] == 35 && br[4] == 1 && br[3] == 4 && br[2] == 1 && br[1] > 35)
	of = of || (br[6] == 1 && br[5] == 35 && br[4] == 1 && br[3] == 4 && br[2] == 1 && br[1] == 35 && br[0] > 3)
	if of {
		return 0, errors.New(fmt.Sprintf("impossible broken key: %v", *br))
	}

	// Calculate key assuming 7 digits in base 36 representation.
	var key uint32
	key = br[6]
	key *= 36
	key += br[5]
	key *= 36
	key += br[4]
	key *= 36
	key += br[3]
	key *= 36
	key += br[2]
	key *= 36
	key += br[1]
	key *= 36
	key += br[0]
	return key, nil
}

// formatPath converts a key to a relative filesystem path.
func formatPath(key uint32, baseDir string) string {
	br := decomposeKey(key)
	return keyComponentPath(&br, 0, baseDir)
}

// smallestKeyNotLessThan takes broken key components, depth level and a base
// directory to compose the path to a subdirectory in the filesystem. Then it
// returns the smallest key that exists under this subdirectory.
func smallestKeyNotLessThanInLevel(br *brokenKey, level int, baseDir string) (brokenKey, bool, error) {

	// Find out where keys will be searched from.
	// This can be in any valid depth level.
	kcDir := keyComponentPath(br, level+1, baseDir)

	// Iterate through key components of this depth level.
	// Assume components are sorted in ascending order.
	kcs, err := listKeyComponentsInDir(kcDir)
	if err != nil {
		return brokenKey{0, 0, 0, 0, 0, 0, 0}, false, errors.New(fmt.Sprintf("cannot list key components in '%s': %s", kcDir, err))
	}
	for _, kc := range kcs {

		// Discard component if it's smaller than the reference.
		if kc < br[level] {
			continue
		}

		if level <= 0 {

			// Found a matching component in the deepest level.
			return brokenKey{kc, br[1], br[2], br[3], br[4], br[5], br[6]}, true, nil

		} else {

			// Found a matching component in not the deepest level.
			// Answer the smallest key under the next depth level from this component.
			brn := *br
			for i := 0; i < level; i++ {
				brn[i] = 0
			}
			brn[level] = kc
			brf, found, err := smallestKeyNotLessThanInLevel(&brn, level-1, baseDir)
			if err != nil {
				return brokenKey{0, 0, 0, 0, 0, 0, 0}, false, err
			}
			if found {
				return brf, true, nil
			}

		}

	}

	// Search exausted and no keys found.
	return brokenKey{0, 0, 0, 0, 0, 0, 0}, false, nil

}

// keyComponentPath mounts the path to a subdirectory for key components
// of a specific depth level.
func keyComponentPath(br *brokenKey, level int, baseDir string) string {
	var r [7]rune
	for i, c := range br {
		r[i] = formatMap[c]
	}
	switch level {

	case 0:
		return fmt.Sprintf(
			"%s%c%c%c%c%c%c%c%c%c%c%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
			os.PathSeparator,
			r[4],
			os.PathSeparator,
			r[3],
			os.PathSeparator,
			r[2],
			os.PathSeparator,
			r[1],
			os.PathSeparator,
			r[0],
		)

	case 1:
		return fmt.Sprintf(
			"%s%c%c%c%c%c%c%c%c%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
			os.PathSeparator,
			r[4],
			os.PathSeparator,
			r[3],
			os.PathSeparator,
			r[2],
			os.PathSeparator,
			r[1],
		)

	case 2:
		return fmt.Sprintf(
			"%s%c%c%c%c%c%c%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
			os.PathSeparator,
			r[4],
			os.PathSeparator,
			r[3],
			os.PathSeparator,
			r[2],
		)

	case 3:
		return fmt.Sprintf(
			"%s%c%c%c%c%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
			os.PathSeparator,
			r[4],
			os.PathSeparator,
			r[3],
		)

	case 4:
		return fmt.Sprintf(
			"%s%c%c%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
			os.PathSeparator,
			r[4],
		)

	case 5:
		return fmt.Sprintf(
			"%s%c%c%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
			os.PathSeparator,
			r[5],
		)

	case 6:
		return fmt.Sprintf(
			"%s%c%c",
			baseDir,
			os.PathSeparator,
			r[6],
		)

	case 7:
		return fmt.Sprintf(
			"%s",
			baseDir,
		)

	}
	panic(fmt.Sprintf("impossible level value %v", level))

}

func findFreeKeyFromLevel(br *brokenKey, level int, baseDir string) (bool, error) {

	var err error
	fullMarkPath := fmt.Sprintf("%s%c%s", keyComponentPath(br, level+1, baseDir), os.PathSeparator, "_")
	_, err = os.Stat(fullMarkPath)
	if err == nil {
		// There is a full mark a this level.
		return false, nil
	}
	if !os.IsNotExist(err) {
		// Cannot verify if full mark exists.
		return false, err
	}
	// Iterate through key components at this level.
	var kc uint32
	for kc = 0; kc < 36; kc++ {
		br[level] = kc
		targetPath := keyComponentPath(br, level, baseDir)
		_, err := os.Stat(targetPath)
		if err == nil {
			if level > 0 {
				// Found an existent key component in a non zero level.
				// Investigate it for a free key.
				ok, err := findFreeKeyFromLevel(br, level-1, baseDir)
				if err != nil {
					return false, err
				}
				if !ok {
					continue
				}
				return true, nil
			} else {
				// Key component already exists in zero level.
				continue
			}
		}
		if !os.IsNotExist(err) {
			return false, err
		}
		// Key component does not exist.
		return true, nil
	}
	// Exausted key components; add a full mark here to save future work.
	err = ioutil.WriteFile(fullMarkPath, nil, 0666)
	if err != nil {
		return false, err
	}
	br[level] = 0
	return false, nil
}
