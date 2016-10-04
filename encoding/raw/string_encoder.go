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

import "io"

type stringEncoder struct{ store *string }

func (stringEncoder) Signature() string {
	return "string"
}

func (e stringEncoder) WriteTo(w io.Writer) (int64, error) {
	var nc int64
	storeLen := len(*e.store)
	n, err := marshalInteger(uint64(storeLen), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	for i := 0; i < storeLen; i++ {
		n, err := marshalInteger(uint64((*e.store)[i]), 1, w)
		nc += n
		if err != nil {
			return nc, err
		}
	}
	return nc, nil
}

func (e stringEncoder) ReadFrom(r io.Reader) (int64, error) {
	var nc int64
	v, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	storeLen := int(v)
	answer := make([]byte, storeLen, storeLen)
	for i := 0; i < storeLen; i++ {
		v, n, err := unmarshalInteger(r, 1)
		nc += n
		if err != nil {
			return nc, err
		}
		answer[i] = uint8(v)
	}
	*e.store = string(answer)
	return nc, nil
}
