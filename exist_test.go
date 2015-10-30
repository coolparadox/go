package exist_test

import "github.com/coolparadox/exist"
import "testing"

func ExampleSayHello() {
	exist.SayHello()
	// Output: exist root '/tmp/exist' ok
}

type MyData struct {
	exist.Exister
	X, Y uint
}

func (self *MyData) Persist(oid uint) (uint, error) {
	if self.Exister.Persist == nil {
		self.Exister.Persist = self.Exister.MakePersist(self, "datastore")
	}
	return self.Exister.Persist(self, oid)
}

func TestPersist(t *testing.T) {
	var data MyData
	data = MyData{X: 22, Y: 2}
	data.Persist(4)
	data.X = 28
	data.Persist(6)
	data = MyData{X: 123, Y: 445}
	data.Persist(8)
	var data2 MyData
	data2 = MyData{X: 873, Y:7765}
	data2.Persist(9)
}
