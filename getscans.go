// 22 august 2012
package main

import (
	"fmt"
	"strings"
)

type Scan struct {
	Console				string
	Region				string
	Cover				string
	Front				string
	Back					string
	Spine				string
	SpineMissing			bool		`db:"zzzzzz"`		// we use this
	DBSpineMissing		string	`db:"spinemissing"`	// our database code uses this
	Square				bool		`db:"xxxxx"`	// we use this
	DBSquare				string	`db:"square"`	// our database code uses this
	SpineCard				string	// TODO we will need to explicitly use a struct tag if sqlx changes to not merely strings.ToLower() the field name
	Cart					string
	Disc					string
	Disk					string		// our database code uses this; we merge it with Disc
	Manual				string
	HasJewelCase			bool			// we use this
	JewelCase				string		// our database code uses this
	JewelCaseFront			string
	JewelCaseBack			string
	JewelCaseSpine		string
	JewelCaseSpineMissing	bool		`db:"yyyyyy"`				// we use this
	DBJCSM				string	`db:"jewelcasespinemissing"`	// our database code uses this
	Items				[]string		// we use this
	Item1				string		// our database code uses these
	Item2				string
	Item3				string
	Item4				string
	Item5				string
	Item6				string
	Item7				string
	Item8				string
	Item1name			string
	Item2name			string
	Item3name			string
	Item4name			string
	Item5name			string
	Item6name			string
	Item7name			string
	Item8name			string		// and that should cover it
	Spine2				string
	Top					string
	Bottom				string

	// we do not use these but sqlx complains if I do not provide them
	DBKey		int		`db:"__key"`
	DBPage		string	`db:"_page"`
	DBTimestamp	int		`db:"_timestamp"`
	Topbottomwidth	string
	Topmarginleft		string
	Bottommarginleft	string
}

var ErrGameNoScans = fmt.Errorf("game has no scans")

func GetScans(game string, consoleNone string) ([]*Scan, error) {
	scans, none, err := sql_getscanboxes(game, consoleNone)
	if err != nil {
		return nil, fmt.Errorf("error retrieving game %s: %v", game, err)
	}
	if none {
		return nil, ErrGameNoScans
	}
	// now to fine-tune the scanboxes
	for _, s := range scans {
		// 1) Region
		// strip <br> and <br/>; otherwise filtering by region (which is case-insensitive) will match those too; replace with a space to look good
		s.Region = strings.Replace(s.Region, "<br>", " ", -1)
		s.Region = strings.Replace(s.Region, "<br/>", " ", -1)

		// 2) boolean switches
		s.SpineMissing = (s.DBSpineMissing == "yes")
		s.Square = (s.DBSquare == "yes")
		s.HasJewelCase = (s.JewelCase == "yes")
		s.JewelCaseSpineMissing = (s.DBJCSM == "yes")

		// 3) disc and disk
		switch {
		case s.Disc != "" && s.Disk != "":
			return nil, fmt.Errorf("game %s console %s region %s has both disc (%s) and disk (%s)", game, s.Console, s.Region, s.Disc, s.Disk)
		case s.Disc != "":
			// do nothing since Disk is ""
		case s.Disk != "":
			s.Disc = s.Disk
		}
		// otherwise neither is defined so do nothing

		// 4) extra items
		add := func(item string, name string) {
			if item == "" && name == "" {		// item unspecified
				return
			}
			s.Items = append(s.Items, item)
		}
		add(s.Item1, s.Item1name)
		add(s.Item2, s.Item2name)
		add(s.Item3, s.Item3name)
		add(s.Item4, s.Item4name)
		add(s.Item5, s.Item5name)
		add(s.Item6, s.Item6name)
		add(s.Item7, s.Item7name)
		add(s.Item8, s.Item8name)
	}
	return scans, nil
}

// kept as a note to myself
//			// these parameters are related to displaying the top and bottom and should thus be ignored
//			case "topbottomwidth", "topmarginleft", "bottommarginleft":
//				// ignore

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
