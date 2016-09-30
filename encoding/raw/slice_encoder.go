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

type sliceEncoder struct {
	store       reflect.Value
	worker      Encoder
	workerStore reflect.Value
}

func (e sliceEncoder) Signature() string {
	return "[]" + e.worker.Signature()
}

func (e sliceEncoder) WriteTo(w io.Writer) (int64, error) {
	var nc int64
	storeVal := e.store.Elem()
	storeLen := storeVal.Len()
	n, err := marshalInteger(uint64(storeLen), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	workerVal := e.workerStore.Elem()
	for i := 0; i < storeLen; i++ {
		workerVal.Set(storeVal.Index(i))
		n, err := e.worker.WriteTo(w)
		nc += n
		if err != nil {
			return nc, err
		}
	}
	return nc, nil
}

func (e sliceEncoder) ReadFrom(r io.Reader) (int64, error) {
	var nc int64
	v, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	storeLen := int(v)
	storeVal := reflect.MakeSlice(e.store.Elem().Type(), storeLen, storeLen)
	e.store.Elem().Set(storeVal)
	workerVal := e.workerStore.Elem()
	for i := 0; i < storeLen; i++ {
		n, err := e.worker.ReadFrom(r)
		nc += n
		if err != nil {
			return nc, err
		}
		storeVal.Index(i).Set(workerVal)
	}
	return nc, nil
}
