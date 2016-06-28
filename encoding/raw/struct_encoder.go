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

type StructEncoder struct{ store []Encoder }

func (self StructEncoder) Signature() string {
	ans := "struct {"
	for i := 0; i < len(self.store); i++ {
		if i > 0 {
			ans += ";"
		}
		ans += " " + self.store[i].Signature()
	}
	ans += " }"
	return ans
}

func (self StructEncoder) Marshal(w io.Writer) (int, error) {
	var count int
	for i := 0; i < len(self.store); i++ {
		n, err := self.store[i].Marshal(w)
		count += n
		if err != nil {
			return count, err
		}
	}
	return count, nil
}

func (self StructEncoder) Unmarshal(r io.Reader) (int, error) {
	var count int
	for i := 0; i < len(self.store); i++ {
		n, err := self.store[i].Unmarshal(r)
		count += n
		if err != nil {
			return count, err
		}
	}
	return count, nil
}
