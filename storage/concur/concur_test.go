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

package concur_test

import "github.com/coolparadox/go/storage/concur"
import "testing"
import "os"
import "bytes"
import "math/rand"
import "time"

const androidMyPath = "/storage/emulated/0/go/var/my_data"
const otherMyPath = "/tmp/my_data"

var myPath string = otherMyPath
var db concur.Concur

func TestInit(t *testing.T) {
	var err error
	myPath = otherMyPath
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Logf("cannot create directory '%s'; assuming Android", myPath)
		myPath = androidMyPath
		err = os.MkdirAll(myPath, 0755)
		if err != nil {
			t.Fatalf("cannot create directory '%s': %s", myPath, err)
		}
	}
	t.Logf("path to concur db is '%s'", myPath)
	rand.Seed(time.Now().Unix())

}

func TestNewEmpty(t *testing.T) {

	var err error
	err = os.RemoveAll(myPath)
	if err != nil {
		t.Fatalf("cannot remove directory '%s': %s", myPath, err)
	}
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Fatalf("cannot create directory '%s': %s", myPath, err)
	}
	db, err = concur.New(myPath)
	if err != nil {
		t.Fatalf("concur.New failed: %s", err)
	}

}

func TestSaveAs(t *testing.T) {
	sample := make([]byte, 100)
	for i, _ := range sample {
		sample[i] = byte(rand.Intn(256))
	}
	var err error
	err = db.SaveAs(sample, 0)
	if err != nil {
		t.Fatalf("concur.SaveAs failed: %s", err)
	}
	loaded, err := db.Load(0)
	if err != nil {
		t.Fatalf("concur.Load failed: %s", err)
	}
	if !bytes.Equal(loaded, sample) {
		t.Fatalf("save & load mismatch: saved %v loaded %v", sample, loaded)
	}
}

type savedItem struct {
	id    uint64
	value [1]byte
}

const howManySaves = 10000

var savedData [howManySaves]savedItem

func TestSaveMany(t *testing.T) {
	for i := 0; i < howManySaves; i++ {
		id := uint64(rand.Int63())
		value := byte(rand.Int() % 256)
		err := db.SaveAs([]byte{value}, id)
		if err != nil {
			t.Fatalf("concur.SaveAs failed: %s", err)
		}
		savedData[i].id = id
		savedData[i].value[0] = value
	}
}

func TestLoadMany(t *testing.T) {
	for i := 0; i < howManySaves; i++ {
		id := savedData[i].id
		loaded, err := db.Load(id)
		if err != nil {
			t.Fatalf("concur.Load failed: %s", err)
		}
		saved := savedData[i].value
		if loaded[0] != saved[0] {
			t.Fatalf("save & load mismatch: saved %v loaded %v", saved, loaded)
		}
	}
}
