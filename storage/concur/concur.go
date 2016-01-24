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

import "path"
import "errors"
import "fmt"
import "os"
import "io"
import "log"
import "io/ioutil"

// KeyMax is a convenience naming for the maximum value of a key.
const KeyMax = 0xFFFFFFFF

// Concur handles a collection of byte sequences stored in a directory of
// the filesystem.
type Concur struct {
	initialized bool
	dir         string
}

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

// SaveAs creates (or updates) a key with a new value.
func (r Concur) SaveAs(key uint32, value []byte) error {
	err := r.concurLabelExists()
	if err != nil {
		return err
	}
	targetDir, targetChar, _ := formatPath(key, r.dir)
	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot create directory '%s': %s", targetDir, err))
	}
	targetPath := joinPathChar(targetDir, targetChar)
	lockFile, err := lockDirForWrite(targetDir)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot lock: %s", err))
	}
	defer lockFile.Close()
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
	targetDir, targetChar, _ := formatPath(key, r.dir)
	targetPath := joinPathChar(targetDir, targetChar)
	lockFile, err := lockDirForRead(targetDir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot lock: %s", err))
	}
	defer lockFile.Close()
	buf, err := ioutil.ReadFile(targetPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read file '%s': %s", targetPath, err))
	}
	return buf, nil
}

// Erase erases a key.
func (r Concur) Erase(key uint32) error {
	err := r.concurLabelExists()
	if err != nil {
		return err
	}
	targetDir, targetChar, br := formatPath(key, r.dir)
	targetPath := joinPathChar(targetDir, targetChar)
	lockFile, err := lockDirForWrite(targetDir)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot lock: %s", err))
	}
	defer lockFile.Close()
	err = os.Remove(targetPath)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot remove file '%s': %s", targetPath, err))
	}
	// Erase full marks up to top level.
	for level := 1; level <= 6; level++ {
		fullMarkPath := fmt.Sprintf("%s%c%s", keyComponentPath(br, level, r.dir), os.PathSeparator, "_")
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
	targetDir, targetChar, _ := formatPath(key, r.dir)
	targetPath := joinPathChar(targetDir, targetChar)
	_, err = os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", targetPath, err))
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

// Save creates a key with a new value.
// The key is automatically assigned and guaranteed to be new.
//
// Returns the assigned key.
func (r Concur) Save(value []byte) (uint32, error) {
	var err error
	err = r.concurLabelExists()
	if err != nil {
		return 0, err
	}
	var br brokenKey
	var targetDir string
	var targetPath string
	var key uint32
	for {
		// Find a free key.
		ok, err := findFreeKeyFromLevel(&br, 6, r.dir)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("cannot find free key: %s", err))
		}
		if !ok {
			// findFreeKeyFromLevel() is supposed to always find a (broken) key,
			// even impossible ones.
			panic("Save() weirdness: no free broken key and no erros?!")
		}
		key, err = composeKey(&br)
		if err != nil {
			// As free keys are searched in ascending order, assume impossible
			// ones indicate exaustion of key space.
			return 0, errors.New(fmt.Sprintf("no more keys available."))
		}
		targetDir = keyComponentPath(br, 1, r.dir)
		err = os.MkdirAll(targetDir, 0777)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("cannot create directory '%s': %s", targetDir, err))
		}
		targetChar := formatMap[br[0]]
		targetPath = joinPathChar(targetDir, targetChar)
		lockFile, err := lockDirForWrite(targetDir)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("cannot lock: %s", err))
		}
		// Make sure another concurrent Save() didn't get the same key.
		_, err = os.Stat(targetPath)
		if err != nil {
			if !os.IsNotExist(err) {
				lockFile.Close()
				return 0, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", targetPath, err))
			}
			// Yay, our key is ours :-)
			defer lockFile.Close()
			break
		}
		// Another concurrent Save() stole our key! >:-/
		lockFile.Close()
	}
	// A free key was found.
	err = ioutil.WriteFile(targetPath, value, 0666)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("cannot write file '%s': %s", targetPath, err))
	}
	return key, nil
}
