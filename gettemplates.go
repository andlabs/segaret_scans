// 22-23 august 2012
package main

import (
	"strings"
	"regexp"
)

type TParam struct {
	Name	string
	Value	string
}

type Template struct {
	Name	string
	Params	[]TParam
}

var nowikiStartTag, nowikiEndTag,
	preStartTag, preEndTag,
	htmlStartTag, htmlEndTag	*regexp.Regexp

func init() {
	const endStartTag = "([ \t\n]+[^>]*)?>"
	const endEndTag = "[ \t\n]*>"

	nowikiStartTag = regexp.MustCompile("<[nN][oO][wW][iI][kK][iI]" + endStartTag)
	nowikiEndTag = regexp.MustCompile("</[nN][oO][wW][iI][kK][iI]" + endEndTag)
	preStartTag = regexp.MustCompile("<[pP][rR][eE]" + endStartTag)
	preEndTag = regexp.MustCompile("</[pP][rR][eE]" + endEndTag)
	htmlStartTag = regexp.MustCompile("<[hH][tT][mM][lL]" + endStartTag)
	htmlEndTag = regexp.MustCompile("</[hH][tT][mM][lL]" + endEndTag)
}

/*
This is a dumb parser. It does only the following steps:
	- ???
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

func GetTemplates(_wikitext string) (list []Template) {
	var curtemplate Template
	var curstring string

	wikitext := []rune(_wikitext)
//	wikitext = stripLiteral(wikitext, nowikiStartTag, nowikiEndTag)
//	wikitext = stripLiteral(wikitext, "pre")
//	wikitext = stripLiteral(wikitext, "html")
	for n := 1; n != 0; {		// we have to recursively strip comments... seriously
		wikitext, n = stripComments(wikitext)
	}
