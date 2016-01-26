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

import "errors"
import "fmt"
import "os"
import "io/ioutil"

// brokenKey is a representation of a key in base components.
type brokenKey []uint32

// newBrokenKey is a convenience wrapper for creating a new initialized
// brokenKey.
func newBrokenKey() brokenKey {
	return make(brokenKey, keyDepth)
}

// decomposeKey converts a key to its components.
func decomposeKey(key uint32) brokenKey {
	answer := newBrokenKey()
	for i := 0; i < keyDepth; i++ {
		answer[i] = key % keyBase
		key /= keyBase
	}
	return answer
}

// multiplyUint32 multiplies two uint32s with overflow detection.
func multiplyUint32(a, b uint32) (uint32, error) {
	c := a * b
	if a <= 1 || b <= 1 {
		return c, nil
	}
	if c/b == a {
		return c, nil
	}
	return 0, errors.New("overflow")
}

// addUint32 adds two uint32s with overflow detection.
func addUint32(a, b uint32) (uint32, error) {
	c := a + b
	if c >= a && c >= b {
		return c, nil
	}
	return 0, errors.New("overflow")
}

// composeKey converts key components to a key.
func composeKey(br brokenKey) (uint32, error) {
	var err error
	var answer uint32 = br[keyDepth-1]
	var i int
	for i = keyDepth - 2; i >= 0; i-- {
		answer, err = multiplyUint32(answer, keyBase)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("impossible broken key '%v'", br))
		}
		answer, err = addUint32(answer, br[i])
		if err != nil {
			return 0, errors.New(fmt.Sprintf("impossible broken key '%v'", br))
		}
	}
	if answer > KeyMax {
		return 0, errors.New(fmt.Sprintf("impossible broken key '%v'", br))
	}
	return answer, nil
}

// keyComponentPath mounts the path to a subdirectory for key components
// of a specific depth level.
func keyComponentPath(br brokenKey, level int, baseDir string) string {
	rs := make([]rune, 0, 2*keyDepth)
	for i := keyDepth - 1; i >= level; i-- {
		rs = append(rs, os.PathSeparator, formatMap[br[i]])
	}
	return fmt.Sprintf("%s%s", baseDir, string(rs))
}

// formatPath converts a key to a path in the filesystem.
// Returns:
// - a directory in filesystem for holding the value file
// - a character for naming the value file under the directory
// - key components, if the caller is interested
func formatPath(key uint32, baseDir string) (string, rune, brokenKey) {
	br := decomposeKey(key)
	dir := keyComponentPath(br, 1, baseDir)
	return dir, formatMap[br[0]], br
}

// joinPathChar adds a character to a filesystem path, after appending the
// path separator.
func joinPathChar(s string, c rune) string {
	return fmt.Sprintf("%s%c%c", s, os.PathSeparator, c)
}

// smallestKeyNotLessThan takes broken key components, depth level and a base
// directory to compose the path to a subdirectory in the filesystem. Then it
// returns the smallest key that exists under this subdirectory.
func smallestKeyNotLessThanInLevel(br brokenKey, level int, baseDir string) (brokenKey, error) {
	// Find out where keys will be searched from.
	// This can be in any valid depth level.
	kcDir := keyComponentPath(br, level+1, baseDir)
	// Iterate through key components of this depth level.
	// Assume components are sorted in ascending order.
	kcs, err := listKeyComponentsInDir(kcDir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot list key components in '%s': %s", kcDir, err))
	}
	for _, kc := range kcs {
		// Discard component if it's smaller than the reference.
		if kc < br[level] {
			continue
		}
		if level <= 0 {
			// Found a matching component in the deepest level.
			answer := newBrokenKey()
			copy(answer, br)
			answer[0] = kc
			return answer, nil
		} else {
			// Found a matching component in not the deepest level.
			// Answer the smallest key under the next depth level from this component.
			brn := newBrokenKey()
			brn[level] = kc
			for i := level + 1; i < keyDepth; i++ {
				brn[i] = br[i]
			}
			return smallestKeyNotLessThanInLevel(brn, level-1, baseDir)
		}
	}
	// Search exausted and no keys found.
	return nil, nil
}

func findFreeKeyFromLevel(from brokenKey, level int, baseDir string) (brokenKey, error) {
	var err error
	br := newBrokenKey()
	copy(br, from)
	fullMarkPath := fmt.Sprintf("%s%c%s", keyComponentPath(br, level+1, baseDir), os.PathSeparator, "_")
	_, err = os.Stat(fullMarkPath)
	if err == nil {
		// There is a full mark a this level.
		return nil, nil
	}
	if !os.IsNotExist(err) {
		// Cannot verify if full mark exists.
		return nil, err
	}
	// Iterate through key components at this level.
	var kc uint32
	for kc = 0; kc < keyBase; kc++ {
		br[level] = kc
		targetPath := keyComponentPath(br, level, baseDir)
		_, err := os.Stat(targetPath)
		if err == nil {
			if level > 0 {
				// Found an existent key component in a superior level.
				// Deep investigate level for a free key.
				answer, err := findFreeKeyFromLevel(br, level-1, baseDir)
				if err != nil {
					return nil, err
				}
				if answer == nil {
					continue
				}
				return answer, nil
			} else {
				// Key component already exists in deepest level.
				continue
			}
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		// Key component does not exist.
		return br, nil
	}
	// Exausted key components; add a full mark here to save future work.
	err = ioutil.WriteFile(fullMarkPath, nil, 0666)
	if err != nil {
		return nil, err
	}
	br[level] = 0
	return nil, nil
}
