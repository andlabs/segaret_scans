// 23 august 2012
package main

import (
	"fmt"
	"strings"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
	"sort"
	"unicode"
)

type SQL struct {
	db			mysql.Conn
	getconsoles	mysql.Stmt
	getgames		mysql.Stmt
	getwikitext	mysql.Stmt
	getredirect	mysql.Stmt
	getcatlist		mysql.Stmt

	// bound parameters
	category		string		// for getgames
	curTitle		string		// for getwikitext
	id			uint32		// for getredirect
	filename		string		// for getcatlist
}

var globsql *SQL

func NewSQL() *SQL {
	s := new(SQL)

	s.db = mysql.New("tcp", "", config.DBServer, config.DBUsername, config.DBPassword, config.DBDatabase)
	err := s.db.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	s.getconsoles, err = s.db.Prepare(
		`SELECT cat_title
			FROM wiki_category
			WHERE cat_title LIKE "%games"
				AND cat_pages > 0
			ORDER BY cat_title ASC;`)
	if err != nil {
		log.Fatalf("could not prepare console list query: %v", err)
	}

	s.getgames, err = s.db.Prepare(
		`SELECT wiki_page.page_title
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_categorylinks.cl_to = ?
				AND wiki_page.page_id = wiki_categorylinks.cl_from
				AND wiki_page.page_namespace = 0
			ORDER BY wiki_page.page_title ASC;`)
	if err != nil {
		log.Fatalf("could not prepare game list query: %v", err)
	}
	s.getgames.Bind(&s.category)

	s.getwikitext, err = s.db.Prepare(
		`SELECT wiki_text.old_text, wiki_page.page_id
			FROM wiki_page, wiki_revision, wiki_text
			WHERE wiki_page.page_namespace = 0
				AND wiki_page.page_title = ?
				AND wiki_page.page_latest = wiki_revision.rev_id
				AND wiki_revision.rev_text_id = wiki_text.old_id;`)
	if err != nil {
		log.Fatalf("could not prepare wikitext query (for scan list): %v", err)
	}
	s.getwikitext.Bind(&s.curTitle)

	s.getredirect, err = s.db.Prepare(
		`SELECT rd_title
			FROM wiki_redirect
			WHERE rd_from = ?
				AND rd_interwiki = "";`)	// don't cross sites
	if err != nil {
		log.Fatalf("could not prepare redirect query (for scan list): %v", err)
	}
	s.getredirect.Bind(&s.id)

	s.getcatlist, err = s.db.Prepare(
		`SELECT wiki_categorylinks.cl_to
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_page.page_namespace = 6
				AND wiki_page.page_title = ?
				AND wiki_categorylinks.cl_from = wiki_page.page_id;`)
	if err != nil {
		log.Fatalf("could not prepare category list query (for checking a scan): %v", err)
	}
	s.getcatlist.Bind(&s.filename)

	return s
}

func init() {
	addInit(func() {
		globsql = NewSQL()
	})
}

func canonicalize(pageName string) string {
	pageName = strings.Replace(pageName, " ", "_", -1)
	k := []rune(pageName)		// force first letter uppercase
	k[0] = unicode.ToUpper(k[0])
	return string(k)
}

func sql_getconsoles(filter func(string) bool) ([]string, error) {
	return globsql.GetConsoleList(filter)
}

func (s *SQL) GetConsoleList(filter func(string) bool) ([]string, error) {
	var consoles []string

	res, err := s.getconsoles.Run()
	if err != nil {
		return nil, fmt.Errorf("could not run console list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		return nil, fmt.Errorf("could not get console list result rows: %v", err)
	}
	for _, v := range gl {
		c := string(v[0].([]byte))
		// make human readable and drop _games
		c = strings.Replace(c, "_", " ", -1)
		c = c[:len(c) - len(" games")]
		if filter(c) {
			consoles = append(consoles, c)
		}
	}
	sort.Strings(consoles)
	return consoles, nil
}

func sql_getgames(console string) ([]string, error) {
	return globsql.GetGameList(console)
}

func (s *SQL) GetGameList(console string) ([]string, error) {
	var games []string

	s.category = canonicalize(console)
	res, err := s.getgames.Run()
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

func sql_getwikitext(page string) ([]byte, error) {
	return globsql.GetWikitext(page)
}

// get wikitext, following all redirects
func (s *SQL) GetWikitext(page string) ([]byte, error) {
	var wikitext []byte

	s.curTitle = canonicalize(page)
	for {
		res, err := s.getwikitext.Run()
		if err != nil {
			return nil, fmt.Errorf("could not run wikitext query (for scan list): %v", err)
		}
		wt, err := res.GetRows()
		if err != nil {
			return nil, fmt.Errorf("could not get wikitext result rows (for scan list): %v", err)
		}
		textField := res.Map("old_text")
		if textField < 0 {
			return nil, fmt.Errorf("could not locate page text (for scan list): %v", err)
		}
		wikitext = wt[0][textField].([]byte)
		idField := res.Map("page_id")
		if idField < 0 {
			return nil, fmt.Errorf("could not locate page id (for scan list): %v", err)
		}
		s.id = wt[0][idField].(uint32)
		redir_res, err := s.getredirect.Run()
		if err != nil {
			return nil, fmt.Errorf("could not get redirect result rows (for scan list): %v", err)
		}
		rd, err := redir_res.GetRows()
		if err != nil {
			return nil, fmt.Errorf("could not get redirect result rows (for scan list): %v", err)
		}
		if len(rd) == 0 {					// no redirect, so finished
			break
		}
		s.curTitle = string(rd[0][0].([]byte))	// not finished; follow redirect
	}
	return wikitext, nil
}

func sql_getcatlist(file string) ([]string, error) {
	return globsql.GetFileCategories(file)
}

func (s *SQL) GetFileCategories(file string) ([]string, error) {
	var categories []string

	s.filename = canonicalize(file)
	res, err := s.getcatlist.Run()
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
