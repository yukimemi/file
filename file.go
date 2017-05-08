package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

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
	Times   []*Time

	matchRe  *regexp.Regexp
	ignoreRe *regexp.Regexp
	getFile  bool
	getDir   bool
}

// Time is filter time option.
type Time struct {
	Base time.Time
	Ope  string
}

// Info is file information struct.
type Info struct {
	Path  string
	Fi    os.FileInfo
	Depth int
	Err   error
}

// DirInfo is directory information struct.
type DirInfo struct {
	Info
	DirSize   int64
	DirCount  int64
	FileCount int64
}

// PathInfo is path information.
type PathInfo struct {
	File     string
	Dir      string
	Name     string
	FileName string
	Info     os.FileInfo
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

// GetPathInfo get PathInfo.
func GetPathInfo(path string) (PathInfo, error) {
	var (
		err error
		pi  = PathInfo{File: path}
	)

	pi.Dir = filepath.Dir(pi.File)
	pi.Name = core.BaseName(pi.File)
	pi.FileName = filepath.Base(pi.File)

	pi.Info, err = os.Stat(pi.File)
	if err != nil {
		return pi, err
	}
	return pi, nil
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

// GetDirInfos return DirSize, FileCount and DirCount under the path.
func GetDirInfos(root string, opt Option) (chan DirInfo, error) {
	var (
		err error
		fn  func(string, int) DirInfo

		q   = make(chan DirInfo, 20)
		sem = make(chan struct{}, runtime.NumCPU())
	)

	// Check exist.
	if !IsExistDir(root) {
		return nil, fmt.Errorf("[%s] is not a directory", root)
	}

	// Compile regexp.
	opt, err = compileRegexps(opt)
	if err != nil {
		return nil, err
	}
	opt.getDir = true

	// qInfo check option and send or not.
	qInfo := func(info DirInfo) {
		if info.Err != nil {
			q <- info
			return
		}

		// Check getFile option.
		if !info.Fi.IsDir() && !opt.getFile {
			return
		}

		// Check getDir option.
		if info.Fi.IsDir() && !opt.getDir {
			return
		}

		// Check Depth option.
		if opt.Depth != 0 && (info.Depth > opt.Depth) {
			return
		}

		if (info.Depth < opt.Depth) && !opt.Recurse {
			return
		}

		// Check time.
		if len(opt.Times) != 0 {
			result := false
			modTime := info.Fi.ModTime()

			for _, t := range opt.Times {
				base := t.Base
				switch t.Ope {
				case "before":
					result = base.Before(modTime)
				case "after":
					result = base.After(modTime)
				case "equal":
					result = base.Equal(modTime)
				default:
					info.Err = fmt.Errorf("Option.Time.Ope: [%v] is not support", t.Ope)
					q <- info
					return
				}
				if !result {
					return
				}
			}
		}

		// Check regexp.
		if opt.matchRe != nil && opt.matchRe.MatchString(info.Path) {
			q <- info
			return
		}

		if opt.ignoreRe != nil && opt.ignoreRe.MatchString(info.Path) {
			return
		}

		if opt.matchRe == nil {
			q <- info
			return
		}
	}

	fn = func(p string, depth int) DirInfo {

		wg := new(sync.WaitGroup)
		fromChild := make(chan DirInfo, 20)
		i := Info{
			Path:  p,
			Depth: depth,
		}
		di := DirInfo{Info: i}
		di.Fi, di.Err = os.Stat(p)
		if di.Err != nil {
			qInfo(di)
			return di
		}

		fis, err := ioutil.ReadDir(p)
		depth++
		if err != nil {
			di.Err = err
			qInfo(di)
			return di
		}

		for _, fi := range fis {
			if fi.IsDir() {
				di.DirCount++
				if (i.Depth < opt.Depth) || opt.Recurse {
					path := filepath.Join(p, fi.Name())
					select {
					case sem <- struct{}{}:
						// Async.
						wg.Add(1)
						go func(path string, depth int) {
							defer wg.Done()
							fromChild <- fn(path, depth)
							<-sem
						}(path, depth)
					default:
						// Sync.
						d := fn(path, depth)
						if d.Err != nil {
							di.Err = d.Err
						}
						di.DirSize += d.DirSize
						di.DirCount += d.DirCount
						di.FileCount += d.FileCount
					}
				}
			} else {
				di.FileCount++
				di.DirSize += fi.Size()
			}
		}

		// Async wait.
		go func() {
			wg.Wait()
			close(fromChild)
		}()

		// Get from child data.
		for cDi := range fromChild {
			if cDi.Err != nil {
				di.Err = cDi.Err
			}
			di.DirSize += cDi.DirSize
			di.DirCount += cDi.DirCount
			di.FileCount += cDi.FileCount
		}

		// Send.
		qInfo(di)
		return di
	}

	// Start and async wait.
	go func() {
		fn(root, 0)
		close(q)
	}()

	return q, err
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
		err error
		fn  func(string, int)

		wg  = new(sync.WaitGroup)
		q   = make(chan Info, 20)
		sem = make(chan struct{}, runtime.NumCPU())
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

		// Check getFile option.
		if !info.Fi.IsDir() && !opt.getFile {
			return
		}

		// Check getDir option.
		if info.Fi.IsDir() && !opt.getDir {
			return
		}

		// Check Depth option.
		if opt.Depth != 0 && (info.Depth > opt.Depth) {
			return
		}

		if (info.Depth < opt.Depth) && !opt.Recurse {
			return
		}

		// Check regexp.
		if opt.matchRe != nil && opt.matchRe.MatchString(info.Path) {
			q <- info
			return
		}

		if opt.ignoreRe != nil && opt.ignoreRe.MatchString(info.Path) {
			return
		}

		// Check time.
		if len(opt.Times) != 0 {
			result := false
			modTime := info.Fi.ModTime()

			for _, t := range opt.Times {
				base := t.Base
				switch t.Ope {
				case "before":
					result = base.Before(modTime)
				case "after":
					result = base.After(modTime)
				case "equal":
					result = base.Equal(modTime)
				default:
					info.Err = fmt.Errorf("Option.Time.Ope: [%v] is not support", t.Ope)
					q <- info
					return
				}
				if !result {
					return
				}
			}
		}

		if opt.matchRe == nil {
			q <- info
			return
		}
	}

	fn = func(p string, depth int) {

		// Send p.
		i := Info{
			Path:  p,
			Depth: depth,
		}
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
		depth++
		if err != nil {
			i.Err = err
			qInfo(i)
			return
		}

		for _, fi := range fis {
			i := Info{
				Path:  filepath.Join(p, fi.Name()),
				Fi:    fi,
				Depth: depth,
			}
			if fi.IsDir() {
				if (i.Depth < opt.Depth) || opt.Recurse {
					select {
					case sem <- struct{}{}:
						// Async.
						wg.Add(1)
						go func(p string, depth int) {
							defer wg.Done()
							fn(p, depth)
							<-sem
						}(i.Path, depth)
					default:
						// Sync.
						fn(i.Path, depth)
					}
				} else {
					qInfo(i)
				}
			} else {
				qInfo(i)
			}
		}
	}

	// Async start get Info list.
	go func() {
		fn(root, 0)
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

// OsCopy is os specific copy command.
func OsCopy(src, dst string) (*core.Cmd, error) {

	var (
		copyCmd string
		copyArg []string
	)
	src = filepath.FromSlash(src)
	dst = filepath.FromSlash(dst)
	if runtime.GOOS == "windows" {
		copyCmd = "cmd"
		copyArg = []string{"/c", "copy"}
	} else {
		copyCmd = "cp"
		copyArg = []string{"-pv"}
	}

	copyArg = append(copyArg, src, dst)

	cmd := core.Cmd{Cmd: exec.Command(copyCmd, copyArg...)}
	err := cmd.CmdRun()
	if err != nil {
		return &cmd, err
	}
	if cmd.ExitCode != 0 {
		return &cmd, fmt.Errorf("copy command exit code: [%v]", cmd.ExitCode)
	}

	return &cmd, nil
}

// Copy is file copy using io.Copy.
func Copy(src, dst string, overwrite bool) (int64, error) {

	src = filepath.FromSlash(src)
	dst = filepath.FromSlash(dst)
	fs, err := os.Open(src)
	if err != nil {
		return -1, err
	}
	defer fs.Close()
	fss, err := fs.Stat()
	if err != nil {
		return -1, err
	}

	if IsExist(dst) && !overwrite {
		fds, err := os.Stat(dst)
		if err != nil {
			return -1, err
		}

		if fss.Size() == fds.Size() && fss.ModTime() == fds.ModTime() {
			return 0, nil
		}
	}

	ds, err := os.Create(dst)
	if err != nil {
		return -1, err
	}
	defer ds.Close()

	n, err := io.Copy(ds, fs)
	if err != nil {
		return -1, err
	}

	err = os.Chtimes(dst, fss.ModTime(), fss.ModTime())
	if err != nil {
		return -1, err
	}

	return n, nil
}
