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
	"strings"
)

import (
	"github.com/voxelbrain/goptions"
)

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"unicode/utf16"

	//	"os"
)

type Xml struct {
	Xml string `xml:",innerxml"`
}

type Comment struct {
	Comment string `xml:"CommentText,attr"`
}

type ContextParameter struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:"Value,attr"`
}

type DataSource struct {
	Name       string `xml:"Name,attr"`
	Connection string `xml:"Connection,attr"`
	Tables     struct {
		DataSourceTable string `xml:",innerxml"`
	}
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

type ValidationRules struct {
	ValidationRule []struct {
		XmlBase
	}
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

	ValidationRules ValidationRules
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

type current struct {
	transaction string
	comment     string
}

func treatWtsXml(w io.Writer, checkOnly bool, decoder *xml.Decoder) error {

	inloop := false
	var cur current
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
					cur.comment = c.Comment
					treatComment(w, cur.comment)
				}
			case "ContextParameter":
				{
					var r ContextParameter
					decoder.DecodeElement(&r, &t)
					fmt.Fprintf(w, "CP: %s=%s\r\n", r.Name, r.Value)
				}
			case "DataSource":
				{
					var r DataSource
					decoder.DecodeElement(&r, &t)
					fmt.Fprintf(w, "DS: (%s, %s) %s\r\n", r.Name, r.Connection, minify(r.Tables.DataSourceTable))
				}
			case "ConditionalRule":
				{
					var r ConditionalRule
					decoder.DecodeElement(&r, &t)
					// The ConditionalRule might be under Condition or Loop
					if inloop {
						fmt.Fprintf(w, "\r\n<=\r\nLP")
						inloop = false
					} else {
						fmt.Fprintf(w, "\r\n<=\r\nCB")
					}
					fmt.Fprintf(w, ": (%s) %s\r\n", r.Name, minify(r.RuleParameters.Xml))
				}
			case "IncludedWebTest":
				{
					var r IncludedWebTest
					decoder.DecodeElement(&r, &t)
					fmt.Fprintf(w, "I: %s\r\n", r.Included)
				}
			case "Loop":
				inloop = true
			case "Request":
				treatRequest(w, checkOnly, decoder, t, cur)

			case "TransactionTimer":
				// <TransactionTimer Name="the transaction name">
				cur.transaction = t.Attr[0].Value
				treatTransaction(w, cur.transaction)
			case "ValidationRules":
				{
					var r ValidationRules
					decoder.DecodeElement(&r, &t)
					for _, v := range r.ValidationRule {
						fmt.Fprintf(w, "VR: (%s) %s\r\n", v.Name, minify(v.RuleParameters.Xml))
					}
				}
			}

		case xml.EndElement:
			switch t.Name.Local {
			case "Condition":
				fmt.Fprintf(w, "CE: \r\n=>\r\n\r\n")
			case "Loop":
				fmt.Fprintf(w, "LP: \r\n=>\r\n\r\n")
			}

		default:
		}
	}

	return nil
}

func treatComment(w io.Writer, v string) {
	fmt.Fprintf(w, "C: %s\r\n", v)

}

// treatRequest will process requests like
// <Request Method="GET" or <Request Method="POST"
func treatRequest(wi io.Writer, checkOnly bool,
	decoder *xml.Decoder, t xml.StartElement, cur current) {
	w := bytes.NewBuffer([]byte{})

	switch t.Attr[0].Value {
	case "GET":
		{
			var r GetRequest
			decoder.DecodeElement(&r, &t)
			//fmt.Fprintf(w,"R: %q\r\n", r)
			fmt.Fprintf(w, "G: (%s,%s) %s\r\n", r.ThinkTime, r.Timeout, r.Url)
			dealReqAddons(w, r.Request)
			checkRequest(checkOnly, r.Request, w, cur)
		}
	case "POST":
		{
			var r PostRequest
			decoder.DecodeElement(&r, &t)
			//fmt.Fprintf(w,"R: %q\r\n", r)
			fmt.Fprintf(w, "P: (%s,%s) %s\r\n", r.ThinkTime, r.Timeout, r.Url)
			fmt.Fprintf(w, "  B: %s\r\n", DecodeStringBody(r.StringBody))
			dealReqAddons(w, r.Request)
			checkRequest(checkOnly, r.Request, w, cur)
		}
	default:
		panic("Internal error parsing Request")
	}

	if !checkOnly {
		wi.Write(w.Bytes())
	}
}

func treatTransaction(w io.Writer, v string) {
	fmt.Fprintf(w, "\r\nT: %s\r\n", v)
}

func dealReqAddons(w io.Writer, r Request) {
	if len(r.RequestPlugins.RequestPlugin) != 0 {
		for _, v := range r.RequestPlugins.RequestPlugin {
			fmt.Fprintf(w, "  R: (%s) %s\r\n", v.Name, minify(v.RuleParameters.Xml))
		}
	}
	if len(r.ExtractionRules.ExtractionRule) != 0 {
		for _, v := range r.ExtractionRules.ExtractionRule {
			fmt.Fprintf(w, "  E: (%s: %s) %s\r\n", v.Name, v.VariableName, minify(v.RuleParameters.Xml))
		}
	}
	if len(r.ValidationRules.ValidationRule) != 0 {
		for _, v := range r.ValidationRules.ValidationRule {
			fmt.Fprintf(w, "  V: (%s) %s\r\n", v.Name, minify(v.RuleParameters.Xml))
		}
	}
	w.Write([]byte("\r\n"))
}

func checkRequest(checkOnly bool, r Request, buf *bytes.Buffer, cur current) {
	if !checkOnly {
		return
	}
	reqs := buf.String()
	tt, _ := strconv.Atoi(r.ThinkTime)
	to, _ := strconv.Atoi(r.Timeout)
	if tt != options.Check.ThinkTime ||
		to != options.Check.Timeout ||
		checkRe.MatchString(reqs) {
		treatTransaction(os.Stdout, cur.transaction)
		treatComment(os.Stdout, cur.comment)
		fmt.Printf(reqs)
	}
}

func DecodeStringBody(s string) string {
	uDec, _ := base64.StdEncoding.DecodeString(s)
	return DecodeUTF16(uDec)
}

func DecodeUTF16(s []byte) string {
	u16s := make([]uint16, len(s)/2)
	for i := range u16s {
		u16s[i] = binary.LittleEndian.Uint16([]byte(s[i*2:]))
	}

	return string(utf16.Decode(u16s))
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
		Fileo *os.File `goptions:"-o, --output, description='The web test script dump output', wronly"`
		Cnr   string   `goptions:"-c, --cnr, description='Comment number removal, for easy comparison'"`
	} `goptions:"dump"`
}

var options = Options{ // Default values goes here
	Check: Check{
		Checks:    `/*20\d\d-*`,
		ThinkTime: 0,
		Timeout:   270,
	},
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
	//fmt.Printf("] %#v\r\n", options)

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

	fileo := options.Dump.Fileo
	if fileo == nil {
		var err error
		fileo, err = os.Create(
			strings.Replace(options.Dump.Filei.Name(), ".webtest", ".webtext", 1))
		check(err)
	}
	defer fileo.Close()

	return treatWtsXml(fileo, false, getDecoder(options.Dump.Filei))
}

var checkRe *regexp.Regexp

func checkCmd(opt Options) error {
	checkRe = regexp.MustCompile(options.Check.Checks)
	//fmt.Printf("] %#v %#v\r\n", options.Check.Checks, checkRe)

	return treatWtsXml(ioutil.Discard, true, getDecoder(options.Check.Filei))
}
