// Copyright 2015 Rafael Lorandi.
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
// along with Keep.  If not, see <http://www.gnu.org/licenses/>.

/*
Package keep offers filesystem persistence for "concrete data" types.
It works with types based on numeric, bool or string types, and also
with any composition of these types by arrays, slices, maps and structs.

Basics

On initialization, package keep looks for environment variable KEEPROOT that
must be an absolute path to a directory in the filesystem for storing persisted
data.

In your code, let's say you have a variable of type MyType containing data you 
want to persist. Create a Keep value for this purpose:

	var myData MyType
	myKeep := keep.New(&myData, "my_data")

In the above example, myKeep opens a collection of MyType values under
directory $KEEPROOT/my_data in the filesystem, using myData as a buffer for
storing/retrieving values to/from the collection:

	myData = ... // populate with data to be persisted
	id := myKeep.Save(0) // persist myData value to the collection as a new item
	...
	myKeep.Load(id) // populate myData with a persisted value
	...
	myKeep.Erase(id) // remove an item from the collection
	...
	myList := myKeep.List() // get ids of all items in the collection
	...

Requirements

* KEEPROOT...
persisted types

Concurreny Issues
Platform independency

...

Todo

replace encoding/binary
add collection keys
document filesystem issues

*/
package keep

import "fmt"
import "os"
import "path"
import "log"
import "io"
import "reflect"
import "errors"
import "unsafe"

var rootDirPath string

func isDirectoryEmpty(name string) (bool, error) {

	dir, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer dir.Close()
	_, err = dir.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err

}

func doesFileExist(name string) (bool, error) {

	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func init() {

	// Make sure keep root path is sane
	var ok bool
	rootDirPath, ok = os.LookupEnv("KEEPROOT")
	if !ok {
		panic("undefined KEEPROOT environment variable")
	}
	log.Printf("KEEPROOT is '%s'", rootDirPath)
	if !path.IsAbs(rootDirPath) {
		panic(fmt.Sprintf("root dir '%s' is not an absolute path", rootDirPath))
	}
	rootDirPath = path.Clean(rootDirPath)
	finfo, err := os.Stat(rootDirPath)
	if err != nil {
		panic(fmt.Sprintf("root dir '%s' unreachable: %s", rootDirPath, err))
	}
	if !finfo.IsDir() {
		panic(fmt.Sprintf("root dir '%s' is not a directory", rootDirPath))
	}
	if finfo.Mode().Perm()&0022 != 0 {
		panic(fmt.Sprintf("root dir '%s' is group or world writable", rootDirPath))
	}
	rootFilePath := path.Join(rootDirPath, ".keepRoot")
	rootFileExists, err := doesFileExist(rootFilePath)
	if err != nil {
		panic(fmt.Sprintf("cannot check for '%s' existence: %s", rootFilePath, err))
	}
	rootDirEmpty, err := isDirectoryEmpty(rootDirPath)
	if err != nil {
		panic(fmt.Sprintf("cannot check if root dir '%s' is empty: %s", rootDirPath, err))
	}
	if !rootFileExists && !rootDirEmpty {
		panic(fmt.Sprintf("root dir '%s' is not empty and is not a keep database", rootDirPath))
	}
	for dir := path.Dir(rootDirPath); ; dir = path.Dir(dir) {
		if ok, _ = doesFileExist(path.Join(dir, ".keepRoot")); ok {
			panic(fmt.Sprintf("keep database detected in '%s' above root dir '%s'", dir, rootDirPath))
		}
		if dir == "/" {
			break
		}
	}
	if rootDirEmpty {
		rootFile, err := os.Create(rootFilePath)
		if err != nil {
			panic(fmt.Sprintf("cannot create database root file '%s': %s", rootFilePath, err))
		}
		err = rootFile.Chmod(0644)
		if err != nil {
			panic(fmt.Sprintf("cannot chmod root file '%s': %s", rootFilePath, err))
		}
		log.Printf("keep database initialized in '%s'", rootDirPath)
	}

}

// SayHello says something to standard output.
func SayHello() {
	fmt.Printf("keep root '%s' ok\n", rootDirPath)
}

// Keep handles a collection of persisted values of a user type.
type Keep struct {
	addr unsafe.Pointer
	typ  reflect.Type
	home string
}

// New creates a Keep value that manages a collection of persisted values of a
// user type.
// 
// The access parameter must be a pointer to any user variable which type agrees
// with persistency requirements (see ...).
// The type of values persisted in the collection is taken from the type of the
// variable pointed by access (ie, type of *access). Moreover, *access becomes the
// user entry point of values to be persisted and recovered, see Save and Load
// methods.
// 
// The home parameter is a relative directory path that, prefixed by the value of
// KEEPROOT environment variable, forms the location of the collection in the
// filesystem.
func New(access interface{}, home string) (Keep, error) {
	v := reflect.ValueOf(access)
	if v.Kind() != reflect.Ptr {
		return Keep{}, errors.New("keep.New(): access parameter is not a pointer")
	}
	v = v.Elem()
	p := unsafe.Pointer(v.UnsafeAddr())
	t := v.Type()
	return Keep{home: home, addr: p, typ: t}, nil
}

// Save creates or updates an item in the collection.
// The item to be created or updated is identified by a non zero id parameter.
// The value to be persisted is taken from the access variable (see New).
// If id is zero, a new sequential id is chosen for the item.
// Returns the id of the item created or updated.
func (k Keep) Save(id uint) (uint, error) {
	v := reflect.NewAt(k.typ, k.addr)
	data := v.Elem().Interface()
	fmt.Printf("%s save id %v in %v: %v\n", k.typ, id, k.home, data)
	return 0, errors.New("not yet implemented")
}

// Load retrieves the value of a persisted item in the collection.
// The retrieved value is stored in the access variable (see New).
func (k Keep) Load(id uint) error {
	v := reflect.NewAt(k.typ, k.addr)
	data := v.Elem().Interface()
	fmt.Printf("%s load id %v in %v: %v\n", k.typ, id, k.home, data)
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

// Wipe removes a collection from the filesystem, allowing the same home to be
// reused for persistency of another type.
func Wipe(home string) error {
	return errors.New("not yet implemented")
}
