// 22 august 2012
package main

import (
	"fmt"
	"encoding/xml"
	"net/url"
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

func getScanList(game string) ([]Scan, error) {
	var scans []Scan
	var gp gamepage
	var template tTemplate

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
		fmt.Println("{{" + v.Name)
		for _, p := range v.Params {
			fmt.Println("| " + p.Name + "=" + p.Value)
		}
		fmt.Println("}}")
	}
_=template;_=scans;	return nil, err
}

func main() { _,b:=getScanList("Thunder Force IV");fmt.Println(b) }
