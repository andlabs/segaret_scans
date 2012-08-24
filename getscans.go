// 22 august 2012
package main

import (
	"fmt"
	"encoding/xml"
	"net/url"
	"strings"
)

type Scan struct {
	Console		string
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

type gamepage struct {
	Source		string	`xml:"query>pages>page>revisions>rev"`
}

func urlForGame(game string) string {
//	return "/api.php?format=xml&action=query&titles=" + url.QueryEscape(game) + "&prop=revisions&rvparse&rvgeneratexml&rvprop=content"
	return "/api.php?action=query&prop=revisions&rvprop=content&format=xml&titles=" + url.QueryEscape(game)
}

func GetScans(game string) ([]Scan, error) {
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
	scanboxes := GetScanboxes(gp.Source)
	for _, v := range scanboxes {
		var s Scan

		for _, p := range v {
			pname := strings.ToLower(strings.TrimSpace(p.Name))
			pvalue := strings.TrimSpace(p.Value)
			switch pname {
			case "console":
				s.Console = pvalue
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
			case "square", "spine2":
				// ignore
				// TODO what to do about spine2?
			default:	// ignore item* and jewelcase*... top* and bottom* too?
				if !strings.HasPrefix(pname, "item") &&
					!strings.HasPrefix(pname, "jewelcase") &&
					!strings.HasPrefix(pname, "top") &&
					!strings.HasPrefix(pname, "bottom") {
					return nil, fmt.Errorf("unknown parameter %s=%s", pname, pvalue)
				}
			}
		}
		scans = append(scans, s)
	}
	return scans, err
}

/*
// test
func main() {
//	scans, err := GetScans("Thunder Force IV")
//	scans, err := GetScans("Light Crusader")
//	scans, err := GetScans("Crusader of Centy")
	scans, err := GetScans("The Lucky Dime Caper Starring Donald Duck")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	for _, v := range scans {
		fmt.Printf("%#v\n", v)
		fmt.Printf("box scan state: %v\n", v.BoxScanState())
		fmt.Printf("cart scan state: %v\n", v.CartScanState())
	}
}
*/
