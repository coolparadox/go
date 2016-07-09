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

package raw

import "math"
import "io"

type complex128Encoder struct{ store *complex128 }

func (complex128Encoder) Signature() string {
	return "complex128"
}

func (self complex128Encoder) Marshal(w io.Writer) (int, error) {
	var nc int
	n, err := marshalInteger(math.Float64bits(real(*self.store)), 8, w)
	nc += n
	if err != nil {
		return nc, err
	}
	n, err = marshalInteger(math.Float64bits(imag(*self.store)), 8, w)
	nc += n
	return nc, err
}

func (self complex128Encoder) Unmarshal(r io.Reader) (int, error) {
	var nc int
	vr, n, err := unmarshalInteger(r, 8)
	nc += n
	if err != nil {
		return nc, err
	}
	vi, n, err := unmarshalInteger(r, 8)
	nc += n
	if err != nil {
		return nc, err
	}
	*self.store = complex(math.Float64frombits(vr), math.Float64frombits(vi))
	return nc, nil
}
