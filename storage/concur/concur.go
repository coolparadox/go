// Copyright 2015 Rafael Lorandi <coolparadox@gmail.com>
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
Package concur is a storage of byte sequences for Go.

Storage keys are numeric and can be automatically assigned.

Basics

	myData := byte[]{0,1,2,3,4,5,6,7,8,9}
	myCollection, _ := concur.New("/path/to/my/collection");
	key, _ := myCollection.Save(myData) // store myData in a new key
	...
	myData2, _ := myCollection.Get(key) // retrieve stored value
	...
	myCollection.Erase(key) // remove a key

Issues

Keys are 32 bit unsigned integers.

Apart from other storage implementations that map a single file as the database,
this package takes a simpler and more naive approach where keys are
managed using filesystem subdirectories. Therefore the filesystem chosen for
storage is the real engine that maps keys to values, and their designers are the
ones who must take credit if this package happens to achieve satisfactory
performance.

Bugs

Concurrent access to a collection is not yet thought of, and can be a
fruitful source of weirdness.

Wish List

Document filesystem guidelines for better performance with package concur.

Protect against concurrent access to collections.

*/
package concur

import "path"
import "errors"
import "fmt"
import "os"
import "io"
import "log"
import "io/ioutil"
import "strings"
import "strconv"

// Concur handles a collection of byte sequences stored in filesystem.
type Concur struct {
	dir string
}

// concurLabel is the file checked for existence of a concur database in a
// directory.
const concurLabel string = ".concur"

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
	concurFile := path.Join(dir, concurLabel)
	concurFileExists := true
	finfo, err = os.Stat(concurFile)
	if err != nil {
		if os.IsNotExist(err) {
			concurFileExists = false
		} else {
			return Concur{}, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", concurFile, err))
		}
	}
	if concurFileExists {
		if finfo.IsDir() {
			return Concur{}, errors.New(fmt.Sprintf("concur db mark file '%s' is a directory", concurFile))
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
		_, err = os.Create(concurFile)
		if err != nil {
			return Concur{}, errors.New(fmt.Sprintf("cannot create concur db mark file '%s'", concurFile))
		}
		log.Printf("concur database initialized in '%s'", dir)
	}
	return Concur{
		dir: dir,
	}, nil
}

// Put creates (or updates) a key with a new value.
func (c Concur) Put(key uint32, value []byte) error {
	var err error
	targetPath := path.Join(c.dir, formatPath(key))
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

// Save creates a key with a new value.
// The key is automatically assigned and guaranteed to be new.
//
// Returns the created key.
func (c Concur) Save(value []byte) (uint32, error) {
	return 0, errors.New("Save() not yet implemented")
}

// Get retrieves the value associated with a key.
func (c Concur) Get(key uint32) ([]byte, error) {
	sourcePath := path.Join(c.dir, formatPath(key))
	buf, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read file '%s': %s", sourcePath, err))
	}
	return buf, nil
}

// Erase erases a key.
func (c Concur) Erase(key uint32) error {
	return errors.New("not yet implemented")
}

// Exists verifies if a key exists.
func (c Concur) Exists(key uint32) (bool, error) {
	return false, errors.New("not yet implemented")
}

// Wipe removes a collection from the filesystem.
func Wipe(dir string) error {
	return errors.New("not yet implemented")
}

// formatPath converts a key to a relative filesystem path.
func formatPath(key uint32) string {
	return strings.Join(
		strings.Split(
			fmt.Sprintf(
				"%07s",
				strconv.FormatUint(
					uint64(key),
					36)),
			""),
		string(os.PathSeparator))
}
