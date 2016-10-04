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
