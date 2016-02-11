# concur
--
    import "github.com/coolparadox/go/storage/concur"

Package concur is a storage of byte sequences for Go with automatic key
generation.


### Basics

Use New to create or open a collection of key/value pairs in the filesystem. The
collection can then be managed by methods of the collection handler.

    db, _ := concur.New("/path/to/my/collection", 0)
    key, _ := db.Save(byte[]{1,3,5,7,9}) // store data in a new key
    val, _ := db.Load(key) // retrieve value of a key
    db.SaveAs(key, byte[]{0,2,4,6,8}) // update existent key
    db.Erase(key) // remove a key


### Issues

Keys are 32 bit unsigned integers. Values are byte sequences of arbitrary
length.

Apart from other storage implementations that map a single file as the database,
this package takes an experimental approach where keys are managed using
filesystem subdirectories (see Key Management below). Therefore the filesystem
chosen for storage is the real engine that maps keys to values, and their
designers are the ones who must take credit if this package happens to achieve
satisfactory performance.

Although concur write methods commit changes to filesystem immediately on
successful return, the operating system can make use of memory buffers for
increasing performance of filesystem access. Users may need to manually flush
updates to disk (eg sync, umount) to guarantee that all updates to the
collection are written to disk.

Wipe method can take a long time to return.


### Key Management

(This is an explanation of how 32 bit keys are internally mapped to values by
the implementation. You don't really need to know it for using concur; feel free
to skip this section.)

Each key is uniquely associated with a distinct file in the filesystem. The path
to the file is derived from the key, eg. a key of 0x12345678, assuming the
numeric base of key components is set to 16, is the file 1/2/3/4/5/6/7/8 under
the database directory. The value associated with the key is the content of the
file. Conversely, keys in the database are retrieved by parsing the path of
existent files.

When creating a new database, user may choose the numeric base of key
components. This value ultimately defines how many directories are allowed to
exist in each subdirectory level towards reaching associated files. The base can
range from MinBase (2, resulting in a level depth of 32 for holding a 32 bit
key) to MaxBase (0x10000, giving a level depth of only 2).

Whether the numeric base chosen, directories and files are named by single
unicode characters, where the first 10 ones in the mapping range are decimal
digits from 0 to 9, and the next 26 ones are upper case letters from A to Z.
Thus component bases up to 36 are guaranteed to be mapped by characters in the
ascii range.

It's worth noting that all this key composition stuff happens transparently to
the user. Poking around the directory of a concur collection, despite it's cool
for the sake of curiosity, is not required for making use of this package.


### Wish List

Document filesystem guidelines for better performance with package concur.

## Usage

```go
const (
	MinBase = 2
	MaxBase = 0x10000
)
```
MinBase and MaxBase define the range of possible values for the numeric base of
key components in the filesystem (see parameter base in New).

```go
const (
	Depth2Base  = 0x10000
	Depth4Base  = 0x100
	Depth8Base  = 0x10
	Depth16Base = 0x4
	Depth32Base = 0x2
)
```
Depth*Base are convenience values of numeric bases of key components to be used
when creating a new database. These values give the most efficient occupation of
subdirectories in the filesystem (see Key Management).

```go
const MaxKey = 0xFFFFFFFF
```
MaxKey represents the maximum value of a key.

#### func  IsKeyNotFoundError

```go
func IsKeyNotFoundError(e error) bool
```
IsKeyNotFoundError tells if an error is of type KeyNotFoundError.

#### func  Wipe

```go
func Wipe(dir string) error
```
Wipe removes a collection from the filesystem.

On success, all content of the given directory is cleaned. The directory itself
is not removed.

Existence of a concur collection in the directory is verified prior to cleaning
it.

#### type Concur

```go
type Concur struct {
}
```

Concur handles a collection of byte sequences stored in a directory of the
filesystem.

#### func  New

```go
func New(dir string, base uint32) (Concur, error)
```
New creates a Concur value.

Parameter dir is an absolute path to a directory in the filesystem for storing
the collection. If it's the first time this directory is used by package concur,
it must be empty.

Parameter base is the numeric base of key components for naming files and
subdirectories under the collection (see Key Management for details). It has
effect only during creation of a collection. Pass zero for a sane default.

#### func (Concur) Erase

```go
func (r Concur) Erase(key uint32) error
```
Erase erases a key.

#### func (Concur) Exists

```go
func (r Concur) Exists(key uint32) (bool, error)
```
Exists verifies if a key exists.

#### func (Concur) FindKey

```go
func (r Concur) FindKey(key uint32, ascending bool) (uint32, error)
```
FindKey takes a key and returns it if it exists. If key does not exist, the
closest key in ascending (or descending) order is returned instead.

A KeyNotFoundError is returned if there are no keys to be answered.

#### func (Concur) Load

```go
func (r Concur) Load(key uint32) ([]byte, error)
```
Load retrieves the value associated with a key.

#### func (Concur) Save

```go
func (r Concur) Save(value []byte) (uint32, error)
```
Save creates a key with a new value. The key is automatically assigned and
guaranteed to be new.

Returns the assigned key.

#### func (Concur) SaveAs

```go
func (r Concur) SaveAs(key uint32, value []byte) error
```
SaveAs creates (or updates) a key with a new value.

#### type KeyNotFoundError

```go
type KeyNotFoundError struct{}
```

KeyNotFoundError is returned by SmallestKeyNotLessThan and
LargestKeyNotGreaterThan when there are no keys available.

#### func (KeyNotFoundError) Error

```go
func (KeyNotFoundError) Error() string
```
