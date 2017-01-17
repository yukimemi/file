package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/yukimemi/core"
)

const (
	// WINSEPARATOR is windows path separator.
	WINSEPARATOR = '\\'
	// NIXSEPARATOR is unix, linux path separator.
	NIXSEPARATOR = '/'
)

// Option is option of GetFiles func.
type Option struct {
	Matches []string
	Ignores []string
	Recurse bool
	Depth   int
	ErrSkip bool
}

// Info is file information struct.
type Info struct {
	Path string
	Fi   os.FileInfo
	Err  error
}

// DirInfo is directory size and count information struct.
type DirInfo struct {
	Path      string
	Fi        os.FileInfo
	Size      int64
	FileCount int64
	DirCount  int64
	Err       error
	Sub       []*DirInfo
}

var (
	shareRe1 = regexp.MustCompile(`\\\\([^\\]+)\\(.)\$\\(.*)`)
	shareRe2 = regexp.MustCompile(`\\\\([^\\]+)\\(.)\$`)
)

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
	if IsShare(path) {
		if shareRe1.MatchString(path) {
			return shareRe1.ReplaceAllString(path, "$2:\\$3")
		}
		if shareRe2.MatchString(path) {
			return shareRe2.ReplaceAllString(path, "$2:")
		}
	}
	return path
}

// GetCmdPath returns cmd abs path.
func GetCmdPath(cmd string) (string, error) {
	if filepath.IsAbs(cmd) {
		return cmd, nil
	}
	return exec.LookPath(cmd)
}

// GetDirInfoAll is get directory size and count.
func GetDirInfoAll(root string, opt Option) (chan DirInfo, error) {

	var (
		err    error
		fn     func(p string) DirInfo
		match  *regexp.Regexp
		ignore *regexp.Regexp
		q      = make(chan DirInfo)
	)

	// Check root is directory.
	if !IsExistDir(root) {
		return nil, fmt.Errorf("[%s] is not a directory", root)
	}

	if len(opt.Matches) != 0 {
		match, err = core.CompileStrs(opt.Matches)
		if err != nil {
			return nil, err
		}
	}
	if len(opt.Ignores) != 0 {
		ignore, err = core.CompileStrs(opt.Ignores)
		if err != nil {
			return nil, err
		}
	}

	// Get directory size and count.
	fn = func(p string) DirInfo {

		var (
			fis []os.FileInfo
			di  = DirInfo{Path: p}
		)

		// Check ignore.
		if ignore != nil && ignore.MatchString(p) {
			return DirInfo{}
		}

		di.Fi, di.Err = os.Stat(p)
		if di.Err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", di.Err)
				return DirInfo{}
			}
			q <- di
			return di
		}

		fis, di.Err = ioutil.ReadDir(p)
		if di.Err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", di.Err)
				return DirInfo{}
			}
			q <- di
			return di
		}

		for _, fi := range fis {

			if fi.IsDir() {
				di.DirCount++
				if opt.Recurse {
					sub := fn(filepath.Join(p, fi.Name()))
					if sub.Path != "" {
						di.Sub = append(di.Sub, &sub)
					}
					if sub.Err != nil {
						if opt.ErrSkip {
							fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", sub.Err)
							continue
						}
						return di
					}
				}
			} else {
				di.FileCount++
				di.Size += fi.Size()
			}
		}

		if match == nil || match.MatchString(di.Path) {
			q <- di
		}
		return di
	}

	// Start get item list.
	go func() {
		_ = fn(root)
		close(q)
	}()

	return q, err
}

// GetDirInfoRecurse is get size and count under the DirInfo recurse.
func GetDirInfoRecurse(di DirInfo, opt Option) DirInfo {

	var (
		d = DirInfo{
			Path:      di.Path,
			Size:      di.Size,
			FileCount: di.FileCount,
			DirCount:  di.DirCount,
			Err:       di.Err,
			Sub:       di.Sub,
		}
	)

	if d.Err != nil {
		return d
	}

	for _, sub := range di.Sub {
		if sub.Err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", sub.Err)
				continue
			}
			d.Err = sub.Err
			return d
		}
		subDi := GetDirInfoRecurse(*sub, opt)
		if subDi.Err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", subDi.Err)
				continue
			}
			d.Err = subDi.Err
			return d
		}
		d.FileCount += subDi.FileCount
		d.DirCount += subDi.DirCount
		d.Size += subDi.Size
	}

	return d
}

// GetDirInfo is get size and count the directory.
func GetDirInfo(root string, opt Option) (DirInfo, error) {

	dis, err := GetDirInfoAll(root, opt)
	if err != nil {
		return DirInfo{}, err
	}

	for di := range dis {
		if di.Err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", di.Err)
				continue
			}
			return DirInfo{}, di.Err
		}
		if di.Path == root {
			return GetDirInfoRecurse(di, opt), nil
		}
	}

	return DirInfo{}, fmt.Errorf("Error occured")
}

func getItem(root string, opt Option, target string) (chan Info, error) {
	var (
		err       error
		fn        func(p string)
		match     *regexp.Regexp
		ignore    *regexp.Regexp
		q         = make(chan Info)
		wg        = new(sync.WaitGroup)
		semaphore = make(chan struct{}, runtime.NumCPU())
	)

	// Check root is directory.
	if !IsExistDir(root) {
		return nil, fmt.Errorf("[%s] is not a directory", root)
	}

	if len(opt.Matches) != 0 {
		match, err = core.CompileStrs(opt.Matches)
		if err != nil {
			return nil, err
		}
	}
	if len(opt.Ignores) != 0 {
		ignore, err = core.CompileStrs(opt.Ignores)
		if err != nil {
			return nil, err
		}
	}

	// Get file list func.
	fn = func(p string) {
		var info Info

		semaphore <- struct{}{}
		defer func() {
			wg.Done()
			<-semaphore
		}()

		fis, err := ioutil.ReadDir(p)
		if err != nil {
			if opt.ErrSkip {
				fmt.Fprintf(os.Stderr, "Warning: [%s] continue.\n", err)
				return
			}
			info.Err = err
			q <- info
			return
		}
		for _, fi := range fis {
			info.Path = filepath.Join(p, fi.Name())
			// Check ignore.
			if ignore != nil && ignore.MatchString(info.Path) {
				continue
			}

			info.Fi = fi

			if fi.IsDir() {
				if target != "file" {
					if match == nil || match.MatchString(info.Path) {
						q <- info
					}
				}
				if opt.Recurse {
					wg.Add(1)
					go fn(info.Path)
				}
			} else {
				if target != "dir" {
					if match == nil || match.MatchString(info.Path) {
						q <- info
					}
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

	return q, err
}

// GetDepth is return path depth.
func GetDepth(path string, sep rune) int {
	if sep == 0 {
		sep = filepath.Separator
	}
	p := strings.Replace(filepath.Clean(path), string(sep)+string(sep), string(sep), -1)
	c := strings.Count(p, string(sep))
	if IsShare(path) {
		return c - 1
	}
	return c
}

// IsShare is whether path is share or not.
func IsShare(path string) bool {

	rPath := []rune(path)
	// Check shared path.
	if ((rPath[0] == WINSEPARATOR) && (rPath[1] == WINSEPARATOR)) ||
		((rPath[0] == NIXSEPARATOR) && (rPath[1] == NIXSEPARATOR)) {
		return true
	}
	return false
}
