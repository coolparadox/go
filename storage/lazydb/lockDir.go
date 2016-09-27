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

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"syscall"
)

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
// The lock can be released by closing the returned file.
func lockDir(dir string, create bool, lockType int) (*os.File, error) {
	lockFile, err := openLockFile(dir, create)
	if err != nil {
		return nil, err
	}
	err = unix.Flock(int(lockFile.Fd()), lockType)
	if err != nil {
		lockFile.Close()
		return nil, err
	}
	return lockFile, nil
}

// lockDirForWrite places an advisory exclusive lock on a directory.
// If parameter create is false, directory must exist;
// otherwise it will be created.
// Returns an open file within the directory that holds the lock.
// The lock can be released by closing the returned file.
func lockDirForWrite(dir string, create bool) (*os.File, error) {
	return lockDir(dir, create, unix.LOCK_EX)
}

// lockDirForWriteNB is a nonblocking version of lockDirForWrite.
// A non nil returned file indicates success in achieving lock.
func lockDirForWriteNB(dir string, create bool) (*os.File, error) {
	file, err := lockDir(dir, create, unix.LOCK_NB|unix.LOCK_EX)
	errno, ok := err.(syscall.Errno)
	if ok && errno.Temporary() {
		err = nil
	}
	return file, err
}

// lockDirForRead places an advisory shared lock on an existent directory.
// Returns an open file within the directory that holds the lock.
// The lock can be released by closing the returned file.
func lockDirForRead(dir string) (*os.File, error) {
	return lockDir(dir, false, unix.LOCK_SH)
}
