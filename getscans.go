// 22 august 2012
package main

import (
	"fmt"
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

func GetScans(game string) ([]Scan, error) {
	var scans []Scan

	wikitext, err := sql_getwikitext(game)
	if err != nil {
		return nil, fmt.Errorf("error retrieving game %s: %v", game, err)
	}
	scanboxes := GetScanboxes(wikitext)
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
