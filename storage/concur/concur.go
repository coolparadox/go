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

/*
Package concur is a storage of byte sequences for Go with automatic key
generation.

Basics

Use New to create or open a collection of key/value pairs in the
filesystem. The collection can then be managed by methods of the collection
handler.

	db, _ := concur.New("/path/to/my/collection")
	key, _ := db.Save(byte[]{1,3,5,7,9}) // store data in a new key
	val, _ := db.Load(key) // retrieve value of a key
	db.SaveAs(key, byte[]{0,2,4,6,8}) // update existent key
	db.Erase(key) // remove a key

Issues

Keys are 32 bit unsigned integers. Values are byte sequences of arbitrary
length.

Apart from other storage implementations that map a single file as the
database, this package takes an experimental approach where keys are managed
using filesystem subdirectories. Therefore the filesystem chosen for storage
is the real engine that maps keys to values, and their designers are the ones
who must take credit if this package happens to achieve satisfactory
performance.

Although concur write methods commit changes to filesystem immediately on
successful return, the operating system can make use of memory buffers for
increasing performance of filesystem access. Users may need to manually
flush updates to disk (eg sync, umount) to guarantee that all updates to the
collection are written to disk.

Wipe method can take a long time to return.

Bugs

Concurrent access to a collection is not yet thought of, and can be a
fruitful source of weirdness.

Wish List

Protect against concurrent access to collections.

Document filesystem guidelines for better performance with package concur.

*/
package concur

import "github.com/coolparadox/go/sort/uint32slice"
import "path"
import "errors"
import "fmt"
import "os"
import "io"
import "log"
import "io/ioutil"

//import "github.com/coolparadox/go/sort/runeslice"

// Concur handles a collection of byte sequences stored in a directory of
// the filesystem.
type Concur struct {
	initialized bool
	dir         string
}

// concurMarkLabel is the file checked for existence of a concur database in a
// directory.
const concurMarkLabel string = ".concur"

// New creates a Concur value.
//
// The dir parameter is an absolute path to a directory in the filesystem
// for storing the collection. If it's the first time this directory is used by
// package concur, it must be empty.
func New(dir string) (Concur, error) {
	if !path.IsAbs(dir) {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not absolute", dir))
	}
	dir = path.Clean(dir)
	finfo, err := os.Stat(dir)
	if err != nil {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is unreachable: %s", dir, err))
	}
	if !finfo.IsDir() {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not a directory", dir))
	}
	concurMarkFile := path.Join(dir, concurMarkLabel)
	concurFileExists := true
	finfo, err = os.Stat(concurMarkFile)
	if err != nil {
		if os.IsNotExist(err) {
			concurFileExists = false
		} else {
			return Concur{}, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", concurMarkFile, err))
		}
	}
	if concurFileExists {
		if finfo.IsDir() {
			return Concur{}, errors.New(fmt.Sprintf("concur db mark file '%s' is a directory", concurMarkFile))
		}
	} else {
		file, err := os.Open(dir)
		if err != nil {
			return Concur{}, errors.New(fmt.Sprintf("cannot open '%s': %s", dir, err))
		}
		defer file.Close()
		_, err = file.Readdir(1)
		if err != io.EOF {
			return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not empty and is not a concur db", dir))
		}
		_, err = os.Create(concurMarkFile)
		if err != nil {
			return Concur{}, errors.New(fmt.Sprintf("cannot create concur db mark file '%s'", concurMarkFile))
		}
		log.Printf("concur database initialized in '%s'", dir)
	}
	return Concur{
		initialized: true,
		dir:         dir,
	}, nil
}

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

// SaveAs creates (or updates) a key with a new value.
func (r Concur) SaveAs(key uint32, value []byte) error {
	err := r.concurLabelExists()
	if err != nil {
		return err
	}
	targetPath := formatPath(key, r.dir)
	targetDir := path.Dir(targetPath)
	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot create directory '%s': %s", targetDir, err))
	}
	err = ioutil.WriteFile(targetPath, value, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot write file '%s': %s", targetPath, err))
	}
	return nil
}

// Load retrieves the value associated with a key.
func (r Concur) Load(key uint32) ([]byte, error) {
	err := r.concurLabelExists()
	if err != nil {
		return nil, err
	}
	sourcePath := formatPath(key, r.dir)
	buf, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read file '%s': %s", sourcePath, err))
	}
	return buf, nil
}

// Erase erases a key.
func (r Concur) Erase(key uint32) error {
	err := r.concurLabelExists()
	if err != nil {
		return err
	}
	targetPath := formatPath(key, r.dir)
	err = os.Remove(targetPath)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot remove file '%s': %s", targetPath, err))
	}
	// Erase full marks up to top level.
	br := decomposeKey(key)
	for level := 1; level <= 6; level++ {
		fullMarkPath := fmt.Sprintf("%s%c%s", keyComponentPath(&br, level, r.dir), os.PathSeparator, "_")
		_ = os.RemoveAll(fullMarkPath)
	}
	return nil
}

// Exists verifies if a key exists.
func (r Concur) Exists(key uint32) (bool, error) {
	err := r.concurLabelExists()
	if err != nil {
		return false, err
	}
	targetPath := formatPath(key, r.dir)
	targetPathExists := true
	_, err = os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			targetPathExists = false
		} else {
			return false, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", targetPath, err))
		}
	}
	return targetPathExists, nil
}

// Wipe removes a collection from the filesystem.
//
// On success, all content of the given directory is cleaned.
// The directory itself is not removed.
//
// Existence of a concur collection in the directory is verified prior to cleaning it.
func Wipe(dir string) error {
	file, err := os.Open(dir)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot open '%s': %s", dir, err))
	}
	_, err = file.Readdir(1)
	if err == io.EOF {
		return nil
	}
	file.Close()
	concurMarkFile := path.Join(dir, concurMarkLabel)
	concurWipingLabel := fmt.Sprintf("%s.wiping", concurMarkLabel)
	concurWipingFile := path.Join(dir, concurWipingLabel)
	_, err = os.Stat(concurMarkFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("cannot check for concur mark file: %s", err))
	}
	if err == nil {
		err = os.Rename(concurMarkFile, concurWipingFile)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot mark collection for wiping: %s", err))
		}
	}
	_, err = os.Stat(concurWipingFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("cannot check for wiping mark file: %s", err))
		}
		return errors.New("missing wiping mark file; aborting")
	}
	file, err = os.Open(dir)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot open '%s': %s", dir, err))
	}
	defer file.Close()
	names, err := file.Readdirnames(0)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot read directory '%s': %s", dir, err))
	}
	for _, name := range names {
		if name == concurWipingLabel {
			continue
		}
		removePath := path.Join(dir, name)
		err := os.RemoveAll(removePath)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot remove '%s': %s", removePath, err))
		}
	}
	err = os.Remove(concurWipingFile)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot remove wiping mark file: %s", err))
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

// SmallestKeyNotLessThan takes a key and returns it if it exists.
// If key does not exist, the closest key in ascending order is returned
// instead.
//
// The bool return value tells if a key was found to be answered.
func (r Concur) SmallestKeyNotLessThan(key uint32) (uint32, bool, error) {

	// Check for unitialized receiver.
	err := r.concurLabelExists()
	if err != nil {
		return 0, false, err
	}

	// minimum represents the smallest admissible value to be answered.
	minimum := decomposeKey(key)

	// Look for a key in descending order of level depth.
	for level := 0; level < 7; level++ {

		if level > 0 {

			// Key was not found in deepest level.
			// Update minimum to represent the first admissible value
			// to be searched in this level.
			for i := 0; i < level; i++ {
				minimum[i] = 35
			}
			k, err := composeKey(&minimum)
			if err != nil {
				return 0, false, nil
			}
			if k < KeyMax {
				k++
			} else {

				// Key range limit reached.
				return 0, false, nil

			}
			minimum = decomposeKey(k)

		}

		// Look for the smallest key not less than the minimum in this depth level.
		br, found, err := smallestKeyNotLessThanInLevel(&minimum, level, r.dir)
		if err != nil {
			return 0, false, errors.New(fmt.Sprintf("cannot lookup key %v: %s", key, err))
		}
		if found {

			// Yay!! Found it :-)
			answer, err := composeKey(&br)
			if err != nil {
				// Assume compose failure is due to garbage leading to impossible broken keys.
				return 0, false, nil
			}
			return answer, true, nil

		}
	}

	// Search exausted in all depth levels.
	return 0, false, nil

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

// Save creates a key with a new value.
// The key is automatically assigned and guaranteed to be new.
//
// Returns the assigned key.
func (r Concur) Save(value []byte) (uint32, error) {
	err := r.concurLabelExists()
	if err != nil {
		return 0, err
	}
	var br brokenKey
	ok, err := findFreeKeyFromLevel(&br, 6, r.dir)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("cannot find free key: %s", err))
	}
	if !ok {
		// findFreeKeyFromLevel() is supposed to always find a (broken) key,
		// even impossible ones.
		panic("Save() weirdness: no free broken key and no erros?!")
	}
	key, err := composeKey(&br)
	if err != nil {
		// As free keys are searched in ascending order, assume impossible
		// ones indicate exaustion of key space.
		return 0, errors.New(fmt.Sprintf("no more keys available."))
	}
	err = r.SaveAs(key, value)
	return key, err
}
