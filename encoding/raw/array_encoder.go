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
	"strconv"
)

type arrayEncoder struct {
	worker      Encoder
	workerStore reflect.Value
	store       reflect.Value
}

func (e arrayEncoder) Signature() string {
	return "[" + strconv.Itoa(e.store.Elem().Len()) + "]" + e.worker.Signature()
}

func (e arrayEncoder) WriteTo(w io.Writer) (int64, error) {
	var nc int64
	storeVal := e.store.Elem()
	storeLen := storeVal.Len()
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

func (e arrayEncoder) ReadFrom(r io.Reader) (int64, error) {
	var nc int64
	storeVal := e.store.Elem()
	storeLen := storeVal.Len()
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
