// Copyright 2015 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of Concur, a generic value storage library
// for the Go language.
//
// Concur is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Concur is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Concur. If not, see <http://www.gnu.org/licenses/>.

package concur

import "path"
import "errors"
import "fmt"
import "os"
import "io"
import "log"
import "io/ioutil"
import "strings"
import "strconv"

// Concur handles a collection of persisted byte sequences.
type Concur struct {
	dir string
}

// concurLabel is the file checked for existance of a concur database in a
// directory.
const concurLabel string = ".concur"

// New creates a Concur value.
//
// The dir parameter is an absolute path to a directory in the filesystem
// for storing the collection. If it's the first time this directory is used by
// package concur, it must be empty.
func New(dir string) (Concur, error) {
	if !path.IsAbs(dir) {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not absolute", dir))
	}
	dir = path.Clean(dir)
	finfo, err := os.Stat(dir)
	if err != nil {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is unreachable: %s", dir, err))
	}
	if !finfo.IsDir() {
		return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not a directory", dir))
	}
	concurFile := path.Join(dir, concurLabel)
	concurFileExists := true
	finfo, err = os.Stat(concurFile)
	if err != nil {
		if os.IsNotExist(err) {
			concurFileExists = false
		} else {
			return Concur{}, errors.New(fmt.Sprintf("cannot check for '%s' existence: %s", concurFile, err))
		}
	}
	if concurFileExists {
		if finfo.IsDir() {
			return Concur{}, errors.New(fmt.Sprintf("concur db mark file '%s' is a directory", concurFile))
		}
	} else {
		file, err := os.Open(dir)
		if err != nil {
			return Concur{}, errors.New(fmt.Sprintf("cannot open '%s': %s", dir, err))
		}
		defer file.Close()
		_, err = file.Readdir(1)
		if err != io.EOF {
			return Concur{}, errors.New(fmt.Sprintf("dir '%s' is not empty and is not a concur db", dir))
		}
		_, err = os.Create(concurFile)
		if err != nil {
			return Concur{}, errors.New(fmt.Sprintf("cannot create concur db mark file '%s'", concurFile))
		}
		log.Printf("concur database initialized in '%s'", dir)
	}
	return Concur{
		dir: dir,
	}, nil
}

// SaveAs stores a byte sequence identified by an id.
func (c Concur) SaveAs(data []byte, id uint64) error {
	var err error
	targetPath := path.Join(c.dir, formatPath(id))
	targetDir := path.Dir(targetPath)
	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot create directory '%s': %s", targetDir, err))
	}
	err = ioutil.WriteFile(targetPath, data, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot write file '%s': %s", targetPath, err))
	}
	return nil
}

// Load retrieves a previously saved byte sequence.
func (c Concur) Load(id uint64) ([]byte, error) {
	sourcePath := path.Join(c.dir, formatPath(id))
	buf, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read file '%s': %s", sourcePath, err))
	}
	return buf, nil
}

// Erase erases an item from the collection.
func (c Concur) Erase(id uint64) error {
	return errors.New("not yet implemented")
}

// Exists verifies if an item exists in the collection.
func (c Concur) Exists(id uint64) (bool, error) {
	return false, errors.New("not yet implemented")
}

// Wipe removes a collection from the filesystem.
func Wipe(dir string) error {
	return errors.New("not yet implemented")
}

// formatPath converts an id to a relative filesystem path.
func formatPath(id uint64) string {
	return strings.Join(
		strings.Split(
			fmt.Sprintf(
				"%013s",
				strconv.FormatUint(
					uint64(id),
					36)),
			""),
		string(os.PathSeparator))
}

// parsePath converts a relative filesystem path to an id.
func parsePath(path string) (uint64, error) {
	id, err := strconv.ParseUint(
		strings.Join(
			strings.Split(
				path,
				string(os.PathSeparator)),
			""),
		36,
		64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("cannot convert '%s' to id: %s", path, err))
	}
	return id, nil
}
