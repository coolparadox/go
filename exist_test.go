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
	var myData MyData
	myKeeper, err := keeper.New(&myData, "/tmp/keeper/my_data")
	if err != nil {
		panic(err)
	}
	myData = MyData{x: 22, y: 88}
	myKeeper.Save(0)
	myKeeper.Load(12)
	myKeeper.Erase(12)
	var exists bool = myKeeper.Exists(23)
	//err = myKeeper.Close()

	var myList []uint = keeper.List("/tmp/keeper/my_data")
	keeper.Wipe("/tmp/keeper/my_data")

}
