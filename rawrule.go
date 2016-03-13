////////////////////////////////////////////////////////////////////////////
// Porgram: rawrule.go
// Purpose: raw-rule handling for wts dump
// authors: Antonio Sun (c) 2016, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

import (
	"gopkg.in/yaml.v2"
)

type RawRule struct {
	Replace yaml.MapSlice
}

type replRuleT struct {
	rexp *regexp.Regexp
	repl string
}

//var replRule []replRuleT

func rawRuleRead(filename string) {

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		if VERBOSITY > 0 {
			fmt.Printf("%s-rawrule:\n  %v\n", progname, err)
			fmt.Printf("%s-rawrule: skip using the .rawrule file\n", progname)
		}
		return
	}
	//debug(string(source), 1)

	var rawRule RawRule
	err = yaml.Unmarshal(source, &rawRule)
	check(err)

	//fmt.Printf("] %+v\n", rawRule)
	//replRule = make([]replRuleT, len(rawRule.Replace))
	for _, av := range rawRule.Replace {
		//replRule[ix].rexp = regexp.MustCompile(av.Key.(string))
		//replRule[ix].repl = av.Value.(string)
		stringBodyFix.ApplyRegexpReplaceAll(av.Key.(string), av.Value.(string))
	}
	//fmt.Printf("] %+v\n", replRule)
}
