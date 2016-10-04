// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Keep, a storage library of typed data for the Go
// language.
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

package keep_test

import (
	"flag"
	"fmt"
	"github.com/coolparadox/go/storage/keep"
	"os"
	"testing"
)

var myPath string

func init() {
	flag.StringVar(&myPath, "dir", "/tmp/my_data", "path to Keep collection")
}

func TestInit(t *testing.T) {
	var err error
	t.Logf("path to Keep collection = '%s'", myPath)
	err = os.RemoveAll(myPath)
	if err != nil {
		t.Fatalf("cannot remove directory '%s': %s", myPath, err)
	}
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Fatalf("cannot create directory '%s': %s", myPath, err)
	}
}

type MyType struct {
	X int64
}

var myData struct {
	MyType
	keep.Keep
}

func TestNewEmpty(t *testing.T) {
	var err error
	myData.Keep, err = keep.New(&myData.MyType, myPath)
	if err != nil {
		t.Fatalf("keep.New failed: %s", err)
	}
}

func TestNewNotEmpty(t *testing.T) {
	var err error
	myData.Keep, err = keep.New(&myData.MyType, myPath)
	if err != nil {
		t.Fatalf("keep.New failed: %s", err)
	}
}

func TestSignature(t *testing.T) {
	t.Logf("type signature: %s", myData.Signature())
}

func TestSaveAs(t *testing.T) {
	var err error
	myData.X = 8765
	err = myData.SaveAs(1)
	if err != nil {
		t.Fatalf("keep.SaveAs failed: %s", err)
	}
}

func TestLoad(t *testing.T) {
	var err error
	myData.X = 0
	err = myData.Load(1)
	if err != nil {
		t.Fatalf("keep.Load failed: %s", err)
	}
	if myData.X != 8765 {
		t.Fatalf("Load mismatch: expected 8765, received %s", myData.X)
	}
}

func TestExistsFalse(t *testing.T) {
	var err error
	ok, err := myData.Exists(2)
	if err != nil {
		t.Fatalf("Exists failed: %s", err)
	}
	if ok {
		t.Fatalf("Exists result mismatch for position 2: expected false, received true")
	}
}

func TestExistsTrue(t *testing.T) {
	var err error
	ok, err := myData.Exists(1)
	if err != nil {
		t.Fatalf("Exists failed: %s", err)
	}
	if !ok {
		t.Fatalf("Exists result mismatch for position 1: expected true, received false")
	}
}

func TestSave(t *testing.T) {
	var err error
	myData.X = 10234
	pos, err := myData.Save()
	if err != nil {
		t.Fatalf("keep.Save failed: %s", err)
	}
	if pos != 2 {
		t.Fatalf("keep.Save position mismatch: expected 2, received %v", pos)
	}
	myData.X = 0
	myData.Load(pos)
	if err != nil {
		t.Fatalf("keep.Load failed: %s", err)
	}
	if myData.X != 10234 {
		t.Fatalf("Load mismatch: expected 10234, received %s", myData.X)
	}
}

func TestErase(t *testing.T) {
	var err error
	TestExistsTrue(t)
	err = myData.Erase(2)
	if err != nil {
		t.Fatalf("Erase failed: %s", err)
	}
	TestExistsFalse(t)
}

func TestFindPos(t *testing.T) {
	var err error
	var pos uint32
	myData.SaveAs(1000)
	pos, err = myData.FindPos(1, true)
	if err != nil {
		t.Fatalf("FindPos failed: %s", err)
	}
	if pos != 1 {
		t.Fatalf("FindPos mismatch: expected 1, received %v", pos)
	}
	pos, err = myData.FindPos(2, true)
	if err != nil {
		t.Fatalf("FindPos failed: %s", err)
	}
	if pos != 1000 {
		t.Fatalf("FindPos mismatch: expected 1000, received %v", pos)
	}
	pos, err = myData.FindPos(1001, true)
	if err == nil {
		t.Fatalf("FindPos with no available positions returned no error")
	}
	if err != keep.PosNotFoundError {
		t.Fatalf("FindPos error mismatch: expected PosNotFoundError, received %v", err)
	}
}

func Example() {

	// Let's say we want to keep a collection of strings.
	// Define a variable to be a placeholder for accessing
	// collection's values
	var myData string

	// Create a collection for your data
	os.RemoveAll("/tmp/my_data")
	os.MkdirAll("/tmp/my_data", 0755)
	k, err := keep.New(&myData, "/tmp/my_data")
	if err != nil {
		panic(err)
	}

	// Save values in new positions
	myData = "goodbye"
	k1, _ := k.Save()
	myData = "cruel"
	k2, _ := k.Save()
	myData = "world"
	k3, _ := k.Save()

	// Update, remove
	myData = "hello"
	k.SaveAs(k1)
	k.Erase(k2)
	myData = "folks"
	k.SaveAs(k3)

	// Loop though positions
	pos, err := k.FindPos(1, true)
	for err == nil {
		// Retrieve value
		k.Load(pos)
		fmt.Printf("position %v: %s\n", pos, myData)
		if pos >= keep.MaxPos {
			// Maximum position reached
			break
		}
		// Find next filled position
		pos, err = k.FindPos(pos+1, true)
	}
	if err != nil && err != keep.PosNotFoundError {
		// An abnormal error occurred
		panic(err)
	}

	// Output:
	// position 1: hello
	// position 3: folks

}

func Example_embedding() {

	// Let's say we want to keep a collection of strings.
	type MyType struct {
		V string
	}

	// Embedding allows accessing placeholder and
	// manipulating the collection from the same variable.
	var myData struct {
		MyType
		keep.Keep
	}

	// Create a collection for your data
	os.RemoveAll("/tmp/my_data")
	os.MkdirAll("/tmp/my_data", 0755)
	var err error
	myData.Keep, err = keep.New(&myData.MyType, "/tmp/my_data")
	if err != nil {
		panic(err)
	}

	// Save values in new positions
	myData.V = "goodbye"
	k1, _ := myData.Save()
	myData.V = "cruel"
	k2, _ := myData.Save()
	myData.V = "world"
	k3, _ := myData.Save()

	// Update, remove
	myData.V = "hello"
	myData.SaveAs(k1)
	myData.Erase(k2)
	myData.V = "folks"
	myData.SaveAs(k3)

	// Loop though positions
	pos, err := myData.FindPos(1, true)
	for err == nil {
		// Retrieve value
		myData.Load(pos)
		fmt.Printf("position %v: %s\n", pos, myData.V)
		if pos >= keep.MaxPos {
			// Maximum position reached
			break
		}
		// Find next filled position
		pos, err = myData.FindPos(pos+1, true)
	}
	if err != nil && err != keep.PosNotFoundError {
		// An abnormal error occurred
		panic(err)
	}

	// Output:
	// position 1: hello
	// position 3: folks

}
