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

type Int16Encoder struct{ store *int16 }

func (Int16Encoder) Signature() string {
	return "int16"
}

func (self Int16Encoder) Marshal(w io.Writer) (int, error) {
	var aux uint16
	if *self.store >= 0 {
		aux = uint16(*self.store) + 1 + 0x7FFF
	} else {
		aux = uint16(*self.store + 1 + 0x7FFF)
	}
	return marshalInteger(uint64(aux), 2, w)
}

func (self Int16Encoder) Unmarshal(r io.Reader) (int, error) {
	value, n, err := unmarshalInteger(r, 2)
	if err != nil {
		return n, err
	}
	if value >= (1 + 0x7FFF) {
		*self.store = int16(value - 1 - 0x7FFF)
	} else {
		*self.store = int16(value) - 1 - 0x7FFF
	}
	return n, nil
}
