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

/*
Package binary implements binary serialization of Go types.

*/
package binary

import "fmt"
import "io"
import "reflect"

type Encoder interface {
	Signature() string
	Marshal(io.Writer) (int, error)
	Unmarshal(io.Reader) (int, error)
}

func NewEncoder(data interface{}) (Encoder, error) {
	return MakeEncoder(reflect.ValueOf(data))
}

func MakeEncoder(v reflect.Value) (Encoder, error) {
	var err error
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("storage variable must be passed by reference")
	}
	k := v.Elem().Kind()
	switch k {
	default:
		return nil, fmt.Errorf("unsupported data type: %s", k)
	case reflect.Uint32:
		return Uint32Encoder{v.Interface().(*uint32)}, nil
	case reflect.Uint64:
		return Uint64Encoder{v.Interface().(*uint64)}, nil
	case reflect.Struct:
		v = v.Elem()
		n := v.NumField()
		store := make([]Encoder, n, n)
		for i := 0; i < n; i++ {
			f := v.Type().Field(i)
			if f.PkgPath != "" {
				return nil, fmt.Errorf("struct field '%s' is unexported", f.Name)
			}
			store[i], err = MakeEncoder(v.Field(i).Addr())
			if err != nil {
				return nil, fmt.Errorf("cannot make encoder for struct field %s: %s", v.Type().Field(i).Name, err)
			}
		}
		return StructEncoder{store}, nil
	}
}
