// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Raw, a binary encoder of Go types.
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

import (
	"bytes"
	"fmt"
	"github.com/coolparadox/go/encoding/raw"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const bufLen = 256

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random_uint8() uint8 {
	return uint8(rand.Uint32() % 0x100)
}

func random_uint16() uint16 {
	return uint16(rand.Uint32() % 0x10000)
}

func random_uint32() uint32 {
	return rand.Uint32()
}

func random_uint64() uint64 {
	return uint64(rand.Uint32())*0x100000000 + uint64(rand.Uint32())
}

func random_int8() int8 {
	aux := random_uint8()
	if aux >= (1 + 0x7F) {
		return int8(aux - 1 - 0x7F)
	} else {
		return int8(aux) - 1 - 0x7F
	}
}

func random_int16() int16 {
	aux := random_uint16()
	if aux >= (1 + 0x7FFF) {
		return int16(aux - 1 - 0x7FFF)
	} else {
		return int16(aux) - 1 - 0x7FFF
	}
}

func random_int32() int32 {
	aux := random_uint32()
	if aux >= (1 + 0x7FFFFFFF) {
		return int32(aux - 1 - 0x7FFFFFFF)
	} else {
		return int32(aux) - 1 - 0x7FFFFFFF
	}
}

func random_int64() int64 {
	aux := random_uint64()
	if aux >= (1 + 0x7FFFFFFFFFFFFFFF) {
		return int64(aux - 1 - 0x7FFFFFFFFFFFFFFF)
	} else {
		return int64(aux) - 1 - 0x7FFFFFFFFFFFFFFF
	}
}

func random_float32() float32 {
	return (rand.Float32() - rand.Float32()) / rand.Float32()
}

func random_float64() float64 {
	return (rand.Float64() - rand.Float64()) / rand.Float64()
}

func TestUint8Encoder(t *testing.T) {
	var myData uint8
	expected_signature := "uint8"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_uint8()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestInt8Encoder(t *testing.T) {
	var myData int8
	expected_signature := "int8"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_int8()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestUint16Encoder(t *testing.T) {
	var myData uint16
	expected_signature := "uint16"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_uint16()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestInt16Encoder(t *testing.T) {
	var myData int16
	expected_signature := "int16"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_int16()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestUint32Encoder(t *testing.T) {
	var myData uint32
	expected_signature := "uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_uint32()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestInt32Encoder(t *testing.T) {
	var myData int32
	expected_signature := "int32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_int32()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestUint64Encoder(t *testing.T) {
	var myData uint64
	expected_signature := "uint64"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_uint64()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestInt64Encoder(t *testing.T) {
	var myData int64
	expected_signature := "int64"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = random_int64()
	b := make([]byte, bufLen)
	n, err := encoder.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received marshal length")
	}
	t.Logf("marshal %v --> %v", myData, b[:n])
	backupData := myData
	backupN := n
	myData = 0
	n, err = encoder.Write(b)
	if err != nil {
		t.Fatalf("Write() failed: %s", err)
	}
	if n <= 0 {
		t.Fatalf("zero received unmarshal length")
	}
	if (n != backupN) {
		t.Fatalf("received byte lenghts mismatch: marshal %v, unmarshal %v", backupN, n)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, backupData) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", backupData, myData)
	}
}

func TestStructEncoder(t *testing.T) {
	var myData struct {
		A uint32
		B int64
		C uint8
	}
	expected_signature := "struct { uint32; int64; uint8 }"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData.A = random_uint32()
	myData.B = random_int64()
	myData.C = random_uint8()
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData.A = 0
	myData.B = 0
	myData.C = 0
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestSliceEncoder(t *testing.T) {
	m := int(random_uint8()%10 + 1)
	myData := make([]uint32, m, m)
	expected_signature := "[]uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	for i := 0; i < len(myData); i++ {
		myData[i] = random_uint32()
	}
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = nil
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestMapEncoder(t *testing.T) {
	m := int(random_uint8()%10 + 1)
	myData := make(map[uint8]uint32, m)
	expected_signature := "map[uint8]uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	for i := 0; i < m; i++ {
		myData[random_uint8()] = random_uint32()
	}
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = nil
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestArrayEncoder(t *testing.T) {
	const alen = 5
	var myData [alen]uint32
	expected_signature := "[" + strconv.Itoa(alen) + "]uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	for i := 0; i < alen; i++ {
		myData[i] = random_uint32()
	}
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = reflect.Zero(reflect.TypeOf(myData)).Interface().([alen]uint32)
	n, err := encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestStringEncoder(t *testing.T) {
	myData := "The Quick Brown Fox Jumps Over The Lazy Dog"
	expected_signature := "string"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = ""
	n, err := encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestBoolEncoder(t *testing.T) {
	myData := true
	expected_signature := "bool"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = !myData
	n, err := encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestFloat32Encoder(t *testing.T) {
	var myData float32
	expected_signature := "float32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = random_float32()
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = 0
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestFloat64Encoder(t *testing.T) {
	var myData float64
	expected_signature := "float64"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = random_float64()
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = 0
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestComplex64Encoder(t *testing.T) {
	var myData complex64
	expected_signature := "complex64"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = complex(random_float32(), random_float32())
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = 0
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestComplex128Encoder(t *testing.T) {
	var myData complex128
	expected_signature := "complex128"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = complex(random_float64(), random_float64())
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = 0
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestPtrEncoderNil(t *testing.T) {
	var myData *uint32
	expected_signature := "*uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = new(uint32)
	*myData = random_uint32()
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}

func TestPtrEncoder(t *testing.T) {
	var myData *uint32
	expected_signature := "*uint32"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	var b bytes.Buffer
	myData = new(uint32)
	*myData = random_uint32()
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal &(%v)@%v --> %v", *myData, myData, b.Bytes())
	myData2 := myData
	myData = nil
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> &(%v)@%v", n, *myData, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected &%v, received &%v", *myData2, *myData)
	}
}

func Example() {

	// Let's say we want to serialize a slice of strings
	var myData []string

	// Bind an encoder to the placeholder variable
	encoder, err := raw.New(&myData)
	if err != nil {
		panic("failed to create encoder: " + err.Error())
	}
	fmt.Printf("encoder signature: %s\n", encoder.Signature())

	// Let's populate placeholder for demonstration
	myData = make([]string, 2)
	myData[0] = "hello"
	myData[1] = "world"
	fmt.Printf("original data: %v\n", myData)

	// Serialize placeholder contents
	var buf bytes.Buffer
	_, err = encoder.WriteTo(&buf)
	if err != nil {
		panic("failed to marshal: " + err.Error())
	}

	// Mess up with placeholder for demonstrating recovery
	myData = make([]string, 3)
	myData[0] = "goodbye"
	myData[1] = "cruel"
	myData[2] = "world"

	// Recover data from serial representation
	_, err = encoder.ReadFrom(&buf)
	if err != nil {
		panic("failed to unmarshal: " + err.Error())
	}
	fmt.Printf("recovered data: %v\n", myData)

	// Output:
	// encoder signature: []string
	// original data: [hello world]
	// recovered data: [hello world]

}

func TestSliceSliceStructEncoder(t *testing.T) {
	type MyStruct struct {
		A uint8
		B float32
	}
	m := int(random_uint8()%10 + 1)
	var myData [][]MyStruct
	expected_signature := "[][]struct { uint8; float32 }"
	encoder, err := raw.New(&myData)
	if err != nil {
		t.Fatalf("New() failed: %s", err)
	}
	signature := encoder.Signature()
	if signature != expected_signature {
		t.Fatalf("signature mismatch: expected '%s', received '%s'", expected_signature, signature)
	}
	t.Logf("myData type signature = %s", signature)
	myData = make([][]MyStruct, m, m)
	for i := 0; i < len(myData); i++ {
		m := int(random_uint8()%10 + 1)
		myData[i] = make([]MyStruct, m, m)
		for j := 0; j < len(myData[i]); j++ {
			myData[i][j].A = random_uint8()
			myData[i][j].B = random_float32()
		}
	}
	var b bytes.Buffer
	_, err = encoder.WriteTo(&b)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err)
	}
	t.Logf("marshal %v --> %v", myData, b.Bytes())
	myData2 := myData
	myData = nil
	var n int64
	n, err = encoder.ReadFrom(&b)
	if err != nil {
		t.Fatalf("ReadFrom() failed: %s", err)
	}
	t.Logf("unmarshal %v bytes --> %v", n, myData)
	if !reflect.DeepEqual(myData, myData2) {
		t.Fatalf("marshal / unmarshal mismatch: expected %v, received %v", myData2, myData)
	}
}
