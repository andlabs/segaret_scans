// 23 august 2012
package main

import (
	"fmt"
	"strings"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
)

const sqlport = "3306"

var db mysql.Conn
var getgames, getwikitext, getcatlist mysql.Stmt

func init() {
	db = mysql.New("tcp", "", "127.0.0.1:" + sqlport,
		"andlabs", "[redacted from repository]", "sonicret_sega")
	err := db.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	getgames, err = db.Prepare(
		`SELECT wiki_page.page_title
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_categorylinks.cl_to = ?
				AND wiki_page.page_id = wiki_categorylinks.cl_from
				AND wiki_page.page_namespace = 0
			ORDER BY wiki_page.page_title ASC;`)
	if err != nil {
		log.Fatalf("could not prepare game list query: %v", err)
	}

	getwikitext, err = db.Prepare(
		`SELECT wiki_text.old_text
			FROM wiki_page, wiki_revision, wiki_text
			WHERE wiki_page.page_namespace = 0
				AND wiki_page.page_title = ?
				AND wiki_page.page_latest = wiki_revision.rev_id
				AND wiki_revision.rev_text_id = wiki_text.old_id;`)
	if err != nil {
		log.Fatalf("could not prepare wikitext query (for scan list): %v", err)
	}

	getcatlist, err = db.Prepare(
		`SELECT wiki_categorylinks.cl_to
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_page.page_namespace = 6
				AND wiki_page.page_title = ?
				AND wiki_categorylinks.cl_from = wiki_page.page_id;`)
	if err != nil {
		log.Fatalf("could not prepare category list query (for checking a scan): %v", err)
	}
}

// TODO see if mediawiki has a better definition
func canonicalize(pageName string) string {
	return strings.Replace(pageName, " ", "_", -1)
}

func sql_getgames(console string) ([]string, error) {
	var games []string

	category := canonicalize(console) + "_games"
	res, err := getgames.Run(category)
	if err != nil {
		return nil, fmt.Errorf("could not run game list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		return nil, fmt.Errorf("could not get game list result rows: %v", err)
	}
	for _, v := range gl {
		games = append(games, string(v[0].([]byte)))
	}
	return games, nil
}

func sql_getwikitext(page string) (string, error) {
	res, err := getwikitext.Run(canonicalize(page))
	if err != nil {
		return "", fmt.Errorf("could not run wikitext query (for scan list): %v", err)
	}
	wt, err := res.GetRows()
	if err != nil {
		return "", fmt.Errorf("could not get wikitext result rows (for scan list): %v", err)
	}
	return string(wt[0][0].([]byte)), nil
}

func sql_getcatlist(file string) ([]string, error) {
	var categories []string

	res, err := getcatlist.Run(canonicalize(file))
	if err != nil {
		return nil, fmt.Errorf("could not run category list query (For checking a scan): %v", err)
	}
	cl, err := res.GetRows()
	if err != nil {
		return nil, fmt.Errorf("could not get category list result rows (for checking a scan): %v", err)
	}
	for _, v := range cl {
		categories = append(categories, string(v[0].([]byte)))
	}
	return categories, nil
}

/*
func main() {
	games, err := sql_getgames("Mega Drive")
	if err != nil {
		fmt.Printf("error grabbing game list: %v\n", err)
	} else {
		fmt.Println(strings.Join(games, "\n"))
	}

	wt, err := sql_getwikitext("Thunder_Force_IV")
	if err != nil {
		fmt.Printf("error grabbing wikitext: %v\n", err)
	} else {
		fmt.Println(wt)
	}

	categories, err := sql_getcatlist("ThunderForce4 MD JP Box.jpg")
	if err != nil {
		fmt.Printf("error grabbing category list: %v\n", err)
	} else {
		fmt.Println(strings.Join(categories, "\n"))
	}
}
*/
