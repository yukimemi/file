package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func setup() string {
	/*
	 * Create dir and file.
	 *
	 * |-- dir0 (dir)
	 * |	|-- bar (dir)
	 * |	|	|-- foo (file)
	 * |	|
	 * |	|-- file0 (file)
	 * |	|-- file1 (file)
	 * |	|-- file2 (file)
	 * |	|-- foo (dir)
	 * |	|	|-- bar (file)
	 * |	|-- hoge (dir)
	 * |-- dir1 (dir)
	 * |	|-- bar (file)
	 * |	|-- foo (file)
	 * |	|-- hoge (file)
	 * |-- dir2 (dir)
	 * |-- file0 (file)
	 * |-- file1 (file)
	 * |-- file2 (file)
	 *
	 */

	tmp, err := ioutil.TempDir("", "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Create top directories.
	dir0 := filepath.Join(tmp, "dir0")
	dir1 := filepath.Join(tmp, "dir1")
	dir2 := filepath.Join(tmp, "dir2")
	os.MkdirAll(dir0, os.ModePerm)
	os.MkdirAll(dir1, os.ModePerm)
	os.MkdirAll(dir2, os.ModePerm)

	// Create top files.
	os.Create(filepath.Join(tmp, "file0"))
	os.Create(filepath.Join(tmp, "file1"))
	os.Create(filepath.Join(tmp, "file2"))

	// Create directories and files under the dir0 directory.
	bar := filepath.Join(dir0, "bar")
	os.MkdirAll(bar, os.ModePerm)
	os.Create(filepath.Join(bar, "foo"))

	os.Create(filepath.Join(dir0, "file0"))
	os.Create(filepath.Join(dir0, "file1"))
	os.Create(filepath.Join(dir0, "file2"))

	foo := filepath.Join(dir0, "foo")
	os.MkdirAll(foo, os.ModePerm)
	os.Create(filepath.Join(foo, "bar"))

	hoge := filepath.Join(dir0, "hoge")
	os.MkdirAll(hoge, os.ModePerm)

	// Create files under the dir1 directory.
	os.Create(filepath.Join(dir1, "bar"))
	os.Create(filepath.Join(dir1, "foo"))
	os.Create(filepath.Join(dir1, "hoge"))

	return tmp
}

func shutdown(tmp string) {
	os.RemoveAll(tmp)
}

func getCnt(fn func(string, Option) (chan Info, error), root string, opt Option, t *testing.T) int {
	infos, err := fn(root, opt)
	if err != nil {
		t.Fatal(err)
	}
	cnt := 0
	for i := range infos {
		t.Log(i.Path)
		cnt++
	}
	return cnt
}

// TestGetFiles is test GetFiles func.
func TestGetFiles(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 3
	cnt := getCnt(GetFiles, tmp, Option{}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetDirs is test GetFiles func.
func TestGetDirs(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 4
	cnt := getCnt(GetDirs, tmp, Option{}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfos is test GetInfos func.
func TestGetInfos(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 7
	cnt := getCnt(GetInfos, tmp, Option{}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetFilesRecurse is test GetFiles func with recurse option.
func TestGetFilesRecurse(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 11
	cnt := getCnt(GetFiles, tmp, Option{Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetDirsRecurse is test GetDirs func with recurse option.
func TestGetDirsRecurse(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 7
	cnt := getCnt(GetDirs, tmp, Option{Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfosRecurse is test GetInfos func with recurse option.
func TestGetInfosRecurse(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 18
	cnt := getCnt(GetInfos, tmp, Option{Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfosDepth is test GetInfos func with option depth.
func TestGetInfosDepth(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 3
	cnt := getCnt(GetFiles, tmp, Option{Depth: 1}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 3
	cnt = getCnt(GetDirs, tmp, Option{Depth: 1}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 9
	cnt = getCnt(GetInfos, tmp, Option{Depth: 2}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp += 7
	cnt = getCnt(GetInfos, tmp, Option{Depth: 2, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfosMatch is test GetInfos func with match option.
func TestGetInfosMatch(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 0
	cnt := getCnt(GetFiles, tmp, Option{Matches: []string{"foo"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 0
	cnt = getCnt(GetDirs, tmp, Option{Matches: []string{"foo"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 0
	cnt = getCnt(GetInfos, tmp, Option{Matches: []string{"foo"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 2
	cnt = getCnt(GetFiles, tmp, Option{Matches: []string{`foo$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 1
	cnt = getCnt(GetDirs, tmp, Option{Matches: []string{`foo$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 3
	cnt = getCnt(GetInfos, tmp, Option{Matches: []string{`foo$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfosIgnore is test GetInfos func with ignore option.
func TestGetInfosIgnore(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 3
	cnt := getCnt(GetFiles, tmp, Option{Ignores: []string{"bar"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 4
	cnt = getCnt(GetDirs, tmp, Option{Ignores: []string{"bar"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 7
	cnt = getCnt(GetInfos, tmp, Option{Ignores: []string{"bar"}}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 9
	cnt = getCnt(GetFiles, tmp, Option{Ignores: []string{`bar$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 6
	cnt = getCnt(GetDirs, tmp, Option{Ignores: []string{`bar$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}

	exp = 15
	cnt = getCnt(GetInfos, tmp, Option{Ignores: []string{`bar$`}, Recurse: true}, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetInfosMatchIgnore is test GetInfos func with match and ignore option.
func TestGetInfosMatchIgnore(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 1
	opt := Option{
		Ignores: []string{`dir0`, `dir1`, `file`},
		Matches: []string{`file0$`},
	}
	cnt := getCnt(GetInfos, tmp, opt, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetMatchIgnoreRecurse is test GetInfos func with match, ignore and recurse option.
func TestGetMatchIgnoreRecurse(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	exp := 3
	opt := Option{
		Ignores: []string{`dir0`, `f`, `fo`},
		Matches: []string{`foo$`},
		Recurse: true,
	}
	cnt := getCnt(GetInfos, tmp, opt, t)
	if cnt != exp {
		t.Fatalf("Expected: [%d] but actual: [%d]\n", exp, cnt)
	}
}

// TestGetDirInfo is test GetDirInfo func.
func TestGetDirInfo(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	f0 := filepath.Join(tmp, "file0")
	d1f := filepath.Join(filepath.Join(tmp, "dir1"), "foo")
	d0bf := filepath.Join(filepath.Join(filepath.Join(tmp, "dir0"), "bar"), "foo")

	var (
		efc int64 = 11
		edc int64 = 6
		es  int64 = 12
	)

	ioutil.WriteFile(f0, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	ioutil.WriteFile(d1f, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	ioutil.WriteFile(d0bf, []byte{'t', 'e', 's', 't'}, os.ModePerm)

	di := GetDirInfo(tmp)
	if di.Err != nil {
		t.Fatal(di.Err)
	}
	afc := di.FileCount
	adc := di.DirCount
	as := di.DirSize

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

// TestGetDirInfos is test GetDirInfos func.
func TestGetDirInfos(t *testing.T) {
	tmp := setup()
	defer shutdown(tmp)

	f0 := filepath.Join(tmp, "file0")
	d1f := filepath.Join(filepath.Join(tmp, "dir1"), "foo")
	d0bf := filepath.Join(filepath.Join(filepath.Join(tmp, "dir0"), "bar"), "foo")

	var (
		efc int64 = 11
		edc int64 = 6
		es  int64 = 12
	)

	ioutil.WriteFile(f0, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	ioutil.WriteFile(d1f, []byte{'t', 'e', 's', 't'}, os.ModePerm)
	ioutil.WriteFile(d0bf, []byte{'t', 'e', 's', 't'}, os.ModePerm)

	dis, err := GetDirInfos(tmp, Option{Recurse: true})
	if err != nil {
		t.Fatal(err)
	}

	for di := range dis {
		t.Log(di.Path, di.DirSize, di.DirCount, di.FileCount)
		if di.Path == tmp {
			afc := di.FileCount
			adc := di.DirCount
			as := di.DirSize

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
	}
}

// TestGetPathInfo is test GetPathInfo func.
func TestGetPathInfo(t *testing.T) {

	var err error

	tmp, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown(tmp)

	test := filepath.Join(tmp, "testfile.txt")
	os.Create(test)

	pi, err := GetPathInfo(test)
	if err != nil {
		t.Fatal(err)
	}

	if pi.File != test {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", pi.File, test)
	}

	if pi.Dir != tmp {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", pi.Dir, tmp)
	}

	if pi.Name != "testfile" {
		t.Fatalf("Expected: [%v] but actual: [%v]\n", pi.Name, "testfile")
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

// TestMain is entry point.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
