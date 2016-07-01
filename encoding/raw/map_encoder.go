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
import "reflect"

type mapEncoder struct {
	store       reflect.Value
	keyWorker      Encoder
	keyWorkerStore reflect.Value
	elemWorker      Encoder
	elemWorkerStore reflect.Value
}

func (self mapEncoder) Signature() string {
	//return "[]" + self.worker.Signature()
	return "not yet implemented"
}

func (self mapEncoder) Marshal(w io.Writer) (int, error) {
	/*
	var nc int
	storeVal := self.store.Elem()
	storeLen := storeVal.Len()
	n, err := marshalInteger(uint64(storeLen), 4, w)
	nc += n
	if err != nil {
		return nc, err
	}
	workerVal := self.workerStore.Elem()
	for i := 0; i < storeLen; i++ {
		workerVal.Set(storeVal.Index(i))
		n, err := self.worker.Marshal(w)
		nc += n
		if err != nil {
			return nc, err
		}
	}
	return nc, nil
	*/
	return 0, fmt.Errorf("not yet implemented")
}

func (self mapEncoder) Unmarshal(r io.Reader) (int, error) {
	/*
	var nc int
	v, n, err := unmarshalInteger(r, 4)
	nc += n
	if err != nil {
		return nc, err
	}
	storeLen := int(v)
	storeVal := reflect.MakeSlice(self.store.Elem().Type(), storeLen, storeLen)
	self.store.Elem().Set(storeVal)
	workerVal := self.workerStore.Elem()
	for i := 0; i < storeLen; i++ {
		n, err := self.worker.Unmarshal(r)
		nc += n
		if err != nil {
			return nc, err
		}
		storeVal.Index(i).Set(workerVal)
	}
	return nc, nil
	*/
	return 0, fmt.Errorf("not yet implemented")
}
