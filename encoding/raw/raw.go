// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Raw, a binary encoder of Go types based on direct copy
// of memory content.
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

/*
Package raw implements binary serialization of Go types based on direct copy of
referenced memory.

*/
package raw

import "fmt"

type Encoder interface {
	TypeSpec() string
	Marshall() []byte
	Unmarshall([]byte) error
}

func New(data interface{}) (Encoder, error) {

	return nil, fmt.Errorf("not yet implemented")

}

