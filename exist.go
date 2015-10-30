package exist

import "fmt"
import "os"
import "path"
import "log"
import "io"

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

	// Make sure exist root path is sane
	var ok bool
	rootDirPath, ok = os.LookupEnv("EXISTROOT")
	if !ok {
		panic("undefined EXISTROOT environment variable")
	}
	log.Printf("EXISTROOT is '%s'", rootDirPath)
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
	rootFilePath := path.Join(rootDirPath, ".existRoot")
	rootFileExists, err := doesFileExist(rootFilePath)
	if err != nil {
		panic(fmt.Sprintf("cannot check for '%s' existence: %s", rootFilePath, err))
	}
	rootDirEmpty, err := isDirectoryEmpty(rootDirPath)
	if err != nil {
		panic(fmt.Sprintf("cannot check if root dir '%s' is empty: %s", rootDirPath, err))
	}
	if !rootFileExists && !rootDirEmpty {
		panic(fmt.Sprintf("root dir '%s' is not empty and is not a exist database", rootDirPath))
	}
	for dir := path.Dir(rootDirPath); ; dir = path.Dir(dir) {
		if ok, _ = doesFileExist(path.Join(dir, ".existRoot")); ok {
			panic(fmt.Sprintf("exist database detected in '%s' above root dir '%s'", dir, rootDirPath))
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
		log.Printf("exist database initialized in '%s'", rootDirPath)
	}

}

func SayHello() {
	fmt.Printf("exist root '%s' ok\n", rootDirPath)
}

type Exister struct{ Persist PersistFunc }

/*
func (self *Exister) Persist(data interface{}, oid uint) (uint, error) {
	fmt.Printf("%T persist (oid %v): %v\n", data, oid, data)
	return oid, nil
}
*/

type PersistFunc func(interface{}, uint) (uint, error)

func (self *Exister) MakePersist(data interface{}, store string) PersistFunc {
	fmt.Printf("MakePersist(%T, %v)\n", data, store)
	return func(data interface{}, oid uint) (uint, error) {
		fmt.Printf("%T persist (%v %v): %v\n", data, store, oid, data)
		return oid, nil
	}
}
