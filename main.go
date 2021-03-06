////////////////////////////////////////////////////////////////////////////
// Porgram: main
// Purpose: wts (web test script) verbs handling
// authors: Antonio Sun (c) 2015, All rights reserved
// Credits: https://github.com/voxelbrain/goptions/tree/master/examples
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"os"
)

import (
	"github.com/voxelbrain/goptions"
)

////////////////////////////////////////////////////////////////////////////
// Configuration variables definitions

var progname = "wts"
var progdesc = " - web test script processing program"
var buildTime = "2016-03-15"

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

type Check struct {
	Filei     *os.File `goptions:"-i, --input, obligatory, description='The web test script to check', rdonly"`
	Checks    string   `goptions:"-c, --check, description='Check regexp'"`
	ThinkTime int      `goptions:"--thinktime, description='ThinkTime canonical value (default: 0)'"`
	Timeout   int      `goptions:"--timeout, description='Timeout canonical value'"`
}

type Options struct {
	Verbosity []bool        `goptions:"-v, --verbose, description='Be verbose'"`
	Quiet     bool          `goptions:"-q, --quiet, description='Do not print anything, even errors (except if --verbose is specified)'"`
	Help      goptions.Help `goptions:"-h, --help, description='Show this help'"`

	goptions.Verbs

	Check `goptions:"check"` // Embedding!

	Dump struct {
		Filei *os.File `goptions:"-i, --input, obligatory, description='The web test script to dump', rdonly"`
		Fileo *os.File `goptions:"-o, --output, description='The web test script dump output (default: .webtext file of input)', wronly"`
		Asis  bool     `goptions:"--asis, description='Output StringBody as-is, no XML decoding'"`
		Cnr   bool     `goptions:"-c, --cnr, description='Comment number removal, for easy comparison'"`
		Tsr   bool     `goptions:"-t, --tsr, description='Time string removal, for easy comparison'"`
		Raw   bool     `goptions:"-r, --raw, description='Raw mode, for fresh recordings and easy comparison\n\t\t\t\tWill enable --cnr as well and \n\t\t\t\tapply rules from the .rawrule file if exist'"`
	} `goptions:"dump"`
}

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var options = Options{ // Default values goes here
	Check: Check{
		Checks:    `\d\d*/\d\d*/20\d\d|20\d\d-`,
		ThinkTime: 0,
		Timeout:   270,
	},
}

type Command func() error

var commands = map[goptions.Verbs]Command{
	"check": checkCmd,
	"dump":  dumpCmd,
}

var (
	VERBOSITY = 0
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

func main() {
	goptions.ParseAndFail(&options)
	//fmt.Printf("] %#v\r\n", options)

	if len(options.Verbs) == 0 {
		fmt.Printf("%s%s \n      built on %s\n\n", progname, progdesc, buildTime)
		goptions.PrintHelp()
		os.Exit(2)
	}

	VERBOSITY = len(options.Verbosity)

	if cmd, found := commands[options.Verbs]; found {
		err := cmd()
		if err != nil {
			if !options.Quiet {
				fmt.Printf("%s error: %v", progname, err)
			}
			os.Exit(1)
		}
	}
}
