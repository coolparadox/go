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

import (
	"io"
	"reflect"
)

type mapEncoder struct {
	store           reflect.Value
	keyWorker       Encoder
	keyWorkerStore  reflect.Value
	elemWorker      Encoder
	elemWorkerStore reflect.Value
}

func (e mapEncoder) Signature() string {
	return "map[" + e.keyWorker.Signature() + "]" + e.elemWorker.Signature()
}

func (e mapEncoder) WriteTo(w io.Writer) (int64, error) {
	var nc int64
	storeVal := e.store.Elem()
	storeLen := storeVal.Len()
	n, err := marshalInteger(uint64(storeLen), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	keyWorkerVal := e.keyWorkerStore.Elem()
	elemWorkerVal := e.elemWorkerStore.Elem()
	keys := storeVal.MapKeys()
	for _, keyVal := range keys {
		keyWorkerVal.Set(keyVal)
		n, err := e.keyWorker.WriteTo(w)
		nc += n
		if err != nil {
			return nc, err
		}
		elemWorkerVal.Set(storeVal.MapIndex(keyVal))
		n, err = e.elemWorker.WriteTo(w)
		nc += n
		if err != nil {
			return nc, err
		}
	}
	return nc, nil
}

func (e mapEncoder) ReadFrom(r io.Reader) (int64, error) {
	var nc int64
	v, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	storeLen := int(v)
	storeVal := reflect.MakeMap(e.store.Elem().Type())
	e.store.Elem().Set(storeVal)
	keyWorkerVal := e.keyWorkerStore.Elem()
	elemWorkerVal := e.elemWorkerStore.Elem()
	for i := 0; i < storeLen; i++ {
		n, err := e.keyWorker.ReadFrom(r)
		nc += n
		if err != nil {
			return nc, err
		}
		n, err = e.elemWorker.ReadFrom(r)
		nc += n
		if err != nil {
			return nc, err
		}
		storeVal.SetMapIndex(keyWorkerVal, elemWorkerVal)
	}
	return nc, nil
}
