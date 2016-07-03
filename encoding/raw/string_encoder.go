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

import "fmt"
import "io"

type stringEncoder struct{ store *string }

func (stringEncoder) Signature() string {
	return "string"
}

func (self stringEncoder) Marshal(w io.Writer) (int, error) {
	var nc int
	storeLen := len(*self.store)
	n, err := marshalInteger(uint64(storeLen), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	for i := 0; i < storeLen; i++ {
		n, err := marshalInteger(uint64((*self.store)[i]), 1, w)
		nc += n
		if err != nil {
			return nc, err
		}
	}
	return nc, nil
}

func (self stringEncoder) Unmarshal(r io.Reader) (int, error) {
	return 0, fmt.Errorf("not yet implemented")
}