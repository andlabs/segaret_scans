// 22 august 2012
package main

import (
	"fmt"
	"strings"
)

type Scan struct {
	Console				string
	Region				string
	Front				string
	Back					string
	Spine				string
	SpineMissing			bool
	Square				bool
	SpineCard				string
	Cart					string
	Disc					string
	Manual				string
	HasJewelCase			bool
	JewelCaseFront			string
	JewelCaseBack			string
	JewelCaseSpine		string
	JewelCaseSpineMissing	bool
	Items				[]string
}

var ErrGameNoScans = fmt.Errorf("game has no scans")

func GetScans(game string, consoleNone string) ([]Scan, error) {
	var scans []Scan

	wikitext, err := sql_getwikitext(game)
	if err != nil {
		return nil, fmt.Errorf("error retrieving game %s: %v", game, err)
	}
	scanboxes, none := GetScanboxes(wikitext, consoleNone)
	if none {
		return nil, ErrGameNoScans
	}
	for _, v := range scanboxes {
		var s Scan
		var items, itemnames [8]string

		for _, p := range v {
			pname := strings.ToLower(strings.TrimSpace(p.Name))
			pvalue := strings.TrimSpace(p.Value)
			switch pname {
			case "console":
				s.Console = pvalue
			case "region":
				// strip <br> and <br/>; otherwise filtering by region (which is case-insensitive) will match those too; replace with a space to look good
				pvalue = strings.Replace(pvalue, "<br>", " ", -1)
				pvalue = strings.Replace(pvalue, "<br/>", " ", -1)
				s.Region = pvalue
			case "front":
				s.Front = pvalue
			case "back":
				s.Back = pvalue
			case "spine":
				s.Spine = pvalue
			case "spinemissing":
				s.SpineMissing = (pvalue == "yes")
			case "square":
				s.Square = (pvalue == "yes")
			case "spinecard":
				s.SpineCard = pvalue
			case "cart":
				s.Cart = pvalue
			case "disc", "disk":
				s.Disc = pvalue
			case "manual":
				s.Manual = pvalue
			case "jewelcase":
				s.HasJewelCase = (pvalue == "yes")
			case "jewelcasefront":
				s.JewelCaseFront = pvalue
			case "jewelcaseback":
				s.JewelCaseBack = pvalue
			case "jewelcasespine":
				s.JewelCaseSpine = pvalue
			case "jewelcasespinemissing":
				s.JewelCaseSpineMissing = (pvalue == "yes")
			case "item1", "item2", "item3", "item4",
				"item5", "item6", "item7", "item8":
				items[pname[4] - '0' - 1] = pvalue
			case "item1name", "item2name", "item3name", "item4name",
				"item5name", "item6name", "item7name", "item8name":
				itemnames[pname[4] - '0' - 1] = pvalue
			case "spine2":
				// ignore
				// TODO what to do about spine2?
			default:	// top* and bottom* too?
				if !strings.HasPrefix(pname, "top") &&
					!strings.HasPrefix(pname, "bottom") {
					return nil, fmt.Errorf("unknown parameter %s=%s", pname, pvalue)
				}
			}
		}
		// handle extra items
		for i := 0; i < len(items); i++ {
			if items[i] == "" && itemnames[i] == "" {		// item unspecified
				continue
			}
			s.Items = append(s.Items, items[i])
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
