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

import "io"

type Uint32 struct{ store *uint32 }

func NewUint32(store *uint32) Uint32 {
	return Uint32{store: store}
}

func (Uint32) Signature() string {
	return "uint32"
}

func (self Uint32) Marshal(w io.Writer) (int, error) {
	aux := *self.store
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
	*self.store = 0
	for i := 0; i < 4; i++ {
		*self.store *= 0x100
		*self.store += uint32(bs[3-i])
	}
	return n, nil
}
