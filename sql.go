// 23 august 2012
package main

import (
	"fmt"
	"strings"
	"os"
	"io/ioutil"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
)

const sqlport = "3306"

var db mysql.Conn
var getconsoles, getgames, getwikitext, getredirect, getcatlist mysql.Stmt

func init() {
	passwd_file, err := os.Open("/home/andlabs/src/segaret_scans/.passwd")
	if err != nil {
		log.Fatalf("could not get password")
	}
	passwd, err := ioutil.ReadAll(passwd_file)
	if err != nil {
		log.Fatalf("could not get password")
	}
	passwd_file.Close()

	db = mysql.New("tcp", "", "127.0.0.1:" + sqlport,
		"andlabs", string(passwd), "sonicret_sega")
	err = db.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	getconsoles, err = db.Prepare(
		`SELECT cat_title, cat_pages
			FROM wiki_category
			ORDER BY cat_title ASC;`)
	if err != nil {
		log.Fatalf("could not prepare console list query: %v", err)
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
		`SELECT wiki_text.old_text, wiki_page.page_id
			FROM wiki_page, wiki_revision, wiki_text
			WHERE wiki_page.page_namespace = 0
				AND wiki_page.page_title = ?
				AND wiki_page.page_latest = wiki_revision.rev_id
				AND wiki_revision.rev_text_id = wiki_text.old_id;`)
	if err != nil {
		log.Fatalf("could not prepare wikitext query (for scan list): %v", err)
	}

	getredirect, err = db.Prepare(
		`SELECT rd_title
			FROM wiki_redirect
			WHERE rd_from = ?
				AND rd_interwiki = "";`)	// don't cross sites
	if err != nil {
		log.Fatalf("could not prepare redirect query (for scan list): %v", err)
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

func sql_getconsoles() ([]string, []int32, error) {
	var consoles []string
	var nMembers []int32

	res, err := getconsoles.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("could not run console list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get console list result rows: %v", err)
	}
	nameField := res.Map("cat_title")
	if nameField < 0 {
		return nil, nil, fmt.Errorf("could not locate console names: %v", err)
	}
	countField := res.Map("cat_pages")
	if countField < 0 {
		return nil, nil, fmt.Errorf("could not locate console game count: %v", err)
	}
	for _, v := range gl {
		c := string(v[nameField].([]byte))
		if strings.HasSuffix(c, "_games") {
			// make human readable and drop _games
			c = strings.Replace(c, "_", " ", -1)
			consoles = append(consoles, c[:len(c) - len(" games")])
			nMembers = append(nMembers, v[countField].(int32))
		}
	}
	return consoles, nMembers, nil
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

// get wikitext, following all redirects
func sql_getwikitext(page string) (string, error) {
	var wikitext string

	curTitle := canonicalize(page)
	for {
		res, err := getwikitext.Run(curTitle)
		if err != nil {
			return "", fmt.Errorf("could not run wikitext query (for scan list): %v", err)
		}
		wt, err := res.GetRows()
		if err != nil {
			return "", fmt.Errorf("could not get wikitext result rows (for scan list): %v", err)
		}
		textField := res.Map("old_text")
		if textField < 0 {
			return "", fmt.Errorf("could not locate page text (for scan list): %v", err)
		}
		wikitext = string(wt[0][textField].([]byte))
		idField := res.Map("page_id")
		if idField < 0 {
			return "", fmt.Errorf("could not locate page id (for scan list): %v", err)
		}
		id := wt[0][idField].(uint32)
		redir_res, err := getredirect.Run(id)
		if err != nil {
			return "", fmt.Errorf("could not get redirect result rows (for scan list): %v", err)
		}
		rd, err := redir_res.GetRows()
		if err != nil {
			return "", fmt.Errorf("could not get redirect result rows (for scan list): %v", err)
		}
		if len(rd) == 0 {					// no redirect, so finished
			break
		}
		curTitle = string(rd[0][0].([]byte))	// not finished; follow redirect
	}
	return wikitext, nil
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
