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

type float64Encoder struct{ store *float64 }

func (float64Encoder) Signature() string {
	return "float64"
}

func (e float64Encoder) WriteTo(w io.Writer) (int64, error) {
	return marshalInteger(math.Float64bits(*e.store), 8, w)
}

func (e float64Encoder) ReadFrom(r io.Reader) (int64, error) {
	value, n, err := unmarshalInteger(r, 8)
	if err != nil {
		return n, err
	}
	*e.store = math.Float64frombits(value)
	return n, nil
}
