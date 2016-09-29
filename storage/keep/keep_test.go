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
	"os"
	"testing"
	"github.com/coolparadox/go/storage/keep"
)

var myPath string

func init() {
	flag.StringVar(&myPath, "dir", "/tmp/my_data", "path to Keep collection")
}

var k keep.Keep

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

var myData int64

func TestNewEmpty(t *testing.T) {
	var err error
	k, err = keep.New(&myData, myPath)
	if err != nil {
		t.Fatalf("keep.New failed: %s", err)
	}
}

func TestNewNotEmpty(t *testing.T) {
	var err error
	k, err = keep.New(&myData, myPath)
	if err != nil {
		t.Fatalf("keep.New failed: %s", err)
	}
}

