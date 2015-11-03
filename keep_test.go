package keep_test

import "github.com/coolparadox/keep"
import "testing"

func ExampleSayHello() {
	keep.SayHello()
	// Output: keep root '/tmp/keep' ok
}

type MyData struct {
	x, y int
}

func TestSave(t *testing.T) {
	var my1 MyData
	myKeep, err := keep.New(&my1, "/tmp/keep/my_data")
	if err != nil {
		panic(err)
	}
	my1 = MyData{x: 22, y: 88}
	my1_copy := my1
	myKeep.Save(1)
	my1 = MyData{}
	myKeep.Load(1)
	if my1 != my1_copy { t.FailNow() }

	//myKeep.Load(12)
	//myKeep.Erase(12)
	//var keeps bool = myKeep.Exists(23)

	//var myList []uint = keep.List("/tmp/keep/my_data")
	//keep.Wipe("/tmp/keep/my_data")

}
