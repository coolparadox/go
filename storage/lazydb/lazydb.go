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

LazyDB is a key/value database library,
where keys are 32 bit unsigned integer numbers
and values are slotted to contain multiple byte sequences of arbitrary length.

Instead of directly implementing indexing algorithms,
LazyDB takes an experimental approach where filesystem directory structure
is used for indexing data saved on disk
(hence the "lazy" part of its name).

Key Mapping Internals

Apart from storage implementations that map a single file as the database,
LazyDB relies on subdirectories of the filesystem for managing keys.
Therefore the filesystem chosen for storage
is the real engine that maps keys to values, and their designers are the ones
who must be given credit if package LazyDB happens to achieve satisfactory
performance.

Each key-slot pair is uniquely associated with a distinct file in the filesystem.
The path to the file is derived from the key-slot,
eg. slot 0 of key 0x12345678,
assuming the numeric base for key mapping is set to 16,
is file 1/2/3/4/5/6/7/8/0 under the database directory.
The data associated with the slot is the content of the file.
Conversely,
keys in the database are retrieved by parsing paths to existent files.

When creating a new database,
user may choose the numeric base for internal key mapping.
The base can range from MinBase to MaxBase,
and it was designed to allow LazyDB to be tuned for the filesystem at use.
The default base is 16,
meaning that a uint32 key requires 8 subdirectories to be mapped.

Whether the numeric base chosen for internal key mapping,
LazyDB uses single unicode characters to name files and subdirectories
whithin the filesystem and manage key mapping.
The first 10 characters in the mapping range are decimal digits from 0 to 9,
and the next 26 ones are upper case letters from A to Z.

Issues

Although LazyDB write methods commit changes to filesystem immediately on
successful return,
commited data may reside temporarily in on memory filesystem's caches.
Users may need to manually
flush updates to disk (eg sync, umount) to guarantee that all updates to the
database are written to disk.

Wipe method can take a long time to return.

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

// MinBase and MaxBase define the range of possible values of the numeric base
// for internal key mapping in the filesystem (see parameter base in New).
const (
	MinBase = 2
	MaxBase = 0x10000
)

// Depth*Base are convenience values of numeric bases for internal key mapping
// to be used when creating a new database.
// These values give the most efficient occupation of subdirectories in the
// filesystem (see Key Mapping Internals).
const (
	Depth2Base  = 0x10000
	Depth4Base  = 0x100
	Depth8Base  = 0x10
	Depth16Base = 0x4
	Depth32Base = 0x2
)

// LazyDB is a handler to a LazyDB database in the filesystem.
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
// of the directory pointed by an initialized database.
func (db LazyDB) lazydbLabelExists() error {
	if !db.initialized {
		return fmt.Errorf("unitialized lazydb.LazyDB")
	}
	lazydbMarkFile := path.Join(db.dir, dbMarkLabel)
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

// New creates a new LazyDB database in the filesystem,
// or opens an existent one.
//
// Parameter dir is an absolute path to a directory in the filesystem
// for storing the database. If it's the first time this directory is used by
// package lazydb, it must be empty.
//
// Parameter keyBase is the numeric base for naming files and
// subdirectories under the database
// (see: Key Mapping Internals, Depth*Base constants).
// It has effect only during creation of a new database
// (it's ignored when opening an existent database).
// Pass zero for a sane default.
func New(dir string, keyBase uint32) (LazyDB, error) {
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
		keyBase, err = bytesToUint32(b)
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot parse base from lazydb mark file: %s", err)
		}
		if keyBase < MinBase || keyBase > MaxBase {
			panic(fmt.Sprintf("key base value from lazydb mark file is out of range: %v", keyBase))
		}
	} else {
		if keyBase == 0 {
			keyBase = Depth8Base
		}
		if keyBase < MinBase || keyBase > MaxBase {
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
		_, err = cFile.Write(uint32ToBytes(keyBase))
		if err != nil {
			return LazyDB{}, fmt.Errorf("cannot write base to lazydb mark file: %s", err)
		}
	}
	var k uint32
	var depth int
	for k = MaxKey; k > 0; k /= keyBase {
		depth++
	}
	return LazyDB{
		initialized: true,
		dir:         dir,
		keyBase:     keyBase,
		keyDepth:    depth,
	}, nil
}

// copyResult carries the result of a data transfer operation.
type copyResult struct {
	slot  int   // slot id
	count int64 // how many bytes were transferred
	err   error // error during transfer
}

// saveSlot saves data to a slot file
// and writes the result of the operation to a channel.
func saveSlot(dir string, slot int, src io.Reader, c chan copyResult) {
	var result copyResult
	result.slot = slot
	targetPath := joinPathChar(dir, formatChar(uint32(slot)))
	var dst *os.File
	dst, result.err = os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if result.err != nil {
		c <- result
		return
	}
	defer dst.Close()
	result.count, result.err = io.Copy(dst, src)
	c <- result
}

// loadSlot loads data from a slot file
// and writes the result of the operation to a channel.
func loadSlot(dir string, slot int, dst io.Writer, c chan copyResult) {
	var result copyResult
	result.slot = slot
	targetPath := joinPathChar(dir, formatChar(uint32(slot)))
	var src *os.File
	src, result.err = os.Open(targetPath)
	if result.err != nil {
		c <- result
		return
	}
	defer src.Close()
	result.count, result.err = io.Copy(dst, src)
	c <- result
}

// SaveAs updates value slots for a given key.
// Key is created if absent.
//
// For all non nil elements of src, data is read until EOF is reached,
// and corresponding slots are updated with read data.
// Previously existent slots corresponding to nil or missing elements of src are left untouched.
//
// Returns the numbers of bytes read from src elements,
// and the first error encountered during operation.
func (db LazyDB) SaveAs(key uint32, src []io.Reader) ([]int64, error) {
	counts := make([]int64, len(src))
	err := db.lazydbLabelExists()
	if err != nil {
		return counts, err
	}
	targetDir, _ := formatPath(key, db.dir, db.keyBase, db.keyDepth)
	lockFile, err := lockDirForWrite(targetDir, true)
	if err != nil {
		return counts, fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	c := make(chan copyResult)
	defer close(c)
	var slotCount int
	for idx, src := range src {
		if src == nil {
			continue
		}
		go saveSlot(targetDir, idx, src, c)
		slotCount++
	}
	for i := 0; i < slotCount; i++ {
		result := <-c
		counts[result.slot] = result.count
		if err != nil {
			continue
		}
		err = result.err
	}
	return counts, err
}

// Load retrieves data from previously saved value slots.
//
// All non nil elements of dst are written with data from
// corresponding slots.
//
// Returns the number of bytes written to dst elements,
// and the first error encountered during operation.
func (db LazyDB) Load(key uint32, dst []io.Writer) ([]int64, error) {
	counts := make([]int64, len(dst))
	err := db.lazydbLabelExists()
	if err != nil {
		return counts, err
	}
	targetDir, _ := formatPath(key, db.dir, db.keyBase, db.keyDepth)
	lockFile, err := lockDirForRead(targetDir)
	if err != nil {
		return counts, fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	c := make(chan copyResult)
	defer close(c)
	var slotCount int
	for idx, dst := range dst {
		if dst == nil {
			continue
		}
		go loadSlot(targetDir, idx, dst, c)
		slotCount++
	}
	for i := 0; i < slotCount; i++ {
		result := <-c
		counts[result.slot] = result.count
		if err != nil {
			continue
		}
		err = result.err
	}
	return counts, err
}

// Erase erases an existent key from the database.
func (db LazyDB) Erase(key uint32) error {
	err := db.lazydbLabelExists()
	if err != nil {
		return err
	}
	targetDir, br := formatPath(key, db.dir, db.keyBase, db.keyDepth)
	lockFile, err := lockDirForWrite(targetDir, false)
	if err != nil {
		return fmt.Errorf("cannot lock: %s", err)
	}
	defer lockFile.Close()
	err = os.RemoveAll(targetDir)
	if err != nil {
		return fmt.Errorf("cannot remove directoty: %s", err)
	}
	for level := 1; level < db.keyDepth; level++ {
		targetDir := keyComponentPath(br, level, db.dir, db.keyDepth)
		// Erase full marks up to top level.
		fullMarkPath := fmt.Sprintf("%s%c%s", targetDir, os.PathSeparator, fullMarkLabel)
		os.RemoveAll(fullMarkPath)
		// Erase empty directories up to top level.
		_, err := findKeyComponentInDir(targetDir, db.keyBase, 0, findModeAny)
		if err == KeyNotFoundError {
			os.RemoveAll(targetDir)
		}
	}
	return nil
}

// Exists verifies if a key-slot pair exists.
func (db LazyDB) Exists(key uint32, slot uint32) (bool, error) {
	err := db.lazydbLabelExists()
	if err != nil {
		return false, err
	}
	targetDir, _ := formatPath(key, db.dir, db.keyBase, db.keyDepth)
	targetPath := joinPathChar(targetDir, formatChar(slot))
	_, err = os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("cannot check for '%s' existence: %s", targetPath, err)
}

// Wipe removes a LazyDB database from the filesystem.
//
// On success, all content of the given directory is cleaned.
// The directory itself is not removed.
//
// Existence of a LazyDB database in the directory is verified
// prior to wiping.
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
			return fmt.Errorf("cannot mark database for wiping: %s", err)
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
func (db LazyDB) FindKey(key uint32, ascending bool) (uint32, error) {
	err := db.lazydbLabelExists()
	if err != nil {
		return 0, err
	}
	// threshold represents the smallest (largest) admissible value to be
	// answered.
	threshold := decomposeKey(key, db.keyBase, db.keyDepth)
	// Look for a key in deepest level first and then above.
	for level := 0; level < db.keyDepth; level++ {
		if level > 0 {
			// Key was not found in deepest level.
			// Update threshold to represent the first admissible value
			// to be searched in this level.
			for i := 0; i < level; i++ {
				if ascending {
					threshold[i] = db.keyBase - 1
				} else {
					threshold[i] = 0
				}
			}
			k, err := composeKey(threshold, db.keyBase, db.keyDepth)
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
			threshold = decomposeKey(k, db.keyBase, db.keyDepth)
		}
		// Look for the smallest (largest) key not less (greater) than the
		// threshold in this depth level.
		br, err := findKeyInLevel(threshold, level, db.dir, db.keyBase, db.keyDepth, ascending)
		if err != nil {
			return 0, fmt.Errorf("cannot lookup key %v: %s", key, err)
		}
		if br != nil {
			// Yay!! Found it :-)
			answer, err := composeKey(br, db.keyBase, db.keyDepth)
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

// Save creates a new key and updates it with given value slots.
// Key is automatically assigned and guaranteed to be new.
//
// For all non nil elements of src, data is read until EOF is reached,
// and corresponding slots are initialized with read data.
//
// Returns the assigned key,
// the number of bytes read from src elements,
// and the first error encountered during operation.
func (db LazyDB) Save(src []io.Reader) (uint32, []int64, error) {
	counts := make([]int64, len(src))
	err := db.lazydbLabelExists()
	if err != nil {
		return 0, counts, err
	}
	var targetDir string
	var key uint32
	// Find a free key.
	for {
		br, err := findFreeKeyFromLevel(newBrokenKey(db.keyDepth), db.keyDepth-1, db.dir, db.keyBase, db.keyDepth)
		if err != nil {
			return 0, counts, fmt.Errorf("cannot find free key: %s", err)
		}
		if br == nil {
			// findFreeKeyFromLevel() is supposed to always find a key,
			// even impossible ones.
			panic("Save() weirdness: no free broken key and no errors?!")
		}
		key, err = composeKey(br, db.keyBase, db.keyDepth)
		if err != nil {
			// As free keys are searched in ascending order, assume impossible
			// ones indicate exaustion of key space.
			return 0, counts, fmt.Errorf("no more keys available")
		}
		targetDir = keyComponentPath(br, 0, db.dir, db.keyDepth)
		lockFile, err := lockDirForWriteNB(targetDir, true)
		if err != nil {
			return 0, counts, fmt.Errorf("cannot lock: %s", err)
		}
		if lockFile != nil {
			// Got a fresh new key. Yay!!
			defer lockFile.Close()
			break
		}
		// Another concurrent Save() stole our key :-/
	}
	// A free key was found.
	c := make(chan copyResult)
	defer close(c)
	var slotCount int
	for idx, src := range src {
		if src == nil {
			continue
		}
		go saveSlot(targetDir, idx, src, c)
		slotCount++
	}
	for i := 0; i < slotCount ; i++ {
		result := <-c
		counts[result.slot] = result.count
		if err != nil {
			continue
		}
		err = result.err
	}
	return key, counts, err
}
