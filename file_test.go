package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const fileCnt = 3
const dirCnt = 3

func setup() {
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")
	os.MkdirAll(test, os.ModePerm)
	for i := 0; i < fileCnt; i++ {
		os.Create(filepath.Join(test, "file"+fmt.Sprint(i)))
	}
	for i := 0; i < dirCnt; i++ {
		d := filepath.Join(test, "dir"+fmt.Sprint(i))
		os.MkdirAll(d, os.ModePerm)
		os.Create(filepath.Join(d, "file"+fmt.Sprint(i)))
	}
}

func shutdown() {
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")
	if IsExistDir(test) {
		os.RemoveAll(test)
	}
}

// TestGetFiles is test GetFiles func.
func TestGetFiles(t *testing.T) {
	setup()
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")

	var opt Option
	files, e := GetFiles(test, opt)
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
	shutdown()
}

// TestGetDirs is test GetFiles func.
func TestGetDirs(t *testing.T) {
	setup()
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")

	var opt Option
	dirs, e := GetDirs(test, opt)
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
	shutdown()
}

// TestGetFilesAndDirs is test GetFilesAndDirs func.
func TestGetFilesAndDirs(t *testing.T) {
	setup()
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")

	var opt Option
	files, e := GetFilesAndDirs(test, opt)
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
	shutdown()
}

// TestGetAllRecurse is test GetFilesAndDirs func with option recurse true.
func TestGetAllRecurse(t *testing.T) {
	setup()
	pwd, _ := os.Getwd()
	test := filepath.Join(pwd, "test")

	opt := Option{Recurse: true}
	files, e := GetFiles(test, opt)
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
	shutdown()
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


// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
	shutdown()
}
