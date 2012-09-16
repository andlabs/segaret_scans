// 22-23 august 2012
package main

import (
	"regexp"
	"strings"		// TrimSpace at last step of key/value pair parsing
	"flag"

	// overall test
//	"fmt"
//	"encoding/xml"
)

var stripStuff = flag.Bool("strip", false, "strip <nowiki>, <pre>, <html> and HTML comments (will dramatically increase processing time, be warned)")

type ScanboxParam struct {
	Name	string
	Value	string
}

type Scanbox []ScanboxParam

var scanboxStart, noScansStart *regexp.Regexp

func getscanboxes_init() {
	scanboxStart = regexp.MustCompile(`\{\{[ \t\n]*[Ss]canbox`)
	noScansStart = regexp.MustCompile(`\{\{[ \t\n]*[Nn]o[Ss]cans`)
}

func init() {
	addInit(getscanboxes_init)
}

/*
This is a dumb parser. It does only the first step of parsing (http://www.mediawiki.org/wiki/Markup_spec/BNF/Nowiki) before looking for templates. It will not handle recursive template definitions (which should not happen in ScanBox anyway). The handling of links with alternate labels ([[abc|def]]) or nested templates is rudimentary. It does not handle situations where two |s appear in a row in a template (like {{Scanbox | a=b | | c=d}}).

	Some people, when confronted with a problem, think
	"I know, I'll use regular expressions."
	Now they have two problems. - jwz
use of regexps for the <nowiki>/<pre>/<html> tags suggested by f2f on #go-nuts
use of regexps for {{Scanbox was my own creation, to just get something off the ground
*/

func getScanboxAt(wikitext []byte) (t Scanbox) {
	i := 0
top:
	for ; i < len(wikitext); i++ {
		c := wikitext[i]
		if c == ' ' || c == '\t' || c == '\n' {
			continue
		}
		if c == '|' {
			i++
			goto beginkv
		}
		if c == '}' {			// end at }}
			if i + 1 < len(wikitext) && wikitext[i + 1] == '}' {
				return
			}			// else break
		}
		break
	}
	panic("unexpected input (expected | or }}) or unfinished template")
beginkv:
	key := make([]byte, 0, 128)
	for ; i < len(wikitext); i++ {
		c := wikitext[i]
		if c == ' ' || c == '\t' || c == '\n' {
			continue
		}
		if c == '=' {
			i++
			goto getvalue
		}
		key = append(key, c)
	}
	panic("key without value or unterminated template")
getvalue:
	value := make([]byte, 0, 128)
	inLink := 0
	for ; i < len(wikitext); i++ {		// don't eat whitespace here; it's crucial (we will tream leading and trailing whitespace later)
		c := wikitext[i]
		if c == '|' && inLink == 0 {
			goto store
		}
		if c == '}' && inLink == 0 {
			goto store
		}
		if c == '[' || c == '{' {
			inLink++
		}
		if (c == ']' || c == '}') && inLink != 0 {
			inLink--
		}
		value = append(value, c)
	}
	panic("unterminated template")
store:
	if inLink != 0 {
		panic("unterminated link")
	}
	// give the result in a uniform manner
	// both names and values are case sensitive (for values, if the value is a filename)
	t = append(t, ScanboxParam{
		Name:	strings.TrimSpace(string(key)),
		Value:	strings.TrimSpace(string(value)),
	})
	goto top

	panic("unreachable")		// please the compiler
}

func GetScanboxes(wikitext []byte, consoleNone string) (list []Scanbox, none bool) {
	if *stripStuff {
		wikitext = stripwikitext(wikitext)
	}

	// check to see if this version of the game has no scans
	allNoScans := noScansStart.FindAllIndex(wikitext, -1)
	if len(allNoScans) != 0 {
		for _, v := range allNoScans {
			k := getScanboxAt(wikitext[v[1]:])
			for _, param := range k {
				if param.Name == "console" &&
					strings.EqualFold(param.Value, consoleNone) {
					none := true
					return nil, none
				}
			}
		}
	}

	allScanboxes := scanboxStart.FindAllIndex(wikitext, -1)
	if len(allScanboxes) == 0 {
		return
	}
	for _, v := range allScanboxes {
		list = append(list, getScanboxAt(wikitext[v[1]:]))
	}
	return
}

/*
// overall test
func main() {
	r, err := getWikiAPIData("/api.php?action=query&prop=revisions&rvprop=content&format=xml&titles=Thunder%20Force%20IV")
	if err != nil {
		fmt.Printf("error retrieving game Thunder Force IV: %v\n", err)
		return
	}

	var dat struct {
		X	string	`xml:"query>pages>page>revisions>rev"`
	}

	err = xml.Unmarshal(r, &dat)
	if err != nil {
		fmt.Printf("error processing games: %v\ndata: %s\n", err, r)
		return
	}

	fmt.Printf("%#v\n", GetScanboxes(dat.X))
}
*/
