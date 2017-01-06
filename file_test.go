package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const fileCnt = 3
const dirCnt = 3

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

func shutdown(temp string) {
	os.RemoveAll(temp)
}

// TestGetFiles is test GetFiles func.
func TestGetFiles(t *testing.T) {
	temp := setup()

	var opt Option
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
	shutdown(temp)
}

// TestGetDirs is test GetFiles func.
func TestGetDirs(t *testing.T) {
	temp := setup()

	var opt Option
	dirs, e := GetDirs(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for d := range dirs {
		t.Log(d.Path)
		cnt++
	}
	if cnt != dirCnt {
		t.Fail()
	}
	shutdown(temp)
}

// TestGetFilesAndDirs is test GetFilesAndDirs func.
func TestGetFilesAndDirs(t *testing.T) {
	temp := setup()

	var opt Option
	files, e := GetFilesAndDirs(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt+dirCnt {
		t.Fail()
	}
	shutdown(temp)
}

// TestGetAllRecurse is test GetFilesAndDirs func with option recurse true.
func TestGetAllRecurse(t *testing.T) {
	temp := setup()

	opt := Option{Recurse: true}
	files, e := GetFiles(temp, opt)
	if e != nil {
		t.FailNow()
	}
	cnt := 0
	for f := range files {
		t.Log(f.Path)
		cnt++
	}
	if cnt != fileCnt+dirCnt {
		t.Fail()
	}
	shutdown(temp)
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
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "\\\\10.10.99.88\\d$\\パス\\トゥ\\日本語パス.txt.ext"
	e = "d:\\パス\\トゥ\\日本語パス.txt.ext"

	a = ShareToAbs(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
	}

	p = "\\\\10.10.99.88\\d\\パス\\トゥ\\日本語パス.txt.ext"
	e = "\\\\10.10.99.88\\d\\パス\\トゥ\\日本語パス.txt.ext"

	a = ShareToAbs(p)
	if a != e {
		t.Errorf("Expected: [%s] but actual: [%s]\n", e, a)
		t.Fail()
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

	p = "C:\\bin\\go"
	e = "C:\\bin\\go"

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
	shutdown(temp)
}

// TestGetFilesIgnore is test GetFiles func with ignore option.
func TestGetFilesIgnore(t *testing.T) {
	temp := setup2()

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
	shutdown(temp)
}

// TestGetFilesMatchIgnore is test GetFiles func with match and ignore option.
func TestGetFilesMatchIgnore(t *testing.T) {
	temp := setup2()

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
	shutdown(temp)
}

// TestGetFilesMatchIgnoreRecurse is test GetFiles func with match, ignore and recurse option.
func TestGetFilesMatchIgnoreRecurse(t *testing.T) {
	temp := setup2()

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
	// shutdown(temp)
}

// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
