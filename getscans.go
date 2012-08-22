// 22 august 2012
package main

import (
	"fmt"
	"encoding/xml"
	"net/url"
	"strings"
)

type Scan struct {
	Region		string
	Front		string
	Back			string
	Spine		string
	SpineMissing	bool
	SpineCard		string
	Cart			string
	Disc			string
	Manual		string
}

type tParam struct {
	Name	string	`xml:"name"`
	Value	string	`xml:"value"`
}

type tTemplate struct {
	Name	string	`xml:"title"`
	Params	[]tParam	`xml:"part"`
}

// the parse tree itself is XML, so we can just unmarshal right out of that
type gp_parsetree struct {
	ParseTree		[]byte		`xml:"parsetree,attr"`
}

type gamepage struct {
	ParseTree		gp_parsetree	`xml:"query>pages>page>revisions>rev"`
	Templates	[]tTemplate	`xml:"template"`
}

func urlForGame(game string) string {
	return "http://segaretro.org/api.php?format=xml&action=query&titles=" + url.QueryEscape(game) + "&prop=revisions&rvparse&rvgeneratexml&rvprop=content"
}

func getScans(game string) ([]Scan, error) {
	var scans []Scan
	var gp gamepage

	r, err := getWikiAPIData(urlForGame(game))
	if err != nil {
		return nil, fmt.Errorf("error retrieving game %s: %v", game, err)
	}
	err = xml.Unmarshal(r, &gp)
	if err != nil {
		return nil, fmt.Errorf("error processing games: %v\ndata: %s", err, r)
	}
	err = xml.Unmarshal(gp.ParseTree.ParseTree, &gp)
	if err != nil {
		return nil, fmt.Errorf("error processing templates: %v\ndata: %s", err, gp.ParseTree.ParseTree)
	}
	for _, v := range gp.Templates {
		// have to trim because MediaWiki likes to keep newlines
		if strings.ToLower(strings.TrimSpace(v.Name)) == "scanbox" {
			var s Scan

			for _, p := range v.Params {
				pname := strings.ToLower(strings.TrimSpace(p.Name))
				pvalue := strings.TrimSpace(p.Value)
				switch pname {
				case "region":
					s.Region = pvalue
				case "front":
					s.Front = pvalue
				case "back":
					s.Back = pvalue
				case "spine":
					s.Spine = pvalue
				case "spinemissing":
					s.SpineMissing = (pvalue == "yes")
				case "spinecard":
					s.SpineCard = pvalue
				case "cart":
					s.Cart = pvalue
				case "disc", "disk":
					s.Disc = pvalue
				case "manual":
					s.Manual = pvalue
				case "console", "square", "spine2":
					// ignore
					// TODO what to do about spine2?
				default:	// ignore item* and jewelcase*
					if (len(pname) > 4 && pname[:4] != "item") &&
						(len(p.Name) >= 9 && pname[:9] != "jewelcase") {
						return nil, fmt.Errorf("unknown parameter %s=%s", pname, pvalue)
					}
				}
			}
			scans = append(scans, s)
		}
	}
	return scans, err
}

// test
func main() {
	scans, err := getScans("Thunder Force IV")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	for _, v := range scans {
		fmt.Printf("%#v\n", v)
	}
}
