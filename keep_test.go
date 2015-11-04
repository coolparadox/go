// Copyright 2015 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Keep, a persistency library for the Go language.
//
// Keep is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Keep is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Keep.  If not, see <http://www.gnu.org/licenses/>.

package keep_test

import "github.com/coolparadox/keep"
import "testing"

func ExampleSayHello() {
	keep.SayHello()
	// Output: keep root '/tmp/keep' ok
}

type MyData struct {
	x, y int
}

func TestSaveLoad(t *testing.T) {

	var err error

	var data1 struct {
		keep.Keep
		MyData
	}
	data1.Keep, err = keep.New(&data1.MyData, "here1")
	if err != nil {
		panic(err)
	}
	data1.MyData = MyData{x: 39}
	data1.y = 101
	data1.Save(3)

	var my1 MyData
	my1k, err := keep.New(&my1, "/tmp/keep/my_data")
	if err != nil {
		panic(err)
	}
	my1 = MyData{x: 22, y: 88}
	my2 := my1
	my1k.Save(1)
	my1 = MyData{}
	my1k.Load(1)
	if my1 != my2 {
		t.FailNow()
	}

	//my1k.Load(12)
	//my1k.Erase(12)
	//var keeps bool = my1k.Exists(23)

	//var myList []uint = keep.List("/tmp/keep/my_data")
	//keep.Wipe("/tmp/keep/my_data")

}
