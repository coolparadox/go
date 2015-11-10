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
// along with Keep. If not, see <http://www.gnu.org/licenses/>.

package keep_test

import "github.com/coolparadox/go/storage/keep"
import "testing"
import "fmt"
import "os"
import "strconv"

type MyType struct {
	x, y uint32
}

var sample MyType = MyType{x: 55, y: 101}

func TestSaveLoad(t *testing.T) {

	n := ^uint(0)
	s := strconv.FormatUint(uint64(n), 36)
	t.Logf("n = %v = %s", n, s)

	m, _ := strconv.ParseUint(s, 36, 64)
	t.Logf("m = %s = %v", s, m)

	var err error
	myPath := "/tmp/my_data"
	err = os.MkdirAll(myPath, 0755)
	if err != nil {
		t.Fatal(fmt.Sprintf("cannot create directory '%s': %s", myPath, err))
	}
	var myData struct {
		MyType
		keep.Keep
	}
	myData.Keep = keep.NewOrPanic(&myData.MyType, myPath)
	myData.MyType = sample
	id, err := myData.Save(1)
	if err != nil {
		t.Fatal(fmt.Sprintf("Save failed: %s", err))
	}
	myData.MyType = MyType{}
	err = myData.Load(id)
	if err != nil {
		t.Fatal(fmt.Sprintf("Load failed: %s", err))
	}
	if myData.MyType != sample {
		t.Fatal("Save / Load value mismatch: saved %v loaded %v", sample, myData.MyType)
	}

}

func ExampleFormatPath() {
	fmt.Println(keep.FormatPath(0))
	fmt.Println(keep.FormatPath(1))
	fmt.Println(keep.FormatPath(^uint64(0)))
	// Output:
	// 0/0/0/0/0/0/0/0/0/0/0/0/0
	// 0/0/0/0/0/0/0/0/0/0/0/0/1
	// 3/w/5/e/1/1/2/6/4/s/g/s/f
}

func ExampleParsePath() {
	var id uint64
	var err error
	id, err = keep.ParsePath("0/0/0/0/0/0/0/0/0/0/0/0/0")
	if err == nil {
		fmt.Println(id)
	} else {
		fmt.Printf("%s\n", err)
	}
	id, err = keep.ParsePath("0/0/0/0/0/0/0/0/0/0/0/0/1")
	if err == nil {
		fmt.Println(id)
	} else {
		fmt.Printf("%s\n", err)
	}
	id, err = keep.ParsePath("3/w/5/e/1/1/2/6/4/s/g/s/f")
	if err == nil {
		fmt.Println(id)
	} else {
		fmt.Printf("%s\n", err)
	}
	// Output:
	// 0
	// 1
	// 18446744073709551615
}
