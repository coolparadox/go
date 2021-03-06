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

// marshalInteger marshals an unsigned integer number of a given octet depth.
// Returns the number of bytes written.
func marshalInteger(value uint64, depth int, w io.Writer) (int64, error) {
	sequence := make([]byte, depth, depth)
	for i := 0; i < depth; i++ {
		sequence[i] = byte(value % 0x100)
		value /= 0x100
	}
	n, err := w.Write(sequence)
	return int64(n), err
}

// unmarshalInteger unmarshals an unsigned integer number of a given octet depth.
// Returns the unmarshaled value and the number of bytes read.
func unmarshalInteger(r io.Reader, depth int) (uint64, int64, error) {
	sequence := make([]byte, depth, depth)
	n, err := r.Read(sequence)
	if err != nil {
		return 0, int64(n), err
	}
	var answer uint64
	for i := 0; i < depth; i++ {
		answer *= 0x100
		answer += uint64(sequence[depth-1-i])
	}
	return answer, int64(n), nil
}
