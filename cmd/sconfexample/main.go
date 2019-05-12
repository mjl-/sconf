package main

import (
	"log"
	"os"

	"github.com/mjl-/sconf"
)

var config struct {
	StringKey string `sconf-doc:"comment for stringKey" sconf:"optional"`
	IntKey    int64
	BoolKey   bool
	Struct    struct {
		A int `sconf-doc:"this is the A-field"`
		B bool
		C string `sconf:"optional"`
	}
	StringArray []string
	Nested      []struct {
		A int
		B bool
		C string
	} `sconf-doc:"nested structs work just as well"`
}

func check(err error, action string) {
	if err != nil {
		log.Fatalf("%s: %s\n", action, err)
	}
}

func usage() {
	log.Fatalln("usage: sconfexample { describe | parse [file.conf] }")
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "describe":
		describe(os.Args[1:])
	case "parse":
		parse(os.Args[1:])
	default:
		usage()
	}
}

func describe(args []string) {
	if len(args) != 1 {
		usage()
	}
	config.StringKey = "value1"
	config.IntKey = 123
	config.BoolKey = true
	config.Struct.A = 321
	config.Struct.B = true
	config.Struct.C = "this is text"
	config.StringArray = []string{"blah", "blah"}
	config.Nested = []struct {
		A int
		B bool
		C string
	}{
		{1, false, "hoi"},
		{-1, true, "hallo"},
	}

	err := sconf.Describe(os.Stdout, config)
	check(err, "describing config")
}

func parse(args []string) {
	switch len(args) {
	case 1:
		err := sconf.Parse(os.Stdin, &config)
		check(err, "parse")
	case 2:
		err := sconf.ParseFile(args[1], &config)
		check(err, "parsefile")
	default:
		usage()
	}
	err := sconf.Describe(os.Stdout, config)
	check(err, "describing config")
}
