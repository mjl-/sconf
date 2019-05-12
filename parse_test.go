package sconf

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type config1 struct {
	Bool           bool
	Int8           int8
	Int16          int16
	Int32          int32
	Int64          int64
	Uint8          uint8
	Uint16         uint16
	Uint32         uint32
	Uint64         uint64
	Float32        float32
	Float64        float64 `sconf:"optional"`
	Byte           byte
	Bytes          []byte
	Uint           uint
	Int            int
	String         string
	BoolList       []bool
	IntList        []int
	StringList     []string
	StringListList [][]string
	Struct         struct {
		Int        int
		String     string
		List       []string
		StructList []struct {
			Int    int
			String string
		}
		Struct struct {
			Int int
		}
	}
	IntPointer    *int
	StructPointer *struct {
		Bool bool
	}
	StringListPointer *[]string
	ListStringPointer []*string
}

func TestParse(t *testing.T) {
	check := func(err error, action string) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %s\n", action, err)
		}
	}

	run := func(dir string, fn func() interface{}) {
		t.Helper()

		test := func(success bool, src, dst string) {
			t.Helper()
			buf, err := ioutil.ReadFile(dst)
			if err != nil && os.IsNotExist(err) {
				return
			}
			check(err, "reading output file")

			checkResult := func(err error, checkSuffix bool) {
				t.Helper()

				if success && err != nil {
					t.Errorf("%s%s: unexpected error", src, err)
				}
				if !success {
					if err == nil {
						t.Errorf("%s: expected error, but non found", src)
					} else if !checkSuffix && string(buf) != err.Error() || checkSuffix && !strings.HasSuffix(err.Error(), string(buf)) {
						t.Errorf("%s: expected error %q, saw %q", src, string(buf), err.Error())
					}
				}
			}

			v := fn()
			err = ParseFile(src, v)
			checkResult(err, true)

			sf, err := os.Open(src)
			check(err, "open input file")
			defer sf.Close()

			v = fn()
			err = Parse(sf, v)
			checkResult(err, false)
		}
		testOK := func(src, dst string) {
			t.Helper()
			test(true, src, dst)
		}
		testBad := func(src, dst string) {
			t.Helper()
			test(false, src, dst)
		}

		l, err := ioutil.ReadDir(dir)
		check(err, "reading tests from dir")
		for _, f := range l {
			if !strings.HasSuffix(f.Name(), ".input") {
				continue
			}
			base := f.Name()
			base = base[:len(base)-len(".input")]

			testOK(dir+"/"+f.Name(), dir+"/"+base+".ok")
			testBad(dir+"/"+f.Name(), dir+"/"+base+".err")
		}
	}

	newConfig1 := func() interface{} {
		return &config1{}
	}
	run("testdata/config1-parse", newConfig1)

	// Test for unsupported types.
	var config2 struct {
		c complex128
	}
	err := Parse(strings.NewReader("c: 123"), &config2)
	if err == nil {
		t.Errorf("expected error for unsupported complex type")
	} else if err.Error() != ":1: cannot parse type complex128" {
		t.Errorf("unexpected error, got %q, expected %q", err.Error(), ":1: cannot parse type complex128")
	}
}
