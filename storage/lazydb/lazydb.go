// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of LazyDB, a generic value storage library
// for the Go language.
//
// LazyDB is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LazyDB is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LazyDB. If not, see <http://www.gnu.org/licenses/>.

/*
Package lazydb is a storage of byte sequences for Go with automatic key
generation.

Basics

Use New to create or open a collection of key/value pairs in the
filesystem. The collection can then be managed by methods of the collection
handler.

	db, _ := lazydb.New("/path/to/my/collection", 0)
	key, _ := db.Save(bytes.NewReader(byte[]{1,3,5,7,9})) // store data in a new key
	val := new(bytes.Buffer)
	db.Load(key, val) // retrieve value of a key
	db.SaveAs(key, bytes.NewReader(byte[]{0,2,4,6,8})) // update existent key
	db.Erase(key) // remove a key

Issues

Keys are 32 bit unsigned integers. Values are byte sequences of arbitrary
length.

Although lazydb write methods commit changes to filesystem immediately on
successful return,
commited data may reside temporarily in on-memory filesystem's caches.
Users may need to manually
flush updates to disk (eg sync, umount) to guarantee that all updates to the
collection are written to disk.

Wipe method can take a long time to return.

Key Mapping Internals

Apart from storage implementations that map a single file as the database,
LazyDB relies on subdirectories of the filesystem for managing keys
(hence the "lazy" part of its name).
Therefore the filesystem chosen for storage
is the real engine that maps keys to values, and their designers are the ones
who must be given credit if package LazyDB happens to achieve satisfactory
performance.

Each key is uniquely associated with a distinct file in the filesystem.
The path to the file is derived from the key, eg. a key of 0x12345678,
assuming the numeric base of key components is set to 16, is file
1/2/3/4/5/6/7/8 under the database directory. The value associated with the
key is the content of the file. Conversely, keys in the database are retrieved
by parsing the path of existent files.

When creating a new database, user may choose the numeric base of key
components. This value ultimately defines how many directories are allowed to
exist in each subdirectory level towards reaching associated files.
The base can range from MinBase (2, resulting in a level depth of 32)
to MaxBase (0x10000, for a level depth of 2).

Whether the numeric base chosen, directories and files are named by single
unicode characters, where the first 10 ones in the mapping range are decimal
digits from 0 to 9, and the next 26 ones are upper case letters from A to Z.
Thus component bases up to 36 are guaranteed to be mapped by characters in the
ascii range.

It's worth noting that all this key composition stuff happens transparently
to the user. Poking around the directory of a lazydb collection, despite it's
cool for the sake of curiosity, is not required for making use of this package.

Wish List

Document filesystem guidelines for better performance with package lazydb.

*/
package lazydb

import "path"
import "fmt"
import "os"
import "io"

// MaxKey represents the maximum value of a key.
const MaxKey = 0xFFFFFFFF

// MinBase and MaxBase define the range of possible values for the numeric base
// of key components in the filesystem (see parameter base in New).
const (
	MinBase = 2
	MaxBase = 0x10000
)

// Depth*Base are convenience values of numeric bases of key components to be
// used when creating a new database.
// These values give the most efficient occupation of subdirectories in the
// filesystem (see Key Mapping Internals).
const (
	Depth2Base  = 0x10000
	Depth4Base  = 0x100
	Depth8Base  = 0x10
	Depth16Base = 0x4
	Depth32Base = 0x2
)

// LazyDB handles a collection of byte sequences stored in a directory of
// the filesystem.
type LazyDB struct {
	initialized bool
	dir         string
	keyBase     uint32
	keyDepth    int
}

// dbMarkLabel is the file checked for existence of a lazydb database in a
// directory.
const dbMarkLabel string = ".lazydb"

// fullMarkLabel is the file that marks if a subdirectory is completely full
// of key components.
const fullMarkLabel string = ".full"

// lazydbLabelExists answers if there is a lazydb label file at the top level
// of the directory pointed by an initialized collection.
func (r LazyDB) lazydbLabelExists() error {
	if !r.initialized {
		return fmt.Errorf("unitialized lazydb.LazyDB")
	}
	lazydbMarkFile := path.Join(r.dir, dbMarkLabel)
	_, err := os.Stat(lazydbMarkFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot check for database label file: %s", err)
		}
		return fmt.Errorf("missing database label file")
	}
	return nil
}

// uint32ToBytes converts a uint32 to byte representation.
func uint32ToBytes(x uint32) []byte {
	answer := make([]byte, 4)
	for i := 0; i < 4; i++ {
		answer[i] = byte(x % 0x100)
		x /= 0x100
	}
	return answer
}

// bytesToUint32 is the counterpart of uint32ToBytes.
func bytesToUint32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("invalid length %v of byte sequence", len(b))
	}
	answer := uint32(b[3])
	for i := 1; i < 4; i++ {
		answer *= 0x100
		answer += uint32(b[3-i])
	}
	return answer, nil
}

// New creates a LazyDB value.
//
// Parameter dir is an absolute path to a directory in the filesystem
// for storing the collection. If it's the first time this directory is used by
// package lazydb, it must be empty.
//
// Parameter base is the numeric base of key components for naming files and
// subdirectories under the collection (see Key Mapping Internals for details).
// It has effect only during creation of a new collection
// (i.e., it's ignored when opening an existent collection).
// Pass zero for a sane default.
func New(dir string, base uint32) (LazyDB, error) {
	if !path.IsAbs(dir) {
		return LazyDB{}, fmt.Errorf("dir '%s' is not absolute", dir)
	}
	dir = path.Clean(dir)
	finfo, err := os.Stat(dir)
	if err != nil {
		return LazyDB{}, fmt.Errorf("dir '%s' is unreachable: %s", dir, err)
	}
	if !finfo.IsDir() {
		return LazyDB{}, fmt.Errorf("dir '%s' is not a directory", dir)
	}
	lazydbMarkFile := path.Join(dir, dbMarkLabel)
	lazydbFileExists := true
	finfo, err = os.Stat(lazydbMarkFile)
	if err != nil {
		if os.IsNotExist(err) {
			lazydbFileExists = false
		} else {
			return LazyDB{}, fmt.Errorf("cannot check for '%s' existence: %s", lazydbMarkFile, err)
		}
	}
	if lazydbFileExists {
		if finfo.IsDir() {
			return LazyDB{}, fmt.Errorf("lazydb db mark file '%s' is a directory", lazydbMarkFile)
		}
		cFile, err := os.Open(lazydbMarkFile)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot open lazydb mark file: %s", err)
		}
		b := make([]byte, 4)
		n, err := cFile.Read(b)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot read lazydb mark file: %s", err)
		}
		if n != 4 {
			return LazyDB{}, fmt.Errorf("weird byte length %v from lazydb mark file", n)
		}
		base, err = bytesToUint32(b)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot parse base from lazydb mark file: %s", err)
		}
		if base < MinBase || base > MaxBase {
			panic(fmt.Sprintf("key base value from lazydb mark file is out of range: %v", base))
		}
	} else {
		if base == 0 {
			base = Depth8Base
		}
		if base < MinBase || base > MaxBase {
			return LazyDB{}, fmt.Errorf("base parameter is out of range")
		}
		dFile, err := os.Open(dir)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot open '%s': %s", dir, err)
		}
		defer dFile.Close()
		_, err = dFile.Readdir(1)
		if err != io.EOF {
			return LazyDB{}, fmt.Errorf("dir '%s' is not empty and is not a lazydb db", dir)
		}
		cFile, err := os.Create(lazydbMarkFile)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot create lazydb db mark file '%s'", lazydbMarkFile)
		}
		defer cFile.Close()
		_, err = cFile.Write(uint32ToBytes(base))
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot write base to lazydb mark file: %s", err)
		}
	}
	var k uint32
	var depth int
	for k = MaxKey; k > 0; k /= base {
		depth++
	}
	return LazyDB{
		initialized: true,
		dir:         dir,
		keyBase:     base,
		keyDepth:    depth,
	}, nil
}

// SaveAs creates (or updates) a given key with a new value.
func (r LazyDB) SaveAs(key uint32, value io.Reader) error {
	err := r.lazydbLabelExists()
	if err != nil {
		return err
	}
	targetDir, targetChar, _ := formatPath(key, r.dir, r.keyBase, r.keyDepth)
	targetPath := joinPathChar(targetDir, targetChar)
	lockFile, err := lockDirForWrite(targetDir, true)
	if err != nil {
		return fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file '%s': %s", targetPath, err)
	}
	defer file.Close()
	_, err = io.Copy(file, value)
	if err != nil {
		return fmt.Errorf("cannot transfer contents of file '%s': %s", targetPath, err)
	}
	return nil
}

// Load retrieves the value associated with a key.
func (r LazyDB) Load(key uint32, value io.Writer) error {
	err := r.lazydbLabelExists()
	if err != nil {
		return err
	}
	targetDir, targetChar, _ := formatPath(key, r.dir, r.keyBase, r.keyDepth)
	targetPath := joinPathChar(targetDir, targetChar)
	lockFile, err := lockDirForRead(targetDir)
	if err != nil {
		return fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	file, err := os.Open(targetPath)
	if err != nil {
		return fmt.Errorf("cannot open file '%s': %s", targetPath, err)
	}
	defer file.Close()
	_, err = io.Copy(value, file)
	if err != nil {
		return fmt.Errorf("cannot transfer contents of file '%s': %s", targetPath, err)
	}
	return nil
}

// Erase erases a key.
func (r LazyDB) Erase(key uint32) error {
	err := r.lazydbLabelExists()
	if err != nil {
		return err
	}
	targetDir, targetChar, br := formatPath(key, r.dir, r.keyBase, r.keyDepth)
	targetPath := joinPathChar(targetDir, targetChar)
	_, err = os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot stat: %s", err)
	}
	lockFile, err := lockDirForWrite(targetDir, false)
	if err != nil {
		return fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	err = os.Remove(targetPath)
	if err != nil {
		return fmt.Errorf("cannot remove file '%s': %s", targetPath, err)
	}
	for level := 1; level < r.keyDepth; level++ {
		targetDir := keyComponentPath(br, level, r.dir, r.keyDepth)
		// Erase full marks up to top level.
		fullMarkPath := fmt.Sprintf("%s%c%s", targetDir, os.PathSeparator, fullMarkLabel)
		os.RemoveAll(fullMarkPath)
		// Erase empty directories up to top level.
		_, err := findKeyComponentInDir(targetDir, r.keyBase, 0, findModeAny)
		if err == KeyNotFoundError {
			os.RemoveAll(targetDir)
		}
	}
	return nil
}

// Exists verifies if a key exists.
func (r LazyDB) Exists(key uint32) (bool, error) {
	err := r.lazydbLabelExists()
	if err != nil {
		return false, err
	}
	targetDir, targetChar, _ := formatPath(key, r.dir, r.keyBase, r.keyDepth)
	targetPath := joinPathChar(targetDir, targetChar)
	_, err = os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("cannot check for '%s' existence: %s", targetPath, err)
}

// Wipe removes a collection from the filesystem.
//
// On success, all content of the given directory is cleaned.
// The directory itself is not removed.
//
// Existence of a lazydb collection in the directory is verified prior to cleaning it.
func Wipe(dir string) error {
	file, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %s", dir, err)
	}
	_, err = file.Readdir(1)
	if err == io.EOF {
		return nil
	}
	file.Close()
	lazydbMarkFile := path.Join(dir, dbMarkLabel)
	lazydbWipingLabel := fmt.Sprintf("%s.wiping", dbMarkLabel)
	lazydbWipingFile := path.Join(dir, lazydbWipingLabel)
	_, err = os.Stat(lazydbMarkFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot check for lazydb mark file: %s", err)
	}
	if err == nil {
		err = os.Rename(lazydbMarkFile, lazydbWipingFile)
		if err != nil {
			return fmt.Errorf("cannot mark collection for wiping: %s", err)
		}
	}
	_, err = os.Stat(lazydbWipingFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot check for wiping mark file: %s", err)
		}
		return fmt.Errorf("missing wiping mark file; aborting")
	}
	file, err = os.Open(dir)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %s", dir, err)
	}
	defer file.Close()
	names, err := file.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("cannot read directory '%s': %s", dir, err)
	}
	for _, name := range names {
		if name == lazydbWipingLabel {
			continue
		}
		removePath := path.Join(dir, name)
		err := os.RemoveAll(removePath)
		if err != nil {
			return fmt.Errorf("cannot remove '%s': %s", removePath, err)
		}
	}
	err = os.Remove(lazydbWipingFile)
	if err != nil {
		return fmt.Errorf("cannot remove wiping mark file: %s", err)
	}
	return nil
}

// FindKey takes a key and returns it if it exists.
// If key does not exist, the closest key in ascending (or descending) order
// is returned instead.
//
// KeyNotFoundError is returned if there are no keys to be answered.
func (r LazyDB) FindKey(key uint32, ascending bool) (uint32, error) {
	err := r.lazydbLabelExists()
	if err != nil {
		return 0, err
	}
	// threshold represents the smallest (largest) admissible value to be
	// answered.
	threshold := decomposeKey(key, r.keyBase, r.keyDepth)
	// Look for a key in deepest level first and then above.
	for level := 0; level < r.keyDepth; level++ {
		if level > 0 {
			// Key was not found in deepest level.
			// Update threshold to represent the first admissible value
			// to be searched in this level.
			for i := 0; i < level; i++ {
				if ascending {
					threshold[i] = r.keyBase - 1
				} else {
					threshold[i] = 0
				}
			}
			k, err := composeKey(threshold, r.keyBase, r.keyDepth)
			if err != nil {
				return 0, KeyNotFoundError
			}
			if ascending && k < MaxKey {
				k++
			} else if !ascending && k > 0 {
				k--
			} else {
				// Key range limit reached.
				return 0, KeyNotFoundError
			}
			threshold = decomposeKey(k, r.keyBase, r.keyDepth)
		}
		// Look for the smallest (largest) key not less (greater) than the
		// threshold in this depth level.
		br, err := findKeyInLevel(threshold, level, r.dir, r.keyBase, r.keyDepth, ascending)
		if err != nil {
			return 0, fmt.Errorf("cannot lookup key %v: %s", key, err)
		}
		if br != nil {
			// Yay!! Found it :-)
			answer, err := composeKey(br, r.keyBase, r.keyDepth)
			if err != nil {
				// Assume compose failure is due to garbage leading to impossible broken keys.
				return 0, KeyNotFoundError
			}
			return answer, nil
		}
	}
	// Search exausted in all depth levels.
	return 0, KeyNotFoundError
}

// Save creates a key with a new value.
// The key is automatically assigned and guaranteed to be new.
//
// Returns the assigned key.
func (r LazyDB) Save(value io.Reader) (uint32, error) {
	var err error
	err = r.lazydbLabelExists()
	if err != nil {
		return 0, err
	}
	var targetDir string
	var targetPath string
	var key uint32
	// Find a free key.
	for {
		br, err := findFreeKeyFromLevel(newBrokenKey(r.keyDepth), r.keyDepth-1, r.dir, r.keyBase, r.keyDepth)
		if err != nil {
			return 0, fmt.Errorf("cannot find free key: %s", err)
		}
		if br == nil {
			// findFreeKeyFromLevel() is supposed to always find a key,
			// even impossible ones.
			panic("Save() weirdness: no free broken key and no errors?!")
		}
		key, err = composeKey(br, r.keyBase, r.keyDepth)
		if err != nil {
			// As free keys are searched in ascending order, assume impossible
			// ones indicate exaustion of key space.
			return 0, fmt.Errorf("no more keys available")
		}
		targetDir = keyComponentPath(br, 1, r.dir, r.keyDepth)
		targetChar := formatChar(br[0])
		targetPath = joinPathChar(targetDir, targetChar)
		lockFile, err := lockDirForWrite(targetDir, true)
		if err != nil {
			return 0, fmt.Errorf("cannot lock: %s", err)
		}
		// Make sure another concurrent Save() didn't get the same key.
		_, err = os.Stat(targetPath)
		if err != nil {
			if !os.IsNotExist(err) {
				lockFile.Close()
				return 0, fmt.Errorf("cannot check for '%s' existence: %s", targetPath, err)
			}
			// Yay, our key is ours :-)
			defer lockFile.Close()
			break
		}
		// Another concurrent Save() stole our key! >:-/
		lockFile.Close()
	}
	// A free key was found.
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return 0, fmt.Errorf("cannot open file '%s': %s", targetPath, err)
	}
	defer file.Close()
	_, err = io.Copy(file, value)
	if err != nil {
		return 0, fmt.Errorf("cannot transfer contents of file '%s': %s", targetPath, err)
	}
	return key, nil
}
