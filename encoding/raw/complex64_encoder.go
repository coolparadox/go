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

type complex64Encoder struct{ store *complex64 }

func (complex64Encoder) Signature() string {
	return "complex64"
}

func (e complex64Encoder) WriteTo(w io.Writer) (int64, error) {
	var nc int64
	n, err := marshalInteger(uint64(math.Float32bits(real(*e.store))), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	n, err = marshalInteger(uint64(math.Float32bits(imag(*e.store))), 4, w)
	nc += n
	return nc, err
}

func (e complex64Encoder) ReadFrom(r io.Reader) (int64, error) {
	var nc int64
	vr, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	vi, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	*e.store = complex(math.Float32frombits(uint32(vr)), math.Float32frombits(uint32(vi)))
	return nc, nil
}
