// Copyright 2015 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Keep, a persistency library for the Go language.
//
// Keep is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Keep is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Keep. If not, see <http://www.gnu.org/licenses/>.

/*
Package keep offers filesystem persistence for typed values in Go.
It works with numeric, bool or string based types, or any composition of these
types by arrays, slices, maps and structs.

Basics

Let's say you have a variable of some type containing data you
want to persist. Create a Keep value for this purpose:

	var myData MyType
	myKeep, _ := keep.New(&myData, "/home/user/my_data")

In the above example, myKeep opens a collection of MyType values under
directory /home/user/my_data in the filesystem, using myData as a buffer for
transfer values to/from the collection:

	myData = ...  // populate with data to be persisted
	id := myKeep.Save(0)  // persist myData value to the collection as a new item
	...
	myKeep.Load(id)  // populate myData with a persisted value
	...
	myKeep.Erase(id)  // remove an item from the collection
	...
	myList := myKeep.List()  // get ids of all items in the collection
	...

Requirements

For a type to be accepted for persistence, it must be based on (or composed of)
numeric types, bool, string, array, slice, map or struct.
Types that contain channels, functions, interfaces or pointers cannot be persisted.

For safety, path parameter in New must be an absolute path to a directory already
existent in the filesystem.

Embedding Keep

Type embedding in Go allows a composition of Keep with your type to feel as your
type has just gained persistence methods:

	var myData struct {
		MyType
		keep.Keep
	}
	myData.Keep = keep.NewOrPanic(&myData.MyData, "/home/user/my_data")

	// populate myData
	myData.MyData = MyData{...}  // by composite literal
	myData.<field> = ...  // by field (in case MyType is a struct)

	id := myData.Save(0)  // persist as new item
	...
	myData.Load(id)  // retrieve a persisted value
	...
	myData.Erase(id)  // remove item from collection
	...
	myList := myData.List()  // get all persisted ids
	...

Issues

A persisted slice loses its original allocation of underlying array; on recovery,
a new array with same length of the slice is created so the slice can reference it.
Consequently, the capacity of a recovered slice equals its length.

Platform independency of filesystem persisted data is not a design goal of package
keep. If so happens at any stage of implementation, it's purely incidental
and can't be assumed to remain.

Bugs

Concurrent access to a collection is not yet tought of, and can be a
fruitful source of weirdness.

Wish List

Document filesystem guidelines for better performance with package keep.

Investigate a possibly faster implementation than encoding/binary for
(un)marshaling persisted values.

Protect against concurrent access to a collection.

Implement key mechanism for sorting items and fast lookup.

*/
package keep

import "fmt"
import "os"
import pth "path"
import "log"
import "io"
import "reflect"
import "errors"
import "unsafe"
import "bytes"
import "encoding/binary"

// Keep handles a collection of persisted values of a type.
type Keep struct {
	access  unsafe.Pointer
	typ     reflect.Type
	path    string
	marshal marshalFn
}

// marshalFn converts the referenced value to a serial representation.
type marshalFn func(unsafe.Pointer) []byte

// newMarshal builds a marshal function dedicated to a specific type.
func newMarshal(t reflect.Type) (marshalFn, error) {
	return func(p unsafe.Pointer) []byte {
		v := reflect.NewAt(t, p).Elem()
		d := v.Interface()
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, d)
		if err != nil {
			panic(fmt.Sprintf("cannot marshal %s: %s", t, err))
		}
		return buf.Bytes()
	}, nil
}

// New creates a Keep value that manages a collection of persisted values of a
// type.
//
// The access parameter must be a pointer to a variable with type that can
// be persisted (see Requirements).
// The variable referenced by access (ie, *access) serves as an access point
// of values to be persisted and recovered (see Save and Load).
//
// The path parameter is an absolute path to a directory in the filesystem
// for storing the collection. If it's the first time this directory is used by
// package keep, it must be empty.
func New(access interface{}, path string) (Keep, error) {
	if !pth.IsAbs(path) {
		return Keep{}, errors.New(fmt.Sprintf("path '%s' is not absolute", path))
	}
	path = pth.Clean(path)
	finfo, err := os.Stat(path)
	if err != nil {
		return Keep{}, errors.New(fmt.Sprintf("path '%s' is unreachable: %s", path, err))
	}
	if !finfo.IsDir() {
		return Keep{}, errors.New(fmt.Sprintf("path '%s' is not a directory", path))
	}
	keepDir := pth.Join(path, ".keep")
	keepDirExists := true
	finfo, err = os.Stat(keepDir)
	if err != nil {
		if os.IsNotExist(err) {
			keepDirExists = false
		} else {
			return Keep{}, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", keepDir, err))
		}
	}
	if keepDirExists {
		if !finfo.IsDir() {
			return Keep{}, errors.New(fmt.Sprintf("keep db dir '%s' is not a directory", keepDir))
		}
	} else {
		dir, err := os.Open(path)
		if err != nil {
			return Keep{}, errors.New(fmt.Sprintf("cannot open '%s': %s", path, err))
		}
		defer dir.Close()
		_, err = dir.Readdir(1)
		if err != io.EOF {
			return Keep{}, errors.New(fmt.Sprintf("path '%s' is not empty and is not a keep db", path))
		}
		err = os.Mkdir(keepDir, 0755)
		if err != nil {
			return Keep{}, errors.New(fmt.Sprintf("cannot create keep db dir '%s'", keepDir))
		}
		log.Printf("keep database initialized in '%s'", path)
	}
	v := reflect.ValueOf(access)
	if v.Kind() != reflect.Ptr {
		return Keep{}, errors.New("access parameter is not a pointer")
	}
	v = v.Elem()
	p := unsafe.Pointer(v.UnsafeAddr())
	t := v.Type()
	marshal, err := newMarshal(t)
	if err != nil {
		return Keep{}, errors.New(fmt.Sprintf("cannot build marshal function for type %s: %s", t, err))
	}
	return Keep{
		path:    keepDir,
		access:  p,
		typ:     t,
		marshal: marshal,
	}, nil
}

// NewOrPanic is a wrapper around New that panics on error.
func NewOrPanic(access interface{}, path string) Keep {
	k, err := New(access, path)
	if err != nil {
		panic(fmt.Sprintf("keep.New failed: %s", err))
	}
	return k
}

// Save creates or updates an item in the collection.
//
// The item to be created or updated is identified by a non zero id parameter.
// The value to be persisted is taken from the access variable (see New).
//
// If id is zero, a new sequential id is chosen for the item.
//
// Returns the id of the item created or updated.
func (k Keep) Save(id uint) (uint, error) {
	v := reflect.NewAt(k.typ, k.access)
	data := v.Elem().Interface()
	fmt.Printf("%s save id %v in %v: %v -> % x\n", k.typ, id, k.path, data, k.marshal(k.access))
	return 0, errors.New("not yet implemented")
}

// Load retrieves the value of a persisted item in the collection.
//
// The retrieved value is stored in the access variable (see New).
func (k Keep) Load(id uint) error {
	v := reflect.NewAt(k.typ, k.access)
	data := v.Elem().Interface()
	fmt.Printf("%s load id %v in %v: %v\n", k.typ, id, k.path, data)
	return errors.New("not yet implemented")
}

// Erase erases an item from the collection.
func (k Keep) Erase(id uint) error {
	return errors.New("not yet implemented")
}

// Exists verifies if an item exists in the collection.
func (k Keep) Exists(id uint) (bool, error) {
	return false, errors.New("not yet implemented")
}

// Wipe removes a collection from the filesystem.
func Wipe(path string) error {
	return errors.New("not yet implemented")
}
