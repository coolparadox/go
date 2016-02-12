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

// KeyNotFoundError is returned by FindKey when there are no keys available.
type KeyNotFoundError struct{}

func (KeyNotFoundError) Error() string {
	return "key not found"
}

// IsKeyNotFoundError tells if an error is of type KeyNotFoundError.
func IsKeyNotFoundError(e error) bool {
	_, answer := e.(KeyNotFoundError)
	return answer
}
