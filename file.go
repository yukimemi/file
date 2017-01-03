package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
)

// Option is option of GetFiles func.
type Option struct {
	Match   []string
	Ignore  []string
	Recurse bool
}

// Info is file information struct.
type Info struct {
	Path string
	Fi   os.FileInfo
	Err  error
}

// GetFiles return file paths.
func GetFiles(root string, opt Option) (chan Info, error) {
	return getItem(root, opt, "file")
}

// GetDirs return directory paths.
func GetDirs(root string, opt Option) (chan Info, error) {
	return getItem(root, opt, "dir")
}

// GetFilesAndDirs return file and directory paths.
func GetFilesAndDirs(root string, opt Option) (chan Info, error) {
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

// BaseName is get file name without extension.
func BaseName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	re := regexp.MustCompile(ext + "$")
	return re.ReplaceAllString(base, "")
}

// ShareToAbs return abs path not shared.
func ShareToAbs(path string) string {
	rPath := []rune(path)
	sep := '\\'
	// Check shared path.
	if (rPath[0] == sep) && (rPath[1] == sep)  {
		re, err := regexp.Compile(`\\\\([^\\]+)\\(.)\$\\(.*)`)
		if err != nil {
			return path
		}
		return re.ReplaceAllString(path, "$2:\\$3")
	}
	return path
}

func getItem(root string, opt Option, target string) (chan Info, error) {
	var (
		e         error
		fn        func(p string)
		q         = make(chan Info)
		wg        = new(sync.WaitGroup)
		semaphore = make(chan int, runtime.NumCPU())
	)

	// Check root is directory.
	if !IsExistDir(root) {
		return nil, fmt.Errorf("[%s] is not a directory", root)
	}

	// Get file list func.
	fn = func(p string) {
		var info Info

		semaphore <- 1
		defer func() {
			wg.Done()
			<-semaphore
		}()

		fis, e := ioutil.ReadDir(p)
		if e != nil {
			info.Err = e
			q <- info
			return
		}
		for _, fi := range fis {
			info.Path = filepath.Join(p, fi.Name())
			info.Fi = fi
			if fi.IsDir() {
				if target != "file" {
					q <- info
				}
				if opt.Recurse {
					wg.Add(1)
					go fn(info.Path)
				}
			} else {
				if target != "dir" {
					q <- info
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
