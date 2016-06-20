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

type Uint16Encoder struct{ store *uint16 }

func (Uint16Encoder) Signature() string {
	return "uint16"
}

func (self Uint16Encoder) Marshal(w io.Writer) (int, error) {
	return marshalInteger(uint64(*self.store), 2, w)
}

func (self Uint16Encoder) Unmarshal(r io.Reader) (int, error) {
	value, n, err := unmarshalInteger(r, 2)
	if err != nil {
		return n, err
	}
	*self.store = uint16(value)
	return n, nil
}
