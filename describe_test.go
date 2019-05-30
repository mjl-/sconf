package sconf

import (
	"bytes"
	"testing"
)

var config = struct {
	Bool   bool
	Float  float64
	Name   string `sconf:"optional"`
	Int    int    `sconf-doc:"Int is a ..." sconf:"optional"`
	List   []string
	Struct struct {
		Word string
	}
	Ptr  *int `sconf:"optional"`
	Ptr2 *int
}{
	Bool:   true,
	Float:  1.23,
	Name:   "gopher",
	List:   []string{"two", "tone"},
	Struct: struct{ Word string }{"word"},
	Ptr2:   new(int),
}

func TestDescribe(t *testing.T) {
	testBad := func(v interface{}, exp string) {
		t.Helper()
		err := Describe(&bytes.Buffer{}, v)
		if err == nil {
			t.Errorf("missing error")
		} else if err.Error() != exp {
			t.Errorf("expected error %q, saw: %s", exp, err.Error())
		}
	}

	err := Describe(&bytes.Buffer{}, "not a struct")
	if err == nil {
		t.Errorf("missing error")
	} else if err.Error() != "top level object must be a struct, is a string" {
		t.Errorf("unexpected error, got %v", err)
	}

	var badFunc struct {
		Func func()
	}
	testBad(&badFunc, "unsupported value func")

	var badInterface struct {
		Interface interface{}
	}
	testBad(&badInterface, "unsupported value interface")

	var badMap struct {
		Map map[string]string
	}
	testBad(&badMap, "unsupported value map")

	var badChan struct {
		Channel chan int
	}
	testBad(&badChan, "unsupported value chan")

	testGood := func(v interface{}, exp string) {
		out := &bytes.Buffer{}
		err := Describe(out, v)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if out.String() != exp {
			t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", exp, out.String())
		}
	}

	configExp := `Bool: true
Float: 1.230000
# (optional)
Name: gopher
# Int is a ... (optional)
Int: 0
List:
	- two
	- tone
Struct:
	Word: word
# (optional)
Ptr: 0
Ptr2: 0
`
	testGood(&config, configExp)
}

func TestWrite(t *testing.T) {
	writeExp := `Bool: true
Float: 1.230000
Name: gopher
List:
	- two
	- tone
Struct:
	Word: word
Ptr2: 0
`
	out := &bytes.Buffer{}
	err := Write(out, &config)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if out.String() != writeExp {
		t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", writeExp, out.String())
	}
}
