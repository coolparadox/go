// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Binary, a binary encoder of Go types.
//
// Binary is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Binary is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Binary. If not, see <http://www.gnu.org/licenses/>.

package binary_test

import "bytes"
import "time"
import "math/rand"
import "testing"
import "github.com/coolparadox/go/encoding/binary"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestUint16Encoder(t *testing.T) {
	var myData uint16
	expected_signature := "uint16"
	encoder, err := binary.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = uint16(rand.Uint32() % 0x10000)
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
	if myData != myData2 {
	t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestUint32Encoder(t *testing.T) {
	var myData uint32
	expected_signature := "uint32"
	encoder, err := binary.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = rand.Uint32()
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
	if myData != myData2 {
	t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestUint64Encoder(t *testing.T) {
	var myData uint64
	expected_signature := "uint64"
	encoder, err := binary.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = uint64(rand.Uint32())
	myData *= 0x100000000
	myData += uint64(rand.Uint32())
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
	if myData != myData2 {
	t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestStructEncoder(t *testing.T) {
	var myData struct {
		A uint32
		B uint32
		C uint32
	}
	expected_signature := "struct { uint32; uint32; uint32 }"
	encoder, err := binary.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData.A = rand.Uint32()
	myData.B = rand.Uint32()
	myData.C = rand.Uint32()
	_, err = encoder.Marshal(&b)
	if err != nil {
		t.Fatalf("Marshal() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData.A = 0
	myData.B = 0
	myData.C = 0
	var n int
	n, err = encoder.Unmarshal(&b)
	if err != nil {
		t.Fatalf("Unmarshal() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if myData != myData2 {
	t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

