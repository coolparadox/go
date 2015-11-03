package keep

import "fmt"
import "os"
import "path"
import "log"
import "io"
import "reflect"
import "errors"
import "unsafe"

var rootDirPath string

func isDirectoryEmpty(name string) (bool, error) {

	dir, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer dir.Close()
	_, err = dir.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err

}

func doesFileExist(name string) (bool, error) {

	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func init() {

	// Make sure keep root path is sane
	var ok bool
	rootDirPath, ok = os.LookupEnv("KEEPROOT")
	if !ok {
		panic("undefined KEEPROOT environment variable")
	}
	log.Printf("KEEPROOT is '%s'", rootDirPath)
	if !path.IsAbs(rootDirPath) {
		panic(fmt.Sprintf("root dir '%s' is not an absolute path", rootDirPath))
	}
	rootDirPath = path.Clean(rootDirPath)
	finfo, err := os.Stat(rootDirPath)
	if err != nil {
		panic(fmt.Sprintf("root dir '%s' unreachable: %s", rootDirPath, err))
	}
	if !finfo.IsDir() {
		panic(fmt.Sprintf("root dir '%s' is not a directory", rootDirPath))
	}
	if finfo.Mode().Perm()&0022 != 0 {
		panic(fmt.Sprintf("root dir '%s' is group or world writable", rootDirPath))
	}
	rootFilePath := path.Join(rootDirPath, ".keepRoot")
	rootFileExists, err := doesFileExist(rootFilePath)
	if err != nil {
		panic(fmt.Sprintf("cannot check for '%s' existence: %s", rootFilePath, err))
	}
	rootDirEmpty, err := isDirectoryEmpty(rootDirPath)
	if err != nil {
		panic(fmt.Sprintf("cannot check if root dir '%s' is empty: %s", rootDirPath, err))
	}
	if !rootFileExists && !rootDirEmpty {
		panic(fmt.Sprintf("root dir '%s' is not empty and is not a keep database", rootDirPath))
	}
	for dir := path.Dir(rootDirPath); ; dir = path.Dir(dir) {
		if ok, _ = doesFileExist(path.Join(dir, ".keepRoot")); ok {
			panic(fmt.Sprintf("keep database detected in '%s' above root dir '%s'", dir, rootDirPath))
		}
		if dir == "/" {
			break
		}
	}
	if rootDirEmpty {
		rootFile, err := os.Create(rootFilePath)
		if err != nil {
			panic(fmt.Sprintf("cannot create database root file '%s': %s", rootFilePath, err))
		}
		err = rootFile.Chmod(0644)
		if err != nil {
			panic(fmt.Sprintf("cannot chmod root file '%s': %s", rootFilePath, err))
		}
		log.Printf("keep database initialized in '%s'", rootDirPath)
	}

}

func SayHello() {
	fmt.Printf("keep root '%s' ok\n", rootDirPath)
}

type Keep struct {
	addr unsafe.Pointer
	typ  reflect.Type
	home string
}

func (self Keep) Save(oid uint) (uint, error) {
	v := reflect.NewAt(self.typ, self.addr)
	data := v.Elem().Interface()
	fmt.Printf("%s save oid %v in %v: %v\n", self.typ, oid, self.home, data)
	return oid, nil
}

func (self Keep) Load(oid uint) error {
	v := reflect.NewAt(self.typ, self.addr)
	data := v.Elem().Interface()
	fmt.Printf("%s load oid %v in %v: %v\n", self.typ, oid, self.home, data)
	return nil
}

func New(storage interface{}, home string) (Keep, error) {
	v := reflect.ValueOf(storage)
	if v.Kind() != reflect.Ptr {
		return Keep{}, errors.New("keep.New(): storage parameter is not a pointer")
	}
	v = v.Elem()
	p := unsafe.Pointer(v.UnsafeAddr())
	t := v.Type()
	return Keep{home: home, addr: p, typ: t}, nil
}
