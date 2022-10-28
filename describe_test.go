package sconf

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type testconfig struct {
	Bool   bool
	Float  float64
	Name   string `sconf:"optional"`
	Int    int    `sconf-doc:"Int is a ..." sconf:"optional"`
	Int2   int    `sconf-doc:"Int2 is a ..." sconf:"optional"`
	List   []string
	Struct struct {
		Word string
	}
	Ptr       *int `sconf:"optional"`
	Ptr2      *int
	EmptyList []string
	Map1      map[string]bool
	Map2      map[string]struct {
		Word string
	}
	Map3 map[string]struct {
		Word string `sconf:"optional"`
	}
	Map4 map[string]struct {
		Word string
	}
	Map5     map[string]string `sconf:"optional"`
	Map6     map[string]string `sconf:"optional"`
	List2    []string          `sconf:"optional"`
	List3    []string          `sconf:"optional"`
	Duration time.Duration
	Ignore   string `sconf:"-"`
}

var config = testconfig{
	Bool:   true,
	Float:  1.23,
	Name:   "gopher",
	Int2:   2,
	List:   []string{"two", "tone"},
	Struct: struct{ Word string }{"word"},
	Ptr2:   new(int),
	Map1:   map[string]bool{"a": true},
	Map2:   map[string]struct{ Word string }{"x": {"x"}},
	Map3: map[string]struct {
		Word string `sconf:"optional"`
	}{"x": {""}},
	Map4:     map[string]struct{ Word string }{"x": {""}},
	Map5:     nil,
	Map6:     map[string]string{},
	List2:    []string{},
	List3:    nil,
	Duration: time.Second,
	Ignore:   "ignored",
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

	var badMap = struct {
		Map map[int]string
	}{}
	testBad(&badMap, "map key must be string")

	testGood := func(v interface{}, exp string) {
		out := &bytes.Buffer{}
		err := Describe(out, v)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if out.String() != exp {
			t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", exp, out.String())
		}
		if err := Parse(out, &testconfig{}); err != nil {
			t.Fatalf("parsing generated config: %v", err)
		}
	}

	configExp := `Bool: true
Float: 1.230000

# (optional)
Name: gopher

# Int is a ... (optional)
Int: 0

# Int2 is a ... (optional)
Int2: 2
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
Map1:
	a: true
Map2:
	x:
		Word: x
Map3:
	x:

		# (optional)
		Word: 
Map4:
	x:
		Word: 

# (optional)
Map5:
	x: 

# (optional)
Map6:
	x: 

# (optional)
List2:
	- 

# (optional)
List3:
	- 
Duration: 1s
`
	testGood(&config, configExp)
}

func TestWrite(t *testing.T) {
	writeExp := `Bool: true
Float: 1.230000
Name: gopher
Int2: 2
List:
	- two
	- tone
Struct:
	Word: word
Ptr2: 0
EmptyList:
	- nonempty
Map1:
	a: true
Map2:
	x:
		Word: x
Map3:
	x: nil
Map4:
	x:
		Word: 
Duration: 1s
`
	config.EmptyList = []string{"nonempty"}
	out := &bytes.Buffer{}
	err := Write(out, &config)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if out.String() != writeExp {
		t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", writeExp, out.String())
	}
	if err := Parse(out, &testconfig{}); err != nil {
		t.Fatalf("parsing generated config: %v", err)
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
	if err := Parse(out, &emptyListOpt); err != nil {
		t.Fatalf("parsing generated config: %v", err)
	}
}

func TestWriteDocs(t *testing.T) {
	writeExp := `Bool: true
Float: 1.230000

# (optional)
Name: gopher

# Int2 is a ... (optional)
Int2: 2
List:
	- two
	- tone
Struct:
	Word: word
Ptr2: 0
EmptyList:
	- nonempty
Map1:
	a: true
Map2:
	x:
		Word: x
Map3:
	x: nil
Map4:
	x:
		Word: 
Duration: 1s
`
	config.EmptyList = []string{"nonempty"}
	out := &bytes.Buffer{}
	err := WriteDocs(out, &config)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if out.String() != writeExp {
		t.Errorf("expected output:\n%s\n\nactual output:\n%s\n", writeExp, out.String())
	}
	if err := Parse(out, &testconfig{}); err != nil {
		t.Fatalf("parsing generated config: %v", err)
	}
	config.EmptyList = nil

	var emptyList struct {
		List []string
	}
	out = &bytes.Buffer{}
	err = WriteDocs(out, &emptyList)
	if err == nil {
		t.Errorf("got nil, expected %s", errNoElem)
	} else if !strings.Contains(err.Error(), errNoElem.Error()) {
		t.Errorf("got %v, expected %v", err, errNoElem)
	}

	var emptyListOpt struct {
		List []string `sconf:"optional"`
	}
	out = &bytes.Buffer{}
	err = WriteDocs(out, &emptyListOpt)
	if err != nil {
		t.Errorf("got %v, expected nil err", err)
	}
	if err := Parse(out, &emptyListOpt); err != nil {
		t.Fatalf("parsing generated config: %v", err)
	}
}
