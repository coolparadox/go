// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
// This file is part of LazyDB, a generic value storage library
// for the Go language.
//
// LazyDB is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LazyDB is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LazyDB. If not, see <http://www.gnu.org/licenses/>.

package lazydb

import "fmt"
import "os"
import "path"
import "golang.org/x/sys/unix"

func openLockFile(dir string, create bool) (*os.File, error) {
	if create {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return nil, fmt.Errorf("cannot create directory '%s': %s", dir, err)
		}
	}
	p := path.Join(dir, ".lock")
	f, err := os.OpenFile(p, os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// lockDir places an advisory lock on a directory.
// Returns an open file within the directory that holds the lock.
// The lock may be released by closing the returned file.
func lockDir(dir string, create bool, lockType int) (*os.File, error) {
	lockFile, err := openLockFile(dir, create)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %s", err)
	}
	err = unix.Flock(int(lockFile.Fd()), lockType)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("cannot place lock: %s", err)
	}
	_, err = os.Stat(dir)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("cannot stat: %s", err)
	}
	return lockFile, nil
}

// lockDirForWrite places an advisory exclusive lock on a directory.
// Returns an open file within the directory that holds the lock.
// The lock may be released by closing the returned file.
func lockDirForWrite(dir string, create bool) (*os.File, error) {
	return lockDir(dir, create, unix.LOCK_EX)
}

// lockDirForRead places an advisory shared lock on a directory.
// Returns an open file within the directory that holds the lock.
// The lock may be released by closing the returned file.
func lockDirForRead(dir string) (*os.File, error) {
	return lockDir(dir, false, unix.LOCK_SH)
}
