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

package binary

import "io"

type Int8Encoder struct{ store *int8 }

func (Int8Encoder) Signature() string {
	return "int8"
}

func (self Int8Encoder) Marshal(w io.Writer) (int, error) {
	var aux uint8
	if *self.store >= 0 {
		aux = uint8(*self.store) + 1 + 0x7F
	} else {
		aux = uint8(*self.store + 1 + 0x7F)
	}
	return marshalInteger(uint64(aux), 1, w)
}

func (self Int8Encoder) Unmarshal(r io.Reader) (int, error) {
	value, n, err := unmarshalInteger(r, 1)
	if err != nil {
		return n, err
	}
	if value >= (1 + 0x7F) {
		*self.store = int8(value - 1 - 0x7F)
	} else {
		*self.store = int8(value) - 1 - 0x7F
	}
	return n, nil
}
