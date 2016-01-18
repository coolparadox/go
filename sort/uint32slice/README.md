# uint32slice
--
    import "github.com/coolparadox/go/sort/uint32slice"

Package uint32slice is an implementation of sort.Interface for slices of
uint32s.

## Usage

#### func  SearchUint32s

```go
func SearchUint32s(a []uint32, x uint32) int
```
SearchUint32s searches for x in a sorted slice of uint32s and returns the index
as specified by sort.Search. The return value is the index to insert x if x is
not present (it could be len(a)). The slice must be sorted in ascending order.

#### func  SortUint32s

```go
func SortUint32s(s []uint32)
```
SortUint32s sorts a slice of uint32 in increasing order.

#### type Uint32Slice

```go
type Uint32Slice []uint32
```

Uint32Slice attaches the methods of sort.Interface to []uint32, sorting in
increasing order.

#### func (Uint32Slice) Len

```go
func (r Uint32Slice) Len() int
```
Len is the number of elements in the collection.

#### func (Uint32Slice) Less

```go
func (r Uint32Slice) Less(i, j int) bool
```
Less reports whether the element with index i should sort before the element
with index j.

#### func (Uint32Slice) Search

```go
func (r Uint32Slice) Search(x uint32) int
```
Search returns the result of applying SearchUint32s to the receiver and x.

#### func (Uint32Slice) Sort

```go
func (r Uint32Slice) Sort()
```
Sort is a convenience method for applying sort.Sort to the receiver.

#### func (Uint32Slice) Swap

```go
func (r Uint32Slice) Swap(i, j int)
```
Swap swaps the elements with indexes i and j.
