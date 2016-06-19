// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Raw, a binary encoder of Go types based on direct copy
// of memory content.
//
// Raw is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Raw is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Raw. If not, see <http://www.gnu.org/licenses/>.

package raw_test

import "bytes"
import "time"
import "math/rand"
import "testing"
import "github.com/coolparadox/go/encoding/raw"

var myData uint32

func TestNew(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	myData = rand.Uint32()
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	t.Logf("myData type signature = %s", encoder.Signature())
	var b bytes.Buffer
	_, err = encoder.Marshal(&b)
	if err != nil {
		t.Fatalf("Marshal() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = 0
	var n int
	n, err = encoder.Unmarshal(&b)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if (myData != myData2) {
		t.Fatalf("marshal / unmarshal mismatch")
	}

}

