package sconf

import (
	"bytes"
	"strings"
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
	Ptr       *int `sconf:"optional"`
	Ptr2      *int
	EmptyList []string
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

	var badString = struct {
		String string
	}{
		"multi\nline\nstring",
	}
	testBad(&badString, "unsupported multiline string")

	type mystring string
	var badCustomString = struct {
		String mystring
	}{
		"multi\nline\nstring",
	}
	testBad(&badCustomString, "unsupported multiline string")

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
EmptyList:
	- 
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
EmptyList:
	- nonempty
`
	config.EmptyList = []string{"nonempty"}
	out := &bytes.Buffer{}
	err := Write(out, &config)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if out.String() != writeExp {
		t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", writeExp, out.String())
	}
	config.EmptyList = nil

	var emptyList struct {
		List []string
	}
	out = &bytes.Buffer{}
	err = Write(out, &emptyList)
	if err == nil {
		t.Errorf("got nil, expected %s", errNoElem)
	} else if !strings.Contains(err.Error(), errNoElem.Error()) {
		t.Errorf("got %v, expected %v", err, errNoElem)
	}

	var emptyListOpt struct {
		List []string `sconf:"optional"`
	}
	out = &bytes.Buffer{}
	err = Write(out, &emptyListOpt)
	if err != nil {
		t.Errorf("got %v, expected nil err", err)
	}
}
