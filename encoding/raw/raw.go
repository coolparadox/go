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

/*
Package raw implements binary serialization of Go types based on direct copy of
referenced memory.

*/
package raw

import "fmt"
import "io"
import "reflect"

type Encoder struct {
	EncodingType
}

func New(data interface{}) (Encoder, error) {

	dataType := reflect.TypeOf(data)
	dataKind := dataType.Kind()
	if (dataKind != reflect.Ptr) {
		return Encoder{}, fmt.Errorf("storage variable must be passed by reference")
	}
	dataType = reflect.ValueOf(data).Elem().Type()
	dataKind = dataType.Kind()
	if (dataKind != reflect.Uint32) {
		return Encoder{}, fmt.Errorf("unsupported data type: %s", dataKind)
	}
	var backstore *uint32 = data.(*uint32)
	return Encoder{EncodingType:Uint32{backstore:backstore}}, nil

}

type EncodingType interface {
	Signature() string
	Marshal(io.Writer) (int, error)
	Unmarshal(io.Reader) (int, error)
}

type Uint32 struct { backstore *uint32 }

func (Uint32) Signature() string {
	return "uint32"
}

func (self Uint32) Marshal(w io.Writer) (int, error) {
	aux := *self.backstore
	bs := make([]byte, 4, 4)
	for i := 0; i < 4; i++ {
		bs[i] = byte(aux % 0x100)
		aux /= 0x100
	}
	return w.Write(bs)
}

func (self Uint32) Unmarshal(r io.Reader) (int, error) {
	bs := make([]byte, 4, 4)
	n, err := r.Read(bs)
	if err != nil {
		return n, err
	}
	*self.backstore = 0
	for i := 0; i < 4; i++ {
		*self.backstore *= 0x100
		*self.backstore += uint32(bs[3-i])
	}
	return n, nil
}

