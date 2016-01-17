# runeslice
--
    import "github.com/coolparadox/go/sort/runeslice"

Package runeslice is an implementation of sort.Interface for slices of runes.

## Usage

#### func  SearchRunes

```go
func SearchRunes(a []rune, x rune) int
```
SearchRunes searches for x in a sorted slice of runes and returns the index as
specified by sort.Search. The return value is the index to insert x if x is not
present (it could be len(a)). The slice must be sorted in ascending order.

#### func  SortRunes

```go
func SortRunes(s []rune)
```
SortRunes sorts a slice of runes in increasing order.

#### type RuneSlice

```go
type RuneSlice []rune
```

RuneSlice attaches the methods of sort.Interface to []rune, sorting in
increasing order.

#### func (RuneSlice) Len

```go
func (r RuneSlice) Len() int
```
Len is the number of elements in the collection.

#### func (RuneSlice) Less

```go
func (r RuneSlice) Less(i, j int) bool
```
Less reports whether the element with index i should sort before the element
with index j.

#### func (RuneSlice) Search

```go
func (r RuneSlice) Search(x rune) int
```
Search returns the result of applying SearchRunes to the receiver and x.

#### func (RuneSlice) Sort

```go
func (r RuneSlice) Sort()
```
Sort is a convenience method for applying sort.Sort to the receiver.

#### func (RuneSlice) Swap

```go
func (r RuneSlice) Swap(i, j int)
```
Swap swaps the elements with indexes i and j.
