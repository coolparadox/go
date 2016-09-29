// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Keep, a storage library of typed data for the Go
// language.
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

/*
Package keep is a storage library of typed data for Go.

Basics

	type MyType struct {
		name string
		age uint
		...
	}

	var myData struct {
		keep.Keep
		MyType
	}

	var err error
	myData.Keep, err = keep.New(&myData.MyType, "/tmp/my_data")

	myData.name = "Agent Smith"
	myData.age = "101"
	k1, err := myData.Save()
	...
	myData.Load(k1)
	...
	myData.Erase(k1)

*/
package keep

import (
	"bytes"
	"fmt"
	"github.com/coolparadox/go/encoding/raw"
	"github.com/coolparadox/go/storage/lazydb"
	"io"
)

const keepLabel = "Keep"

// Keep is a handler to a collection of typed Go data stored in the filesystem.
type Keep struct {
	encoder raw.Encoder
	db      lazydb.LazyDB
}

// New creates a new Keep collection,
// or opens an existent one.
//
// Parameter placeholder must be a pointer to a variable of any supported type
// (see Supported Types).
// The resulting Keep collection uses this variable as a placeholder of typed data.
//
// Parameter dir is an absolute path to a directory in the filesystem
// for storing the collection.
// If it's the first time this directory is used by package keep,
// it must be empty.
//
// Returns a Keep handler.
func New(placeholder interface{}, dir string) (Keep, error) {
	encoder, err := raw.New(placeholder)
	if err != nil {
		return Keep{}, fmt.Errorf("failed to initialize encoder: %s", err)
	}
	db, err := lazydb.New(dir, 0)
	if err != nil {
		return Keep{}, fmt.Errorf("failed to initialize database: %s", err)
	}
	_, err = db.FindKey(0, true)
	if err == lazydb.KeyNotFoundError {
		// Initialize empty database
		_, err = db.SaveAs(0, []io.Reader{
			bytes.NewReader([]byte(keepLabel)),
			bytes.NewReader([]byte(encoder.Signature())),
		})
		if err != nil {
			return Keep{}, fmt.Errorf("failed to initialize database: %s", err)
		}
	} else if err != nil {
		return Keep{}, fmt.Errorf("failed to query database: %s", err)
	}
	ok, err := db.Exists(0, 0)
	if err != nil {
		return Keep{}, fmt.Errorf("failed to query database: %s", err)
	}
	if !ok {
		return Keep{}, fmt.Errorf("not a Keep database")
	}
	ok, err = db.Exists(0, 1)
	if err != nil {
		return Keep{}, fmt.Errorf("failed to query database: %s", err)
	}
	if !ok {
		return Keep{}, fmt.Errorf("not a Keep database")
	}
	dbKeepLabel := new(bytes.Buffer)
	dbSignature := new(bytes.Buffer)
	_, err = db.Load(0, []io.Writer{dbKeepLabel, dbSignature})
	if err != nil {
		return Keep{}, fmt.Errorf("failed to query database: %s", err)
	}
	if string(dbKeepLabel.Bytes()) != keepLabel {
		return Keep{}, fmt.Errorf("not a Keep database")
	}
	if string(dbSignature.Bytes()) != encoder.Signature() {
		return Keep{}, fmt.Errorf("type signature mismatch: expected '%s', found in database '%s'", encoder.Signature, string(dbSignature.Bytes()))
	}
	return Keep{
		encoder: encoder,
		db:      db,
	}, nil
}

