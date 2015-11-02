package exist_test

import "github.com/coolparadox/exist"
import "testing"

func ExampleSayHello() {
	exist.SayHello()
	// Output: exist root '/tmp/exist' ok
}

type MyData struct {
	x, y int
}

func TestPersist(t *testing.T) {
	var data MyData
	exister, err := exist.MakeExister(&data, "/tmp/exist/data")
	if err != nil {
		panic(err)
	}
	data = MyData{x: 22, y: 88}
	exister.Persist(45)
}
