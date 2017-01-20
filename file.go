package file

import (
	"fmt"
	"io/ioutil"
	"os"
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
	Matches  []string
	Ignores  []string
	Recurse  bool
	matchRe  *regexp.Regexp
	ignoreRe *regexp.Regexp
	getFile  bool
	getDir   bool
	Depth    int
}

// Info is file information struct.
type Info struct {
	Path string
	Fi   os.FileInfo
	Err  error
}

// DirInfo is directory information struct.
type DirInfo struct {
	Info
	DirSize   int64
	DirCount  int64
	FileCount int64
}

var (
	shareRe1 = regexp.MustCompile(`\\\\([^\\]+)\\(.)\$\\(.*)`)
	shareRe2 = regexp.MustCompile(`\\\\([^\\]+)\\(.)\$`)
)

// GetFiles return file infos.
func GetFiles(root string, opt Option) (chan Info, error) {
	opt.getFile = true
	opt, err := compileRegexps(opt)
	if err != nil {
		return nil, err
	}
	return getInfo(root, opt)
}

// GetDirs return directory infos.
func GetDirs(root string, opt Option) (chan Info, error) {
	opt.getDir = true
	opt, err := compileRegexps(opt)
	if err != nil {
		return nil, err
	}
	return getInfo(root, opt)
}

// GetInfos return file and directory infos.
func GetInfos(root string, opt Option) (chan Info, error) {
	opt.getFile, opt.getDir = true, true
	opt, err := compileRegexps(opt)
	if err != nil {
		return nil, err
	}
	return getInfo(root, opt)
}

func asyncToSync(root string, opt Option, fn func(string, Option) (chan Info, error)) Info {
	infos, err := fn(root, opt)
	if err != nil {
		return Info{Err: err}
	}
	for info := range infos {
		if info.Path == root {
			return info
		}
	}
	return Info{Err: fmt.Errorf("Error ! Not found [%v]", root)}
}

// GetFile return file info.
func GetFile(root string, opt Option) Info {
	return asyncToSync(root, opt, GetFiles)
}

// GetDir return directory info.
func GetDir(root string, opt Option) Info {
	return asyncToSync(root, opt, GetDirs)
}

// GetInfo return file and directory info.
func GetInfo(root string, opt Option) Info {
	return asyncToSync(root, opt, GetInfos)
}

// IsExist is check file or directory exist.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// IsExistFile is check file exist.
func IsExistFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return false
	}
	return true
}

// IsExistDir is check directory exist.
func IsExistDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return false
	}
	return true
}

// ShareToAbs return abs path not shared.
func ShareToAbs(path string) string {
	if IsShare(path) {
		if shareRe1.MatchString(path) {
			return shareRe1.ReplaceAllString(path, "$2:\\$3")
		}
		if shareRe2.MatchString(path) {
			return shareRe2.ReplaceAllString(path, "$2:\\")
		}
	}
	return path
}

// GetDirInfo return DirSize, FileCount and DirCount.
func GetDirInfo(path string) DirInfo {

	i := Info{Path: path}
	di := DirInfo{Info: i}

	infos, err := GetInfos(path, Option{Recurse: true})
	if err != nil {
		di.Err = err
		return di
	}

	for i := range infos {
		if i.Err != nil {
			di.Err = i.Err
			continue
		}
		if i.Fi.IsDir() {
			di.DirCount++
		} else {
			di.FileCount++
			di.DirSize += i.Fi.Size()
		}
	}
	di.DirCount--
	return di
}

func getInfo(root string, opt Option) (chan Info, error) {
	var (
		err       error
		fn        func(string)
		wg        = new(sync.WaitGroup)
		q         = make(chan Info, 20)
		semaphore = make(chan struct{}, runtime.NumCPU())
	)

	// Check exist.
	if !IsExist(root) {
		return nil, fmt.Errorf("[%s] is not found", root)
	}

	// qInfo check option and send or not.
	qInfo := func(info Info) {
		if info.Err != nil {
			q <- info
			return
		}

		// Check option.
		if (!opt.getFile && !info.Fi.IsDir()) ||
			(!opt.getDir && info.Fi.IsDir()) {
			// Not send.
		} else if opt.matchRe != nil && opt.matchRe.MatchString(info.Path) {
			// Send.
			q <- info
		} else if opt.ignoreRe != nil && opt.ignoreRe.MatchString(info.Path) {
			// Not send.
		} else if opt.matchRe == nil {
			// Send.
			q <- info
		}
	}

	fn = func(p string) {
		semaphore <- struct{}{}
		defer func() {
			wg.Done()
			<-semaphore
		}()

		// Send p.
		i := Info{Path: p}
		i.Fi, i.Err = os.Stat(p)
		qInfo(i)
		if i.Err != nil {
			return
		}

		// File pattern.
		if !i.Fi.IsDir() {
			return
		}

		fis, err := ioutil.ReadDir(p)
		if err != nil {
			i.Err = err
			qInfo(i)
			return
		}

		for _, fi := range fis {
			i := Info{
				Path: filepath.Join(p, fi.Name()),
				Fi:   fi,
			}
			if fi.IsDir() && opt.Recurse {
				wg.Add(1)
				go fn(i.Path)
			} else {
				qInfo(i)
			}
		}
	}

	// Start get item list.
	go func() {
		wg.Add(1)
		fn(root)
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

func compileRegexps(opt Option) (Option, error) {
	var err error
	// Compile regexp.
	if len(opt.Matches) != 0 {
		opt.matchRe, err = core.CompileStrs(opt.Matches)
		if err != nil {
			return opt, err
		}
	}
	if len(opt.Ignores) != 0 {
		opt.ignoreRe, err = core.CompileStrs(opt.Ignores)
		if err != nil {
			return opt, err
		}
	}
	return opt, err
}
