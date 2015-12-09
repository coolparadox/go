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
import "io"

const (
	androidMyPath = "/storage/emulated/0/go/var/my_data"
	otherMyPath   = "/tmp/my_data"
)

const howManySaves = 10000

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
	err = concur.Wipe(myPath)
	if err != nil {
		t.Fatalf("concur.Wipe failed: %s", err)
	}
	file, err := os.Open(myPath)
	if err != nil {
		t.Fatalf("cannot open '%s': %s", myPath, err)
	}
	defer file.Close()
	_, err = file.Readdir(1)
	if err != io.EOF {
		t.Fatalf("concur.Wipe did not empty directory '%s'", myPath)
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

	err = db.Put(0, sample)
	if err != nil {
		t.Fatalf("concur.Put failed: %s", err)
	}
	loaded, err := db.Get(0)
	if err != nil {
		t.Fatalf("concur.Get failed: %s", err)
	}
	if !bytes.Equal(loaded, sample) {
		t.Fatalf("save & load mismatch: saved %v loaded %v", sample, loaded)
	}

	err = db.Put(4294967295, sample)
	if err != nil {
		t.Fatalf("concur.Put failed: %s", err)
	}
	loaded, err = db.Get(4294967295)
	if err != nil {
		t.Fatalf("concur.Get failed: %s", err)
	}
	if !bytes.Equal(loaded, sample) {
		t.Fatalf("save & load mismatch: saved %v loaded %v", sample, loaded)
	}

}

type savedItem struct {
	id    uint32
	value [1]byte
}

var savedData [howManySaves]savedItem

func TestSaveMany(t *testing.T) {
	for i := 0; i < howManySaves; i++ {
		id := uint32(rand.Int63())
		value := byte(id % 256)
		err := db.Put(id, []byte{value})
		if err != nil {
			t.Fatalf("concur.Put failed: %s", err)
		}
		savedData[i].id = id
		savedData[i].value[0] = value
	}
}

func TestLoadMany(t *testing.T) {
	for i := 0; i < howManySaves; i++ {
		id := savedData[i].id
		loaded, err := db.Get(id)
		if err != nil {
			t.Fatalf("concur.Get failed: %s", err)
		}
		saved := savedData[i].value
		if loaded[0] != saved[0] {
			t.Fatalf("save & load mismatch: saved %v loaded %v id %v", saved, loaded, id)
		}
	}
}
