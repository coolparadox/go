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
import "flag"
import "sort"

var myPath string
var howManySaves uint

func init() {
	flag.StringVar(&myPath, "dir", "/tmp/my_data", "path to concur collection")
	flag.UintVar(&howManySaves, "saves", 1000, "how many keys to create")
}

var db concur.Concur

func TestInit(t *testing.T) {
	var err error
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Fatalf("cannot create directory '%s': %s", myPath, err)
	}
	t.Logf("path to concur db = '%s'", myPath)
	t.Logf("save test count = %v", howManySaves)
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
	key   uint32
	value [1]byte
}

var savedData []savedItem

func TestSaveMany(t *testing.T) {
	savedData = make([]savedItem, howManySaves)
	savedKeys := make(map[uint32]interface{})
	for i := uint(0); i < howManySaves; i++ {
		var key uint32
		for {
			key = uint32(rand.Int63())
			_, ok := savedKeys[key]
			if !ok {
				break
			}
		}
		savedKeys[key] = nil
		value := byte(key % 256)
		err := db.Put(key, []byte{value})
		if err != nil {
			t.Fatalf("concur.Put failed: %s", err)
		}
		savedData[i].key = key
		savedData[i].value[0] = value
	}
}

func TestLoadMany(t *testing.T) {
	for i := uint(0); i < howManySaves; i++ {
		key := savedData[i].key
		loaded, err := db.Get(key)
		if err != nil {
			t.Fatalf("concur.Get failed: %s", err)
		}
		saved := savedData[i].value
		if loaded[0] != saved[0] {
			t.Fatalf("save & load mismatch: saved %v loaded %v key %v", saved, loaded, key)
		}
	}
}

func TestKeyList(t *testing.T) {
	key, ok, err := db.SmallestKeyNotLessThan(0)
	if err != nil {
		t.Fatalf("concur.SmallestKeyNotLessThan failed: %s", err)
	}
	if !ok {
		t.Fatalf("empty database!?")
	}
	receivedKeys := make([]uint32, 0)
	for ok {
		receivedKeys = append(receivedKeys, key)
		if key == concur.KeyMax {
			break
		}
		key, ok, err = db.SmallestKeyNotLessThan(key+1)
		if err != nil {
			t.Fatalf("concur.SmallestKeyNotLessThan failed: %s", err)
		}
	}
	savedKeys := make([]int, 0)
	for _, data := range savedData {
		savedKeys = append(savedKeys, int(data.key))
	}
	rl := len(receivedKeys)
	sl := len(savedKeys)
	if rl != sl {
		t.Fatalf("received key length mismatch: received %v expected %v", rl, sl)
	}
	sort.Ints(savedKeys)
	for i, rk := range receivedKeys {
		sk := uint32(savedKeys[i])
		if sk != rk {
			t.Fatalf("received key mismatch: received %v expected %v", rk, sk)
		}
	}
}

func TestErase(t *testing.T) {
	limit := howManySaves / 10
	for i := uint(0); i < limit; i++ {
		key := savedData[i].key
		err := db.Erase(key)
		if err != nil {
			t.Fatalf("concur.Erase failed: %s", err)
		}
	}
}

func TestExists(t *testing.T) {
	limit := howManySaves / 10
	for i := uint(0); i < howManySaves; i++ {
		key := savedData[i].key
		exists, err := db.Exists(key)
		if err != nil {
			t.Fatalf("concur.Exists failed: %s", err)
		}
		if i < limit {
			if exists {
				t.Fatalf("concur.Exists mismatch for key %v: %v", key, exists)
			}
		} else {
			if !exists {
				t.Fatalf("concur.Exists mismatch for key %v: %v", key, exists)
			}
		}
	}
}
