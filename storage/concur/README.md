# concur
--
    import "github.com/coolparadox/go/storage/concur"

Package concur is a storage of byte sequences for Go with automatic key
generation.


### Basics

Use New to create or open a collection of key/value pairs in the filesystem. The
collection can then be managed by methods of the collection handler.

    db, _ := concur.New("/path/to/my/collection")
    key, _ := db.Save(byte[]{1,3,5,7,9}) // store data in a new key
    val, _ := db.Load(key) // retrieve value of a key
    db.Put(key, byte[]{0,2,4,6,8}) // update existent key
    db.Erase(key) // remove a key


### Issues

Keys are 32 bit unsigned integers. Values are byte sequences of arbitrary
length.

Apart from other storage implementations that map a single file as the database,
this package takes an experimental approach where keys are managed using
filesystem subdirectories. Therefore the filesystem chosen for storage is the
real engine that maps keys to values, and their designers are the ones who must
take credit if this package happens to achieve satisfactory performance.

Although concur write methods commit changes to filesystem immediately on
successful return, the operating system can make use of memory buffers for
increasing performance of filesystem access. Users may need to manually flush
updates to disk (eg sync, umount) to guarantee that all updates to the
collection are written to disk.

Wipe method can take a long time to return.


### Bugs

Concurrent access to a collection is not yet thought of, and can be a fruitful
source of weirdness.


### Wish List

Protect against concurrent access to collections.

Document filesystem guidelines for better performance with package concur.

## Usage

```go
const KeyMax = 0xFFFFFFFF
```
KeyMax is the maximum value of a key.

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
func New(dir string) (Concur, error)
```
New creates a Concur value.

The dir parameter is an absolute path to a directory in the filesystem for
storing the collection. If it's the first time this directory is used by package
concur, it must be empty.

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

#### func (Concur) Get

```go
func (r Concur) Get(key uint32) ([]byte, error)
```
Get retrieves the value associated with a key.

#### func (Concur) Load

```go
func (r Concur) Load(key uint32) ([]byte, error)
```
Load is a synonym for Get.

#### func (Concur) Put

```go
func (r Concur) Put(key uint32, value []byte) error
```
Put creates (or updates) a key with a new value.

#### func (Concur) Save

```go
func (r Concur) Save(value []byte) (uint32, error)
```
Save creates a key with a new value. The key is automatically assigned and
guaranteed to be new.

Returns the assigned key.

#### func (Concur) SmallestKeyNotLessThan

```go
func (r Concur) SmallestKeyNotLessThan(key uint32) (uint32, bool, error)
```
SmallestKeyNotLessThan takes a key and returns it if it exists. If key does not
exist, the closest key in ascending order is returned instead.

The bool return value tells if a key was found to be answered.
