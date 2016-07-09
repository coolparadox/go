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
import "reflect"

type ptrEncoder struct {
	worker      Encoder
	workerStore reflect.Value
	store       reflect.Value
}

func (self ptrEncoder) Signature() string {
	return "*" + self.worker.Signature()
}

func (self ptrEncoder) Marshal(w io.Writer) (int, error) {
	var nc int
	storeVal := self.store.Elem()
	if reflect.DeepEqual(storeVal.Interface(), reflect.Zero(storeVal.Type()).Interface()) {
		n, err := marshalInteger(0x00, 1, w)
		nc += n
		return nc, err
	}
	n, err := marshalInteger(0xFF, 1, w)
	nc += n
	if err != nil {
		return nc, nil
	}
	workerVal := self.workerStore.Elem()
	workerVal.Set(storeVal.Elem())
	n, err = self.worker.Marshal(w)
	nc += n
	return nc, err
}

func (self ptrEncoder) Unmarshal(r io.Reader) (int, error) {
	var nc int
	v, n, err := unmarshalInteger(r, 1)
	nc += n
	if err != nil {
		return n, err
	}
	storeVal := self.store.Elem()
	if v == 0 {
		storeVal.Set(reflect.Zero(storeVal.Type()))
		return nc, nil
	}
	storeVal.Set(reflect.New(storeVal.Type().Elem()))
	n, err = self.worker.Unmarshal(r)
	nc += n
	if err != nil {
		return nc, err
	}
	storeVal.Elem().Set(self.workerStore.Elem())
	return nc, nil
}
