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
Package concur is a storage of byte sequences for Go with automatic key
generation.

Basics

Use New to create (or open) a collection of key / value pairs in the filesystem.
The collection can then be managed by methods of the collection handler.

	myData := byte[]{0,1,2,3,4,5,6,7,8,9}
	myCollection, _ := concur.New("/path/to/my/collection");
	key, _ := myCollection.Save(myData) // store myData in a new key
	...
	myData2, _ := myCollection.Get(key) // retrieve stored value
	...
	myCollection.Erase(key) // remove a key

Issues

Keys are 32 bit unsigned integers. Values are byte sequences of arbitrary length.

Apart from other storage implementations that map a single file as the database,
this package takes an experimental, more naive (and simpler) approach where keys
are managed using filesystem subdirectories. Therefore the filesystem chosen for
storage is the real engine that maps keys to values, and their designers are the
ones who must take credit if this package happens to achieve satisfactory
performance.

Wipe method can take a long time to return.

Bugs

Concurrent access to a collection is not yet thought of, and can be a
fruitful source of weirdness.

If wipe method fails or is interrupted before
termination, subsequent calls to the same directory will fail due to detection of
non empty, non concur storage directory. If this happens, remove directory contents
manually as a workaround.

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
import "strings"
import "strconv"
import "sort"

// Concur handles a collection of byte sequences stored in a directory of
// the filesystem.
type Concur struct {
	initialized bool
	dir         string
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
		initialized: true,
		dir:         dir,
	}, nil
}

// Put creates (or updates) a key with a new value.
func (c Concur) Put(key uint32, value []byte) error {
	if !c.initialized {
		return errors.New("unitialized concur.Concur")
	}
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
// Returns the assigned key.
func (c Concur) Save(value []byte) (uint32, error) {
	if !c.initialized {
		return 0, errors.New("unitialized concur.Concur")
	}
	return 0, errors.New("Save() not yet implemented")
}

// Get retrieves the value associated with a key.
func (c Concur) Get(key uint32) ([]byte, error) {
	if !c.initialized {
		return nil, errors.New("unitialized concur.Concur")
	}
	sourcePath := path.Join(c.dir, formatPath(key))
	buf, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read file '%s': %s", sourcePath, err))
	}
	return buf, nil
}

// Erase erases a key.
func (c Concur) Erase(key uint32) error {
	if !c.initialized {
		return errors.New("unitialized concur.Concur")
	}
	var err error
	targetPath := path.Join(c.dir, formatPath(key))
	err = os.Remove(targetPath)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot remove file '%s': %s", targetPath, err))
	}
	return nil
}

// Exists verifies if a key exists.
func (c Concur) Exists(key uint32) (bool, error) {
	if !c.initialized {
		return false, errors.New("unitialized concur.Concur")
	}
	var err error
	targetPath := path.Join(c.dir, formatPath(key))
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
	concurFile := path.Join(dir, concurLabel)
	concurFileExists := true
	_, err = os.Stat(concurFile)
	if err != nil {
		if os.IsNotExist(err) {
			concurFileExists = false
		} else {
			return errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", concurFile, err))
		}
	}
	if !concurFileExists {
		return errors.New(fmt.Sprintf("directory '%s' does not contain a concur collection", dir))
	}
	err = os.Remove(concurFile)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot remove '%s': %s", concurFile, err))
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
		removePath := path.Join(dir, name)
		err := os.RemoveAll(removePath)
		if err != nil {
			return errors.New(fmt.Sprintf("cannot remove '%s': %s", removePath, err))
		}
	}
	return nil
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

// NewKeyList creates a channel for retrieval of stored keys.
//
// The keys channel answers keys in ascending order.
// When the last key is answered, the channel is closed.
// Changes in key set ocurring after creation of keys channel are not
// guaranteed to be detected, nor to be not.
//
// Closing the done channel at any time also causes the keys channel to be closed.
func (c Concur) NewKeyList() (keys <-chan uint32, done chan<- interface{}, err error) {
	return nil, nil, errors.New("not yet implemented")
}

const formatSequence = "0123456789abcdefghijklmnopqrstuvwxyz"
var formatMap map[rune]uint32

func init() {

	// Initialize format map
	if len(formatSequence) < 36 {
		panic("missing format characters")
	}
	formatMap = make(map[rune]uint32, 36)
	for i := 0; i < 36 ; i++ {
		key := rune(formatSequence[i])
		formatMap[key] = uint32(i)
	}
}

type runeSlice []rune

func (s runeSlice) Len() int {
	return len(s)
}

func (s runeSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s runeSlice) Swap(i, j int) {
	aux := s[i]
	s[i] = s[j]
	s[j] = aux
}

func ListFormatCharsInDir(dir string) ([]rune, error) {
	var err error
	f, err := os.Open(dir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot open directory '%s': %s", dir, err))
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read directory '%s': %s", dir, err))
	}
	ans := make([]rune, 0, 36)
	for _, name := range names {
		if len(name) > 1 {
			continue
		}
		char := rune(name[0])
		_, ok := formatMap[char]
		if ok {
			ans = append(ans, char)
		}
	}
	sort.Sort(runeSlice(ans))
	return ans, nil
}
