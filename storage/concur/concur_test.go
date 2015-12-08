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

func TestSaveLoad(t *testing.T) {

	rand.Seed(time.Now().Unix())
	sample := make([]byte, 100)
	for i, _ := range sample {
		sample[i] = byte(rand.Intn(256))
	}
	var err error
	myPath := "/tmp/my_data"
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Logf("cannot create directory '%s'; assuming Android", myPath)
		myPath = "/storage/emulated/0/go/var/my_data"
		err = os.MkdirAll(myPath, 0755)
		if err != nil {
			t.Fatalf("cannot create directory '%s': %s", myPath, err)
		}
	}
	db, err := concur.New(myPath)
	if err != nil {
		t.Fatalf("concur.New failed: %s", err)
	}
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
