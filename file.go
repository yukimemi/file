package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Option is option of GetFiles func.
type Option struct {
	match   []string
	ignore  []string
	recurse bool
}

// Path is return struct.
type Path struct {
	path string
	e    error
}

// GetFiles return file paths.
func GetFiles(root string, opt Option) (chan Path, error) {
	return getItem(root, opt, "file")
}

// GetDirs return directory paths.
func GetDirs(root string, opt Option) (chan Path, error) {
	return getItem(root, opt, "dir")
}

// GetFilesAndDirs return file and directory paths.
func GetFilesAndDirs(root string, opt Option) (chan Path, error) {
	return getItem(root, opt, "all")
}

// IsExist is check file or directory exist.
func IsExist(path string) bool {
	_, e := os.Stat(path)
	if e != nil {
		return false
	}
	return true
}

// IsExistFile is check file exist.
func IsExistFile(path string) bool {
	fi, e := os.Stat(path)
	if e != nil || fi.IsDir() {
		return false
	}
	return true
}

// IsExistDir is check directory exist.
func IsExistDir(path string) bool {
	fi, e := os.Stat(path)
	if e != nil || !fi.IsDir() {
		return false
	}
	return true
}

func getItem(root string, opt Option, target string) (chan Path, error) {
	var (
		e         error
		fn        func(p string)
		q         = make(chan Path)
		wg        = new(sync.WaitGroup)
		semaphore = make(chan int, runtime.NumCPU())
	)

	// Check root is directory.
	if !IsExistDir(root) {
		return nil, fmt.Errorf("[%s] is not a directory", root)
	}

	// Get file list func.
	fn = func(p string) {
		var path Path

		semaphore <- 1
		defer func() {
			wg.Done()
			<-semaphore
		}()

		fis, e := ioutil.ReadDir(p)
		if e != nil {
			path.e = e
			q <- path
			return
		}
		for _, fi := range fis {
			path.path = filepath.Join(p, fi.Name())
			if fi.IsDir() {
				if target != "file" {
					q <- path
				}
				if opt.recurse {
					wg.Add(1)
					go fn(path.path)
				}
			} else {
				if target != "dir" {
					q <- path
				}
			}
		}
	}

	// Start get item list.
	wg.Add(1)
	go fn(root)

	// Wait.
	go func() {
		wg.Wait()
		close(q)
	}()

	return q, e
}
