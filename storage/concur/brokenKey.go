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

import "fmt"
import "os"
import "io/ioutil"

// brokenKey is a representation of a key in base components.
type brokenKey []uint32

// newBrokenKey is a convenience wrapper for creating a new initialized
// brokenKey.
func newBrokenKey(keyDepth int) brokenKey {
	return make(brokenKey, keyDepth)
}

// decomposeKey converts a key to its components.
func decomposeKey(key uint32, keyBase uint32, keyDepth int) brokenKey {
	answer := newBrokenKey(keyDepth)
	for i := 0; i < keyDepth; i++ {
		answer[i] = key % keyBase
		key /= keyBase
	}
	return answer
}

// multiplyUint32 multiplies two uint32s with overflow detection.
func multiplyUint32(a, b uint32) (uint32, error) {
	c := a * b
	if a <= 1 || b <= 1 || c/b == a {
		return c, nil
	}
	return 0, fmt.Errorf("overflow")
}

// addUint32 adds two uint32s with overflow detection.
func addUint32(a, b uint32) (uint32, error) {
	c := a + b
	if c >= a && c >= b {
		return c, nil
	}
	return 0, fmt.Errorf("overflow")
}

// composeKey converts key components to a key.
func composeKey(br brokenKey, keyBase uint32, keyDepth int) (uint32, error) {
	var err error
	answer := br[keyDepth-1]
	var i int
	for i = keyDepth - 2; i >= 0; i-- {
		answer, err = multiplyUint32(answer, keyBase)
		if err != nil {
			return 0, fmt.Errorf("impossible broken key '%v'", br)
		}
		answer, err = addUint32(answer, br[i])
		if err != nil {
			return 0, fmt.Errorf("impossible broken key '%v'", br)
		}
	}
	if answer > MaxKey {
		return 0, fmt.Errorf("impossible broken key '%v'", br)
	}
	return answer, nil
}

// keyComponentPath mounts the path to a subdirectory for key components
// of a specific depth level.
func keyComponentPath(br brokenKey, level int, baseDir string, keyDepth int) string {
	rs := make([]rune, 0, 2*keyDepth)
	for i := keyDepth - 1; i >= level; i-- {
		rs = append(rs, os.PathSeparator, formatChar(br[i]))
	}
	return fmt.Sprintf("%s%s", baseDir, string(rs))
}

// formatPath converts a key to a path in the filesystem.
// Returns:
// - a directory in filesystem for holding the value file
// - a character for naming the value file under the directory
// - key components, if the caller is interested
func formatPath(key uint32, baseDir string, keyBase uint32, keyDepth int) (string, rune, brokenKey) {
	br := decomposeKey(key, keyBase, keyDepth)
	dir := keyComponentPath(br, 1, baseDir, keyDepth)
	return dir, formatChar(br[0]), br
}

// joinPathChar adds a character to a filesystem path, after appending the
// path separator.
func joinPathChar(s string, c rune) string {
	return fmt.Sprintf("%s%c%c", s, os.PathSeparator, c)
}

// findKeyInLevel takes broken key components, depth level and a base
// directory to compose the path to a subdirectory in the filesystem. Then it
// returns the smallest (largest) key that exists under this subdirectory
// that is greater (less) than or equal to the given broken key.
func findKeyInLevel(br brokenKey, level int, baseDir string, keyBase uint32, keyDepth int, ascending bool) (brokenKey, error) {
	// Find out where keys will be searched from.
	// This can be in any valid depth level.
	kcDir := keyComponentPath(br, level+1, baseDir, keyDepth)
	// Iterate through key components of this depth level.
	// Assume components are sorted in ascending (descending) order.
	kcs, err := listKeyComponentsInDir(kcDir, keyBase, ascending, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot list key components in '%s': %s", kcDir, err)
	}
	for _, kc := range kcs {
		// Discard component if it's smaller (larger) than the reference.
		if (ascending && kc < br[level]) || (!ascending && kc > br[level]) {
			continue
		}
		if level <= 0 {
			// Found a matching component in the deepest level.
			answer := newBrokenKey(keyDepth)
			copy(answer, br)
			answer[0] = kc
			return answer, nil
		}
		// Found a matching component in not the deepest level.
		// Answer the smallest (largest) key under the next depth level
		// from this component.
		brn := newBrokenKey(keyDepth)
		if !ascending {
			for i := 0; i < level; i++ {
				brn[i] = keyBase - 1
			}
		}
		brn[level] = kc
		for i := level + 1; i < keyDepth; i++ {
			brn[i] = br[i]
		}
		return findKeyInLevel(brn, level-1, baseDir, keyBase, keyDepth, ascending)
	}
	// Search exausted and no keys found.
	return nil, nil
}

func findFreeKeyFromLevel(from brokenKey, level int, baseDir string, keyBase uint32, keyDepth int) (brokenKey, error) {
	var err error
	br := newBrokenKey(keyDepth)
	copy(br, from)
	fullMarkPath := fmt.Sprintf("%s%c%s", keyComponentPath(br, level+1, baseDir, keyDepth), os.PathSeparator, fullMarkLabel)
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
		targetPath := keyComponentPath(br, level, baseDir, keyDepth)
		_, err := os.Stat(targetPath)
		if err == nil {
			if level > 0 {
				// Found an existent key component in a superior level.
				// Deep investigate level for a free key.
				answer, err := findFreeKeyFromLevel(br, level-1, baseDir, keyBase, keyDepth)
				if err != nil {
					return nil, err
				}
				if answer == nil {
					continue
				}
				return answer, nil
			}
			// Key component already exists in deepest level.
			continue
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
