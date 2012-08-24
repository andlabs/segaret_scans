// 22-23 august 2012
package main

import (
	"strings"
	"regexp"

	// overall test
//	"fmt"
//	"encoding/xml"
)

type ScanboxParam struct {
	Name	string
	Value	string
}

type Scanbox []ScanboxParam

var nowikiStartTag, nowikiEndTag,
	preStartTag, preEndTag,
	htmlStartTag, htmlEndTag	*regexp.Regexp
var scanboxStart *regexp.Regexp

func init() {
	const endStartTag = "([ \t\n]+[^>]*)?>"
	const endEndTag = "[ \t\n]*>"

	nowikiStartTag = regexp.MustCompile("<[nN][oO][wW][iI][kK][iI]" + endStartTag)
	nowikiEndTag = regexp.MustCompile("</[nN][oO][wW][iI][kK][iI]" + endEndTag)
	preStartTag = regexp.MustCompile("<[pP][rR][eE]" + endStartTag)
	preEndTag = regexp.MustCompile("</[pP][rR][eE]" + endEndTag)
	htmlStartTag = regexp.MustCompile("<[hH][tT][mM][lL]" + endStartTag)
	htmlEndTag = regexp.MustCompile("</[hH][tT][mM][lL]" + endEndTag)
	scanboxStart = regexp.MustCompile(`\{\{[ \t\n]*[Ss]canbox`)
}

/*
This is a dumb parser. It does only the first step of parsing (http://www.mediawiki.org/wiki/Markup_spec/BNF/Nowiki) before looking for templates. It will not handle recursive template definitions (which should not happen in ScanBox anyway). The handling of links with alternate labels ([[abc|def]]) is rudimentary. It does not handle situations where two |s appear in a row in a template (like {{Scanbox | a=b | | c=d}}).

	Some people, when confronted with a problem, think
	"I know, I'll use regular expressions."
	Now they have two problems. - jwz
use of regexps for the <nowiki>/<pre>/<html> tags suggested by f2f on #go-nuts
use of regexps for {{Scanbox was my own creation, to just get something off the ground
*/

// strip text between given markers
func stripLiteral(wikitext string, start *regexp.Regexp, end *regexp.Regexp) (out string) {
	var loc []int

top:
	loc = start.FindStringIndex(wikitext)
	if loc != nil {			// match?
		goto strip
	}
	out = out + wikitext		// add what's left
	return
strip:
	out = out + wikitext[:loc[0]]
	wikitext = wikitext[loc[1]:]
	loc = end.FindStringIndex(wikitext)
	if loc != nil {			// match?
		goto endstrip
	}
	return				// assume end at EOF if no match
endstrip:
	wikitext = wikitext[loc[1]:]
	goto top

	panic("unreachable")	// please the compiler
}
/* test:
func main() {
	nowiki := func(s string) string {
		return stripLiteral(s, nowikiStartTag, nowikiEndTag)
	}
	pre := func(s string) string {
		return stripLiteral(s, preStartTag, preEndTag)
	}
	html := func(s string) string {
		return stripLiteral(s, htmlStartTag, htmlEndTag)
	}
	fmt.Println(nowiki("hello<nowiki>dear</nowiki> world"))	// expected: hello world
	fmt.Println(nowiki("<nowiki>abcdefg</nowiki>"))			// expected: [blank]
	fmt.Println(pre("<pre>a</pre>b<pre>c</pre>"))			// expected: b
	fmt.Println(html("nothing"))							// expected: nothing
	fmt.Println(html("<html>something</html> else"))			// expected:  else
}
*/

// strip comments, returning the number of comment stripped
func stripComments(wikitext string) (out string, n int) {
	var i int

top:
	for i = 0; i < len(wikitext); i++ {
		if strings.HasPrefix(wikitext[i:], "<!--") {
			goto strip
		}
	}
	out = out + wikitext			// add what's left
	return
strip:
	n++
	out = out + wikitext[:i]
	wikitext = wikitext[i + 4:]		// skip <!--
	for i = 0; i < len(wikitext); i++ {
		if strings.HasPrefix(wikitext[i:], "-->") {
			goto endstrip
		}
	}
	return					// unclosed comment (TODO really return?)
endstrip:
	wikitext = wikitext[i + 3:]		// skip -->
	goto top

	panic("unreachable")		// please the compiler
}
/* test:
func stripall(wikitext string) string {
	for i := 1; i != 0; {
		wikitext, i = stripComments(wikitext)
	}
	return wikitext
}
func main() {
	fmt.Println(stripall("hello"))				// expected: hello
	fmt.Println(stripall("<!-- comment -->"))		// expected: [blank]
	fmt.Println(stripall("abc<!--d-->efg"))		// expected: abcefg
	fmt.Println(stripall("<!<!---->---->"))		// expected: [blank]
	fmt.Println(stripall("<!--<!---->-->"))		// expected: -->
}
*/

func getScanboxAt(wikitext string) (t Scanbox) {
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
	key := ""
	for ; i < len(wikitext); i++ {
		c := wikitext[i]
		if c == ' ' || c == '\t' || c == '\n' {
			continue
		}
		if c == '=' {
			i++
			goto getvalue
		}
		key = key + string(c)
	}
	panic("key without value or unterminated template")
getvalue:
	value := []byte{}
	inLink := false
	for ; i < len(wikitext); i++ {		// don't eat whitespace here; it's crucial (we will tream leading and trailing whitespace later)
		c := wikitext[i]
		if c == '|' && !inLink {
			goto store
		}
		if c == '}' {
			goto store
		}
		if c == '[' {
			inLink = true
		}
		if c == ']' && inLink {
			inLink = false
		}
		value = append(value, c)
	}
	panic("unterminated template")
store:
	if inLink {
		panic("unterminated link")
	}
	t = append(t, ScanboxParam{
		Name:	key,
		Value:	strings.TrimSpace(string(value)),
	})
	goto top

	panic("unreachable")		// please the compiler
}

func GetScanboxes(wikitext string) (list []Scanbox) {
	wikitext = stripLiteral(wikitext, nowikiStartTag, nowikiEndTag)
	wikitext = stripLiteral(wikitext, preStartTag, preEndTag)
	wikitext = stripLiteral(wikitext, htmlStartTag, htmlEndTag)
	for n := 1; n != 0; {		// we have to recursively strip comments... seriously
		wikitext, n = stripComments(wikitext)
	}

	allScanboxes := scanboxStart.FindAllStringIndex(wikitext, -1)
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
