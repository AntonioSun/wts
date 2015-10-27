////////////////////////////////////////////////////////////////////////////
// Porgram: wts-main - web test script main
// Purpose: wts handling
// authors: Antonio Sun (c) 2015, All rights reserved
// Credits: https://github.com/voxelbrain/goptions/tree/master/examples
//
//
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"os"
)

import (
	"github.com/voxelbrain/goptions"
)

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	//	"os"
)

type Comment struct {
	Comment string `xml:"CommentText,attr"`
}

type IncludedWebTest struct {
	Included string `xml:"Name,attr"`
}

func getDecoder(wtsFile string) *xml.Decoder {
	content, err := ioutil.ReadFile(wtsFile)
	check(err)
	return xml.NewDecoder(bytes.NewBuffer(content))
}

func dumpWtsXml(decoder *xml.Decoder) error {

	for {
		// Read tokens from the XML document in a stream.
		token, _ := decoder.Token()
		if token == nil {
			break
		}

		// Inspect the type of the token just read.
		switch t := token.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			inElement := t.Name.Local
			switch inElement {
			case "Comment":
				{
					var c Comment
					// decode a whole chunk of following XML into the
					// variable c which is a Comment (t above)
					decoder.DecodeElement(&c, &t)
					fmt.Printf("C: %s\n", c.Comment)
				}
			case "IncludedWebTest":
				{
					var r IncludedWebTest
					decoder.DecodeElement(&r, &t)
					fmt.Printf("I: %s\n", r.Included)
				}
			}
		default:
		}
	}

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Options struct {
	Verbosity []bool        `goptions:"-v, --verbose, description='Be verbose'"`
	Quiet     bool          `goptions:"-q, --quiet, description='Do not print anything, even errors (except if --verbose is specified)'"`
	Help      goptions.Help `goptions:"-h, --help, description='Show this help'"`

	goptions.Verbs

	Check struct {
	} `goptions:"check"`

	Dump struct {
		Cnr       string `goptions:"-c, --cnr, mutexgroup='input', description='Comment number removal, for easy comparison'"`
		Remainder goptions.Remainder
	} `goptions:"dump"`
}

var options = Options{ // Default values goes here
}

type Command func(Options) error

var commands = map[goptions.Verbs]Command{
	"check": checkCmd,
	"dump":  dumpCmd,
}

var (
	VERBOSITY = 0
)

func main() {
	goptions.ParseAndFail(&options)

	if len(options.Verbs) == 0 {
		goptions.PrintHelp()
		os.Exit(2)
	}

	VERBOSITY = len(options.Verbosity)

	if cmd, found := commands[options.Verbs]; found {
		err := cmd(options)
		if err != nil {
			if !options.Quiet {
				fmt.Println("error:", err)
			}
			os.Exit(1)
		}
	}
}

func dumpCmd(options Options) error {
	return dumpWtsXml(getDecoder(options.Dump.Remainder[0]))
}

func checkCmd(opt Options) error {
	return nil
}
