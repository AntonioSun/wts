////////////////////////////////////////////////////////////////////////////
// Porgram: wts-dump
// Purpose: wts (web test script) dump handling
// authors: Antonio Sun (c) 2015-16, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

import (
	"github.com/AntonioSun/shaper"
)

////////////////////////////////////////////////////////////////////////////
// Extending shaper.Shaper
type Shaper struct {
	*shaper.Shaper
}

// Make a new Shaper filter and start adding bits
func NewFilter() *Shaper {
	//return &Shaper{ShaperStack: PassThrough}
	return &Shaper{Shaper: shaper.NewFilter()}
}

func (me *Shaper) ApplyXMLDecode() *Shaper {
	me.AddFilter(html.UnescapeString)
	return me
}

func (sp *Shaper) GetDFSID() *Shaper {
	sp.AddFilter(func(s string) string {
		dfSIDRe := regexp.MustCompile(`<SessionTicket>(.*?)</SessionTicket>`)
		dfSID := dfSIDRe.FindStringSubmatch(s)
		if len(dfSID) < 2 {
			return ""
		}
		return dfSID[1]
	})
	return sp
}

func (sp *Shaper) GetDFCBID() *Shaper {
	sp.AddFilter(func(s string) string {
		re := regexp.MustCompile(`force/u/(.*?)/`)
		id := re.FindStringSubmatch(s)
		if len(id) < 2 {
			return ""
		}
		//debug(id[1], 1)
		return id[1]
	})
	return sp
}

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

////////////////////////////////////////////////////////////////////////////
// Function definitions

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

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// Script-wide processing

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
					fmt.Fprintf(w, "DS: (%s, %s) %s\r\n",
						r.Name, r.Connection, minify.Process(r.Tables.DataSourceTable))
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
					fmt.Fprintf(w, ": (%s) %s\r\n",
						r.Name, minify.Process(r.RuleParameters.Xml))
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
						fmt.Fprintf(w, "VR: (%s) %s\r\n",
							v.Name, minify.Process(v.RuleParameters.Xml))
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

	if options.Dump.Tsr {
		// list of all date time strings used in the script in sorted order
		var keys []string
		for k := range dateCol {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Fprintf(w, "TS: %s: %d\n", k, dateCol[k])
		}
	}

	return nil
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// Item-level processing

func treatComment(w io.Writer, v string) {
	if options.Dump.Cnr {
		v = cmtRe.ReplaceAllString(v, "[]")
	}
	fmt.Fprintf(w, "C: %s\r\n", v)
}

var DFClientBrowserId = ""
var DFClientBrowserIdFound = false
var dfCBIDReplace = shaper.NewFilter()
var urlFix = shaper.NewFilter().ApplyRegexpReplaceAll(
	`http.*\w+\.\w+\.com`, "{{Param_TestServer}}")

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
			if options.Dump.Raw {
				r.ThinkTime = "0"
				r.Timeout = "0"
				if len(DFClientBrowserId) == 0 {
					DFClientBrowserId = NewFilter().GetDFCBID().Process(r.Url)
				}
				if len(DFClientBrowserId) != 0 {
					if !DFClientBrowserIdFound {
						DFClientBrowserIdFound = true
						urlFix.ApplyReplace(DFClientBrowserId,
							"{{Param_ClientBrowserId}}", -1)
					}
				}
				r.Url = urlFix.Process(r.Url)
			}
			//fmt.Fprintf(w,"R: %q\r\n", r)
			fmt.Fprintf(w, "G: (%s,%s) %s (%s):%s\r\n",
				r.ThinkTime, r.Timeout, r.Url, r.ReportingName, r.RecordResult)
			dealReqAddons(w, r.Request)
			checkRequest(checkOnly, r.Request, w, cur)
		}
	case "POST":
		{
			var r PostRequest
			decoder.DecodeElement(&r, &t)
			stringBody := DecodeStringBody(r.StringBody)
			coreService := ""
			if options.Dump.Raw {
				r.ThinkTime = "0"
				r.Timeout = "0"
				r.Url = urlFix.Process(r.Url)
				if len(r.StringBody) != 0 {
					regReplace := shaper.NewFilter().ApplyRegexpReplaceAll(
						`.*(Get)</ReadableRequestName><RequestName>(.*?)</RequestName>.*&lt;MethodName&gt;(.*?)&lt;/MethodName&gt;.*`,
						"$1.$2.$3").ApplyRegexpReplaceAll(
						`.*<ReadableCorrelator>|</ReadableCorrelator>.*`,
						"").ApplyRegexpReplaceAll(
						`.*<ReadableRequestName>|</ReadableRequestName>.*`, "")
					coreService = regReplace.Process(stringBody)
				}
			}
			//fmt.Fprintf(w,"R: %q\r\n", r)
			fmt.Fprintf(w, "P: (%s,%s) %s %s (%s):%s\r\n", r.ThinkTime, r.Timeout,
				r.Url, coreService, r.ReportingName, r.RecordResult)
			if len(r.StringBody) != 0 {
				fmt.Fprintf(w, "%s\r\n",
					dealRequest(stringBodyDump.Process(stringBody)))
			}
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

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// Request specific processing

func dealReqAddons(w io.Writer, r Request) {
	if len(r.QueryStringParameters.QueryStringParameter) != 0 {
		prefixTag := "  Q: "
		split := minify.Copy()
		split.ApplyRegexpReplaceAll(`( />)(<)`, "$1\r\n"+prefixTag+"$2")
		fmt.Fprintf(w, "%s%s\r\n", prefixTag,
			split.Process(r.QueryStringParameters.QueryStringParameter))
	}
	if len(r.FormPostHttpBody.FormPostParameter) != 0 {
		fmt.Fprintf(w, "  F: %s\r\n",
			minify.Process(r.FormPostHttpBody.FormPostParameter))
	}
	if len(r.RequestPlugins.RequestPlugin) != 0 {
		for _, v := range r.RequestPlugins.RequestPlugin {
			fmt.Fprintf(w, "  R: (%s) %s\r\n",
				v.Name, minify.Process(v.RuleParameters.Xml))
		}
	}
	if len(r.ExtractionRules.ExtractionRule) != 0 {
		for _, v := range r.ExtractionRules.ExtractionRule {
			fmt.Fprintf(w, "  E: (%s: %s) %s\r\n",
				v.Name, v.VariableName, minify.Process(v.RuleParameters.Xml))
		}
	}
	if len(r.ValidationRules.ValidationRule) != 0 {
		for _, v := range r.ValidationRules.ValidationRule {
			fmt.Fprintf(w, "  V: (%s) %s\r\n",
				v.Name, minify.Process(v.RuleParameters.Xml))
		}
	}
	w.Write([]byte("\r\n"))
}

var DFSessionTicket = ""
var DFSessionTicketFound = false
var dfSidReplace = shaper.NewFilter()

// date string collection
var dateCol map[string]int

func init() {
	dateCol = make(map[string]int)
}

// dealRequest
// a filter to deal with POST StringBody and GET QueryStringParameters
// Functionality:
//   - collect and replace date strings
func dealRequest(v string) string {
	// Get the 1st DFSessionTicket
	if options.Dump.Raw {
		if len(DFSessionTicket) == 0 {
			DFSessionTicket = NewFilter().GetDFSID().Process(v)
			// debug(DFSessionTicket, 1)
		}
		if len(DFSessionTicket) != 0 {
			if !DFSessionTicketFound {
				DFSessionTicketFound = true
				dfSidReplace.ApplyReplace(DFSessionTicket, "{{Param_SessionId}}", -1)
			}
			v = dfSidReplace.Process(v)
		}
		v = stringBodyFix.Process(v)
	}

	if !options.Dump.Tsr {
		return v
	}

	for _, m := range tmsRe.FindAllString(v, -1) {
		//debug(m, 1)
		dateCol[m]++
	}
	v = tmsRe.ReplaceAllString(v, "-time-string-")
	return v
}

func checkRequest(checkOnly bool, r Request, buf *bytes.Buffer, cur current) {
	if !checkOnly {
		return
	}
	reqs := buf.String()
	tt, _ := strconv.Atoi(r.ThinkTime)
	to, _ := strconv.Atoi(r.Timeout)
	if tt != options.Check.ThinkTime ||
		to < options.Check.Timeout ||
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

var minify *shaper.Shaper

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// Main dispatch functions

var cmtRe *regexp.Regexp
var tmsRe *regexp.Regexp

var stringBodyDump *Shaper
var stringBodyFix *shaper.Shaper

func dumpCmd() error {

	fileo := options.Dump.Fileo
	if fileo == nil {
		var err error
		fileo, err = os.Create(
			strings.Replace(options.Dump.Filei.Name(), ".webtest", ".webtext", 1))
		check(err)
	}
	defer fileo.Close()

	stringBodyDump = NewFilter()
	stringBodyFix = shaper.NewFilter()

	if !options.Dump.Asis {
		stringBodyDump.ApplyXMLDecode()
	}
	if options.Dump.Raw {
		options.Dump.Cnr = true
		rawRuleRead(
			strings.Replace(options.Dump.Filei.Name(), ".webtest", ".rawrule", 1))
	}
	if options.Dump.Cnr {
		cmtRe = regexp.MustCompile(`\[#\d+]`)
	}
	if options.Dump.Tsr {
		tmsRe = regexp.MustCompile(`(20\d{2}-\d{1,2}-\d{1,2}[T0-9:.]*|\d{1,2}/\d{1,2}/20\d{2})`)
	}
	minify = shaper.NewFilter().ApplyRegexpReplaceAll("\r*\n *", "")
	return treatWtsXml(fileo, false, getDecoder(options.Dump.Filei))
}

var checkRe *regexp.Regexp

func checkCmd() error {
	checkRe = regexp.MustCompile(options.Check.Checks)
	//fmt.Printf("] %#v %#v\r\n", options.Check.Checks, checkRe)

	return treatWtsXml(ioutil.Discard, true, getDecoder(options.Check.Filei))
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// Support functions

func check(e error) {
	if e != nil {
		fmt.Printf("%s error: ", progname)
		panic(e)
	}
}

func debug(input string, threshold int) {
	if !(VERBOSITY >= threshold) {
		return
	}
	print("] ")
	print(input)
	print("\n")
}
