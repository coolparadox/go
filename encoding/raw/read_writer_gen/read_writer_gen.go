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

package main

import (
	"fmt"
	"io"
	"os"
)

var encoders = []string{
	"uint8Encoder",
	"uint16Encoder",
	"uint32Encoder",
	"uint64Encoder",
	"int8Encoder",
	"int16Encoder",
	"int32Encoder",
	"int64Encoder",
	"float32Encoder",
	"float64Encoder",
	"complex64Encoder",
	"complex128Encoder",
	"boolEncoder",
	"stringEncoder",
	"arrayEncoder",
	"sliceEncoder",
	"structEncoder",
	"ptrEncoder",
	"mapEncoder",
}

// main generates read_writer.go
func main() {
	target, err := os.OpenFile("read_writer.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	prefix, err := os.Open("read_writer_gen/read_writer_gen.prefix")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(target, prefix)
	if err != nil {
		panic(err)
	}
	for _, encoder := range encoders {
		gen_read(target, encoder)
		gen_write(target, encoder)
	}
}

func gen_read(t io.Writer, e string) {
	fmt.Fprintf(t, "\nfunc (e %s) Read(p []byte) (int, error) {\n", e)
	fmt.Fprintf(t, "\treturn readEncoder(e, p)\n")
	fmt.Fprintf(t, "}\n")
}

func gen_write(t io.Writer, e string) {
	fmt.Fprintf(t, "\nfunc (e %s) Write(p []byte) (int, error) {\n", e)
	fmt.Fprintf(t, "\treturn writeEncoder(e, p)\n")
	fmt.Fprintf(t, "}\n")
}
