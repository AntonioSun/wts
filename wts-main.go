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
	"encoding/base64"
	"encoding/xml"
	"io"
	"io/ioutil"
	"regexp"
	//	"os"
)

type Xml struct {
	Xml string `xml:",innerxml"`
}

type Comment struct {
	Comment string `xml:"CommentText,attr"`
}

type XmlBase struct {
	Name           string `xml:"DisplayName,attr"`
	RuleParameters Xml
}

type ConditionalRule struct {
	XmlBase
}

type IncludedWebTest struct {
	Included string `xml:"Name,attr"`
}

type Request struct {
	Url       string `xml:"Url,attr"`
	ThinkTime string `xml:"ThinkTime,attr"`
	Timeout   string `xml:"Timeout,attr"`

	RequestPlugins struct {
		RequestPlugin []struct {
			XmlBase
		}
	}

	ExtractionRules struct {
		ExtractionRule []struct {
			XmlBase
			VariableName string `xml:"VariableName,attr"`
		}
	}

	ValidationRules struct {
		ValidationRule []struct {
			XmlBase
		}
	}
}

type GetRequest struct {
	Request
}

type PostRequest struct {
	Request
	StringBody string `xml:"StringHttpBody"`
}

func getDecoder(Script *os.File) *xml.Decoder {
	defer Script.Close()

	content, err := ioutil.ReadFile(Script.Name())
	check(err)
	return xml.NewDecoder(bytes.NewBuffer(content))
}

func dumpWtsXml(decoder *xml.Decoder, w io.Writer) error {

	inloop := false
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
			switch inElement := t.Name.Local; inElement {
			case "Comment":
				{
					var c Comment
					// decode a whole chunk of following XML into the
					// variable c which is a Comment (t above)
					decoder.DecodeElement(&c, &t)
					fmt.Fprintf(w, "C: %s\n", c.Comment)
				}
			case "ConditionalRule":
				{
					var r ConditionalRule
					decoder.DecodeElement(&r, &t)
					// The ConditionalRule might be under Condition or Loop
					if inloop {
						fmt.Fprintf(w, "\n<=\nLP")
						inloop = false
					} else {
						fmt.Fprintf(w, "\n<=\nCB")
					}
					fmt.Fprintf(w, ": (%s) %s\n", r.Name, minify(r.RuleParameters.Xml))
				}
			case "IncludedWebTest":
				{
					var r IncludedWebTest
					decoder.DecodeElement(&r, &t)
					fmt.Fprintf(w, "I: %s\n", r.Included)
				}
			case "Loop":
				inloop = true
			case "Request":
				// <Request Method="GET" or <Request Method="POST"
				//fmt.Fprintf(w,"R: %q, %q\n", t.Attr)
				switch t.Attr[0].Value {
				case "GET":
					{
						var r GetRequest
						decoder.DecodeElement(&r, &t)
						//fmt.Fprintf(w,"R: %q\n", r)
						fmt.Fprintf(w, "G: (%s,%s) %s\n", r.ThinkTime, r.Timeout, r.Url)
						fmt.Fprintf(w, getReqAddons(r.Request))
					}
				case "POST":
					{
						var r PostRequest
						decoder.DecodeElement(&r, &t)
						//fmt.Fprintf(w,"R: %q\n", r)
						fmt.Fprintf(w, "P: (%s,%s) %s\n", r.ThinkTime, r.Timeout, r.Url)
						uDec, _ := base64.StdEncoding.DecodeString(r.StringBody)
						fmt.Fprintf(w, "  B: %s\n", string(uDec))
						fmt.Fprintf(w, getReqAddons(r.Request))
					}
				default:
					panic("Internal error parsing Request")
				}
			case "TransactionTimer":
				// <TransactionTimer Name="the transaction name">
				fmt.Fprintf(w, "\nT: %s\n", t.Attr[0].Value)
			}

		case xml.EndElement:
			switch t.Name.Local {
			case "Condition":
				fmt.Fprintf(w, "CE: \n=>\n\n")
			case "Loop":
				fmt.Fprintf(w, "LP: \n=>\n\n")
			}

		default:
		}
	}

	return nil
}

func getReqAddons(r Request) string {
	ret := ""
	if len(r.RequestPlugins.RequestPlugin) != 0 {
		for _, v := range r.RequestPlugins.RequestPlugin {
			ret += fmt.Sprintf("  R: (%s) %s\n", v.Name, minify(v.RuleParameters.Xml))
		}
	}
	if len(r.ExtractionRules.ExtractionRule) != 0 {
		for _, v := range r.ExtractionRules.ExtractionRule {
			ret += fmt.Sprintf("  E: (%s: %s) %s\n", v.Name, v.VariableName, minify(v.RuleParameters.Xml))
		}
	}
	if len(r.ValidationRules.ValidationRule) != 0 {
		for _, v := range r.ValidationRules.ValidationRule {
			ret += fmt.Sprintf("  V: (%s) %s\n", v.Name, minify(v.RuleParameters.Xml))
		}
	}
	return ret + "\n"
}

func minify(xs string) string {
	re := regexp.MustCompile("\r*\n *")
	return re.ReplaceAllString(xs, "")
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
		Filei *os.File `goptions:"-i, --input, obligatory, description='The web test script to work on', rdonly"`
		Fileo *os.File `goptions:"-o, --output, description='The web test script dump output', wronly"`
		Cnr   string   `goptions:"-c, --cnr, mutexgroup='input', description='Comment number removal, for easy comparison'"`
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

	//fmt.Printf("] F: %#v %#v\n", options.Dump.Filei, options.Dump.Fileo)

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

	return dumpWtsXml(getDecoder(options.Dump.Filei), os.Stdout) // options.Dump.Fileo)
}

func checkCmd(opt Options) error {
	return nil
}
