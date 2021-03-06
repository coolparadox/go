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

package lazydb_test

import "github.com/coolparadox/go/storage/lazydb"
import "github.com/coolparadox/go/sort/uint32slice"
import "testing"
import "os"
import "bytes"
import "math/rand"
import "time"
import "io"
import "flag"
import "fmt"

var myPath string
var howManySaves uint
var keyBase uint
var maxSlots uint
var maxSequenceLength uint

func init() {
	flag.StringVar(&myPath, "dir", "/tmp/my_data", "path to LazyDB collection")
	flag.UintVar(&howManySaves, "saves", 1000, "how many keys to create")
	flag.UintVar(&keyBase, "base", 0, "numeric base of key components")
	flag.UintVar(&maxSlots, "slots", 32, "maximum number of value slots")
	flag.UintVar(&maxSequenceLength, "bytes", 1024, "maximum number of bytes per value slot")
	rand.Seed(time.Now().Unix())
}

var db lazydb.LazyDB

func TestInit(t *testing.T) {
	t.Logf("path to lazydb collection = '%s'", myPath)
	t.Logf("save test count = %v", howManySaves)
	t.Logf("numeric base of key components = %v", keyBase)
	t.Logf("maximum number of value slots = %v", maxSlots)
	t.Logf("maximum number of bytes per value slot = %v", maxSequenceLength)
	var err error
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Fatalf("cannot create directory '%s': %s", myPath, err)
	}
}

func TestWipeEmpty(t *testing.T) {
	var err error
	err = lazydb.Wipe(myPath)
	if err != nil {
		t.Fatalf("lazydb.Wipe failed: %s", err)
	}
	file, err := os.Open(myPath)
	if err != nil {
		t.Fatalf("cannot open '%s': %s", myPath, err)
	}
	defer file.Close()
	_, err = file.Readdir(1)
	if err != io.EOF {
		t.Fatalf("lazydb.Wipe did not empty directory '%s'", myPath)
	}
}

func TestNewEmpty(t *testing.T) {
	var err error
	db, err = lazydb.New(myPath, uint32(keyBase))
	if err != nil {
		t.Fatalf("lazydb.New failed in creating a new database: %s", err)
	}
}

func TestNewNotEmpty(t *testing.T) {
	var err error
	_, err = lazydb.New(myPath, rand.Uint32())
	if err != nil {
		t.Fatalf("lazydb.New failed in opening an existent database: %s", err)
	}
}

func TestSaveAs(t *testing.T) {
	sample := make([]byte, 100)
	for i := range sample {
		sample[i] = byte(rand.Intn(256))
	}
	var err error
	src := make([]io.Reader, 1)
	src[0] = bytes.NewReader(sample)
	_, err = db.SaveAs(0, src)
	if err != nil {
		t.Fatalf("lazydb.SaveAs failed: %s", err)
	}
	dst := make([]io.Writer, 1)
	loaded := new(bytes.Buffer)
	dst[0] = loaded
	_, err = db.Load(0, dst)
	if err != nil {
		t.Fatalf("lazydb.Load failed: %s", err)
	}
	if !bytes.Equal(loaded.Bytes(), sample) {
		t.Fatalf("save & load mismatch: saved %v loaded %v", sample, loaded)
	}

	src[0] = bytes.NewReader(sample)
	_, err = db.SaveAs(lazydb.MaxKey, src)
	if err != nil {
		t.Fatalf("lazydb.SaveAs failed: %s", err)
	}
	loaded = new(bytes.Buffer)
	dst[0] = loaded
	_, err = db.Load(lazydb.MaxKey, dst)
	if err != nil {
		t.Fatalf("lazydb.Load failed: %s", err)
	}
	if !bytes.Equal(loaded.Bytes(), sample) {
		t.Fatalf("save & load mismatch: saved %v loaded %v", sample, loaded)
	}

	err = db.Erase(0)
	if err != nil {
		t.Fatalf("lazydb.Erase(0) failed: %s", err)
	}

	err = db.Erase(lazydb.MaxKey)
	if err != nil {
		t.Fatalf("lazydb.Erase(lazydb.MaxKey) failed: %s", err)
	}

}

type savedItem struct {
	key   uint32
	value [][]byte
}

var savedData []savedItem

func TestSaveMany(t *testing.T) {
	var err error
	savedData = make([]savedItem, howManySaves)
	savedKeys := make(map[uint32]interface{})
	for saveIdx := 0; saveIdx < int(howManySaves); saveIdx++ {
		howManySlots := uint(rand.Float64() * float64(maxSlots))
		value := make([][]byte, howManySlots)
		src := make([]io.Reader, howManySlots)
		for slotIdx := 0; slotIdx < int(howManySlots); slotIdx++ {
			if rand.Float64() < 0.5 {
				continue
			}
			sequence := make([]byte, int(rand.Float64()*float64(maxSequenceLength)))
			rand.Read(sequence)
			value[slotIdx] = sequence
			src[slotIdx] = bytes.NewReader(value[slotIdx])
		}
		var key uint32
		var counts []int64
		if saveIdx%2 == 0 {
			// test lazydb.SaveAs
			for {
				key = rand.Uint32()
				if key >= lazydb.MaxKey {
					continue
				}
				_, ok := savedKeys[key]
				if !ok {
					break
				}
			}
			counts, err = db.SaveAs(key, src)
			if err != nil {
				t.Fatalf("lazydb.SaveAs failed: %s", err)
			}
		} else {
			// test lazydb.Save
			key, counts, err = db.Save(src)
			if err != nil {
				t.Fatalf("lazydb.Save failed: %s", err)
			}
		}
		for slotIdx, sequence := range value {
			if int64(len(sequence)) != counts[slotIdx] {
				t.Fatalf("byte count mismatch in key %u slot %i: expected %i received %i", key, slotIdx, len(sequence), counts[slotIdx])
			}
		}
		savedKeys[key] = nil
		savedData[saveIdx].key = key
		savedData[saveIdx].value = value
	}
}

func TestLoadMany(t *testing.T) {
	for _, savedItem := range savedData {
		key := savedItem.key
		loaded := make([]*bytes.Buffer, len(savedItem.value))
		dst := make([]io.Writer, len(loaded))
		for slotIdx, savedSeq := range savedItem.value {
			if savedSeq != nil {
				loaded[slotIdx] = new(bytes.Buffer)
				dst[slotIdx] = loaded[slotIdx]
			}
		}
		counts, err := db.Load(key, dst)
		if err != nil {
			t.Fatalf("lazydb.Load failed: %s", err)
		}
		for slotIdx, savedSeq := range savedItem.value {
			if savedSeq == nil {
				continue
			}
			if int64(len(savedSeq)) != counts[slotIdx] {
				t.Fatalf("byte count mismatch in key %u slot %i: expected %i received %i", key, slotIdx, len(savedSeq), counts[slotIdx])
			}
			loadedSeq := loaded[slotIdx].Bytes()
			if !bytes.Equal(savedSeq, loadedSeq) {
				t.Fatalf("save & load mismatch in key %u slot %i: saved %v loaded %v", key, slotIdx, savedSeq, loadedSeq)
			}
		}
	}
}

func TestKeyListAscending(t *testing.T) {
	var receivedKeys []uint32
	key, err := db.FindKey(0, true)
	for err == nil {
		//t.Logf("found key: %v", key)
		receivedKeys = append(receivedKeys, key)
		if key >= lazydb.MaxKey {
			break
		}
		key, err = db.FindKey(key+1, true)
	}
	if err != nil && err != lazydb.KeyNotFoundError {
		t.Fatalf("lazydb.FindKey failed: %s", err)
	}
	var savedKeys []uint32
	for _, data := range savedData {
		savedKeys = append(savedKeys, data.key)
	}
	rl := len(receivedKeys)
	sl := len(savedKeys)
	if rl != sl {
		t.Fatalf("received key length mismatch: received %v expected %v", rl, sl)
	}
	uint32slice.SortUint32s(savedKeys)
	for i, rk := range receivedKeys {
		sk := savedKeys[i]
		if sk != rk {
			t.Fatalf("received key mismatch: received %v expected %v", rk, sk)
		}
	}
}

func TestKeyListDescending(t *testing.T) {
	var receivedKeys []uint32
	key, err := db.FindKey(lazydb.MaxKey, false)
	for err == nil {
		//t.Logf("found key: %v", key)
		receivedKeys = append(receivedKeys, key)
		if key <= 0 {
			break
		}
		key, err = db.FindKey(key-1, false)
	}
	if err != nil && err != lazydb.KeyNotFoundError {
		t.Fatalf("lazydb.FindKey failed: %s", err)
	}
	var savedKeys []uint32
	for _, data := range savedData {
		savedKeys = append(savedKeys, data.key)
	}
	rl := len(receivedKeys)
	sl := len(savedKeys)
	if rl != sl {
		t.Fatalf("received key length mismatch: received %v expected %v", rl, sl)
	}
	uint32slice.ReversedSortUint32s(savedKeys)
	for i, rk := range receivedKeys {
		sk := savedKeys[i]
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
			t.Fatalf("lazydb.Erase failed: %s", err)
		}
	}
}

func TestExists(t *testing.T) {
	limit := howManySaves / 10
	for saveIdx, savedItem := range savedData {
		key := savedItem.key
		var slot int = -1
		for slotIdx, savedSeq := range savedItem.value {
			if savedSeq == nil {
				continue
			}
			slot = slotIdx
			break
		}
		if slot < 0 {
			continue
		}
		expected := saveIdx >= int(limit)
		exists, err := db.Exists(key, uint32(slot))
		if err != nil {
			t.Fatalf("lazydb.Exists failed: %s", err)
		}
		if exists != expected {
			t.Fatalf("lazydb.Exists mismatch for key %v slot %v: expected %v received %v", key, slot, expected, exists)
		}
	}
}

func TestWipeNotEmpty(t *testing.T) {
	var err error
	err = lazydb.Wipe(myPath)
	if err != nil {
		t.Fatalf("lazydb.Wipe failed: %s", err)
	}
}

func Example() {

	// error handling purposely ignored
	// in some places for didactic purposes.

	// Create an empty database
	dbPath := "/tmp/my_db"
	os.MkdirAll(dbPath, 0755)
	lazydb.Wipe(dbPath)
	db, _ := lazydb.New(dbPath, 0)

	// Save values in new keys
	k1, _, _ := db.Save([]io.Reader{bytes.NewReader([]byte("goodbye"))})
	k2, _, _ := db.Save([]io.Reader{bytes.NewReader([]byte("cruel"))})
	k3, _, _ := db.Save([]io.Reader{bytes.NewReader([]byte("world"))})

	// Update, remove
	db.SaveAs(k1, []io.Reader{bytes.NewReader([]byte("hello"))})
	db.Erase(k2)
	db.SaveAs(k3, []io.Reader{bytes.NewReader([]byte("folks"))})

	// Loop through keys
	key, err := db.FindKey(0, true)
	for err == nil {
		// Print value
		val := new(bytes.Buffer)
		db.Load(key, []io.Writer{val})
		fmt.Printf("key %v: %s\n", key, string(val.Bytes()))
		if key >= lazydb.MaxKey {
			// Maximum key reached
			break
		}
		// Find next existent key
		key, err = db.FindKey(key+1, true)
	}
	if err != nil && err != lazydb.KeyNotFoundError {
		// An abnormal error occurred
		panic(err)
	}

	// Output:
	// key 0: hello
	// key 2: folks

}
