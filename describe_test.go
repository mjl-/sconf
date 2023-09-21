package sconf

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

type testconfig struct {
	Bool   bool `sconf-doc:"First comment.\n\n\nSecond section.\nWrapped line."`
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
	private  int
}

var config = testconfig{
	Bool:   true,
	Float:  1.23,
	Name:   "gopher",
	Int2:   2,
	List:   []string{"two", "tone"},
	Struct: struct{ Word string }{"word"},
	Ptr2:   new(int),
	Map1:   map[string]bool{"a": true, "b": true, "c": true},
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
	private:  1,
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

	configExp := `# First comment.


# Second section. Wrapped line.
Bool: true
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
	b: true
	c: true
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
	b: true
	c: true
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
	writeExp := `# First comment.


# Second section. Wrapped line.
Bool: true
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
	b: true
	c: true
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

// Test that zero values, taking ignored into account, are properly ignored and
// replaced by nil in a map when writing a config.
func TestMapNil(t *testing.T) {
	type Sub struct {
		Other   string
		Ignored string `sconf:"-"`
	}
	type Value struct {
		Name    string         `sconf:"optional"`
		Sub     Sub            `sconf:"optional"`
		List    []Sub          `sconf:"optional"`
		Map     map[string]Sub `sconf:"optional"`
		Ignored string         `sconf:"-"`
	}
	type xconfig struct {
		Map map[string]Value
	}
	config := xconfig{
		Map: map[string]Value{
			"a": {"", Sub{Ignored: "x"}, nil, nil, "test"},
			"b": {"test", Sub{Ignored: "x"}, nil, nil, ""},
			"c": {"test", Sub{Ignored: "x"}, nil, nil, "test"},
			"d": {"test", Sub{Other: "test"}, nil, nil, "test"},
		},
	}
	// Without the ignored fields.
	expConfig := xconfig{
		Map: map[string]Value{
			"a": {"", Sub{}, nil, nil, ""},
			"b": {"test", Sub{}, nil, nil, ""},
			"c": {"test", Sub{}, nil, nil, ""},
			"d": {"test", Sub{Other: "test"}, nil, nil, ""},
		},
	}
	out := bytes.Buffer{}
	err := Write(&out, config)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	got := out.String()
	exp := `Map:
	a: nil
	b:
		Name: test
	c:
		Name: test
	d:
		Name: test
		Sub:
			Other: test
`
	if got != exp {
		t.Fatalf("got:\n%s\n\nexpected %s", got, exp)
	}

	var nconfig xconfig
	if err := Parse(strings.NewReader(exp), &nconfig); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !reflect.DeepEqual(nconfig, expConfig) {
		t.Fatalf("parse: got %#v, expected %#v", nconfig, expConfig)
	}
}
