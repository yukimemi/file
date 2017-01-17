package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	fileCnt = 3
	dirCnt  = 3
)

var (
	err error
)

func setup() string {
	temp, err := ioutil.TempDir("", "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(temp, "file"+fmt.Sprint(i)))
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(temp, "dir"+fmt.Sprint(i))
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Create(filepath.Join(d, "file"+fmt.Sprint(i)))
	}
	return temp
}

func setup2() string {
	temp, err := ioutil.TempDir("", "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(temp, "hoge"+fmt.Sprint(i)))
	}
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(temp, "fuga"+fmt.Sprint(i)))
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(temp, "foo"+fmt.Sprint(i))
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Create(filepath.Join(d, "hoge"+fmt.Sprint(i)))
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(temp, "bar"+fmt.Sprint(i))
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Create(filepath.Join(d, "fuga"+fmt.Sprint(i)))
	}
	return temp
}

func setup3() string {
	temp, err := ioutil.TempDir("", "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(temp, "file"+fmt.Sprint(i)))
	}
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(temp, "file2"+fmt.Sprint(i)))
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(temp, "dir"+fmt.Sprint(i))
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Create(filepath.Join(d, "file"+fmt.Sprint(i)))
		for j := 0; j < dirCnt; j++ {
			d := filepath.Join(filepath.Join(temp, "dir"+fmt.Sprint(j)), "dir"+fmt.Sprint(j))
			err := os.MkdirAll(d, os.ModePerm)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Create(filepath.Join(d, "file"+fmt.Sprint(j)))
		}
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(temp, "bar"+fmt.Sprint(i))
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Create(filepath.Join(d, "fuga"+fmt.Sprint(i)))
		for j := 0; j < dirCnt; j++ {
			d := filepath.Join(filepath.Join(temp, "bar"+fmt.Sprint(j)), "baz"+fmt.Sprint(j))
			err := os.MkdirAll(d, os.ModePerm)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Create(filepath.Join(d, "hoge"+fmt.Sprint(j)))
		}
	}
	return temp
}

func shutdown(temp string) {
	os.RemoveAll(temp)
}

// TestGetFiles is test GetFiles func.
func TestGetFiles(t *testing.T) {
	temp := setup()
	defer shutdown(temp)

	var opt Option
	files, err := GetFiles(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", fileCnt, cnt)
	}
}

// TestGetDirs is test GetFiles func.
func TestGetDirs(t *testing.T) {
	temp := setup()
	defer shutdown(temp)

	var opt Option
	dirs, err := GetDirs(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for d := range dirs {
		t.Log(d.Path)
		cnt++
	}
	if cnt != dirCnt {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", fileCnt, cnt)
	}
}

// TestGetFilesAndDirs is test GetFilesAndDirs func.
func TestGetFilesAndDirs(t *testing.T) {
	temp := setup()
	defer shutdown(temp)

	var opt Option
	files, err := GetFilesAndDirs(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt+dirCnt {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", fileCnt, cnt)
	}
}

// TestGetAllRecurse is test GetFilesAndDirs func with option recurse true.
func TestGetAllRecurse(t *testing.T) {
	temp := setup()
	defer shutdown(temp)

	opt := Option{Recurse: true}
	files, err := GetFiles(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt+dirCnt {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", fileCnt, cnt)
	}
}

// TestBaseName is test BaseName fucn.
func TestBaseName(t *testing.T) {
	p := "/path/to/file.txt"
	e := "file"

	a := BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "/path/to/file.txt.ext"
	e = "file.txt"

	a = BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "/パス/トゥ/日本語パス.txt.ext"
	e = "日本語パス.txt"

	a = BaseName(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}
}

// TestShareToAbs test ShareToAbs func.
func TestShareToAbs(t *testing.T) {
	p := "\\\\192.168.1.1\\C$\\test\\hoge\\bar.txt"
	e := "C:\\test\\hoge\\bar.txt"

	a := ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\d$\\パス\\トゥ\\日本語パス.txt.ext"
	e = "d:\\パス\\トゥ\\日本語パス.txt.ext"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\d\\パス\\トゥ\\日本語パス.txt.ext"
	e = "\\\\10.10.99.88\\d\\パス\\トゥ\\日本語パス.txt.ext"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\C$"
	e = "C:\\"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\C$"
	e = "C:\\"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\D$"
	e = "D:\\"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}

	p = "\\\\10.10.99.88\\d$"
	e = "d:\\"

	a = ShareToAbs(p)
	if a != e {
		t.Fatalf("Expected: [%s] but actual: [%s]\n", e, a)
	}
}

// TestGetDepth test GetDepth func.
func TestGetDepth(t *testing.T) {
	p := "C:\\test\\hoge\\bar.txt"
	e := 3

	a := GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}

	p = "\\\\10.10.99.88\\d$\\パス\\トゥ\\日本語パス.txt.ext"
	e = 4

	a = GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}

	p = "C:\\"
	e = 1

	a = GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}

	p = "C:"
	e = 0

	a = GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}

	p = "\\\\10.10.99.88\\C$\\"
	e = 2

	a = GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}

	p = "\\\\10.10.99.88\\C$"
	e = 1

	a = GetDepth(p, '\\')
	if a != e {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", e, a)
	}
}

// TestGetCmdPath test GetCmdPath func.
func TestGetCmdPath(t *testing.T) {
	p := "go"
	e, err := exec.LookPath("go")
	if err != nil {
		t.Fail()
	}

	a, err := GetCmdPath(p)
	if err != nil {
		t.Fail()
	}
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	if runtime.GOOS == "windows" {
		p = "C:\\bin\\go"
		e = "C:\\bin\\go"
	} else {
		p = "/opt/local/bin/go"
		e = "/opt/local/bin/go"
	}

	a, err = GetCmdPath(p)
	if err != nil {
		t.Fail()
	}
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}
}

// TestGetFilesMatch is test GetFiles func with match option.
func TestGetFilesMatch(t *testing.T) {
	temp := setup2()
	defer shutdown(temp)

	opt := Option{
		Matches: []string{"hoge"},
	}
	files, e := GetFiles(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt {
		t.Fail()
	}
}

// TestGetFilesIgnore is test GetFiles func with ignore option.
func TestGetFilesIgnore(t *testing.T) {
	temp := setup2()
	defer shutdown(temp)

	opt := Option{
		Ignores: []string{"hoge"},
	}
	files, e := GetFiles(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt {
		t.Fail()
	}
}

// TestGetFilesMatchIgnore is test GetFiles func with match and ignore option.
func TestGetFilesMatchIgnore(t *testing.T) {
	temp := setup2()
	defer shutdown(temp)

	opt := Option{
		Matches: []string{"fuga"},
		Ignores: []string{"hoge", "fuga0$"},
	}
	files, e := GetFiles(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt-1 {
		t.Fail()
	}
}

// TestGetFilesMatchIgnoreRecurse is test GetFiles func with match, ignore and recurse option.
func TestGetFilesMatchIgnoreRecurse(t *testing.T) {
	temp := setup2()
	defer shutdown(temp)

	opt := Option{
		Matches: []string{"fuga"},
		Ignores: []string{"hoge", "fuga0$"},
		Recurse: true,
	}
	files, e := GetFiles(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt*2-2 {
		t.Fail()
	}
}

// TestGetDirInfoAll is test GetDirInfoAll func.
func TestGetDirInfoAll(t *testing.T) {
	var (
		temp         string
		dis          chan DirInfo
		afc, adc, as int64

		opt       = Option{Recurse: true}
		efc int64 = fileCnt * 2
		edc int64 = dirCnt
		es  int64 = 4 * 2
	)

	temp = setup()
	defer shutdown(temp)

	err = ioutil.WriteFile(filepath.Join(temp, "file0"), []byte{'t', 'e', 's', 't'}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(filepath.Join(temp, "dir0"), "file0"), []byte{'t', 'e', 's', 't'}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	dis, err = GetDirInfoAll(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	for di := range dis {
		afc += di.FileCount
		adc += di.DirCount
		as += di.Size
	}
	if afc != efc {
		t.Fatalf("File count expected: [%d] but actual: [%d]\n", efc, afc)
	}
	if adc != edc {
		t.Fatalf("Dir count expected: [%d] but actual: [%d]\n", edc, adc)
	}
	if as != es {
		t.Fatalf("Size expected: [%d] but actual: [%d]\n", es, as)
	}

}

// TestGetDirInfo is test GetDirInfo func.
func TestGetDirInfo(t *testing.T) {
	var (
		temp         string
		di           DirInfo
		afc, adc, as int64

		opt       = Option{Recurse: true}
		efc int64 = fileCnt*2 + dirCnt*4
		edc int64 = dirCnt * 4
		es  int64 = 4 * 3
	)
	temp = setup3()
	defer shutdown(temp)
	f0 := filepath.Join(temp, "file0")
	d0f0 := filepath.Join(filepath.Join(temp, "dir0"), "file0")
	d0d0f0 := filepath.Join(filepath.Join(filepath.Join(temp, "dir0"), "dir0"), "file0")

	err = ioutil.WriteFile(f0, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(d0f0, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(d0d0f0, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	di, err = GetDirInfo(temp, opt)
	if err != nil {
		t.Fatal(err)
	}
	afc = di.FileCount
	adc = di.DirCount
	as = di.Size

	if afc != efc {
		t.Fatalf("File count expected: [%d] but actual: [%d]\n", efc, afc)
	}
	if adc != edc {
		t.Fatalf("Dir count expected: [%d] but actual: [%d]\n", edc, adc)
	}
	if as != es {
		t.Fatalf("Size expected: [%d] but actual: [%d]\n", es, as)
	}

}

// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
