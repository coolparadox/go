// Copyright 2016 Rafael Lorandi <coolparadox@gmail.com>
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

import "fmt"
import "os"
import "path"
import "golang.org/x/sys/unix"

func openLockFile(dir string) (*os.File, error) {
	p := path.Join(dir, ".lock")
	f, err := os.OpenFile(p, os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// lockDirForWrite places an advisory exclusive lock on a directory.
// Returns an open file within the directory that holds the lock.
// The lock may be released by closing the returned file.
func lockDirForWrite(dir string) (*os.File, error) {
	lockFile, err := openLockFile(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %s", err)
	}
	err = unix.Flock(int(lockFile.Fd()), unix.LOCK_EX)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("cannot place exclusive lock: %s", err)
	}
	return lockFile, nil
}

// lockDirForRead places an advisory shared lock on a directory.
// Returns an open file within the directory that holds the lock.
// The lock may be released by closing the returned file.
func lockDirForRead(dir string) (*os.File, error) {
	lockFile, err := openLockFile(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %s", err)
	}
	err = unix.Flock(int(lockFile.Fd()), unix.LOCK_SH)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("cannot place shared lock: %s", err)
	}
	return lockFile, nil
}
