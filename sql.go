// 23 august 2012
package main

import (
	"fmt"
	"strings"
	"bytes"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
//	_ "github.com/Go-SQL-Driver/MySQL"
	"log"
	"sort"
	"unicode"
)

type SQL struct {
	db			*sql.DB
	getconsoles	*sql.Stmt
	getgames		*sql.Stmt
	getwikitext	*sql.Stmt
	getredirect	*sql.Stmt
	getcatlist		*sql.Stmt
}

var globsql *SQL

func NewSQL() *SQL {
	var err error

	s := new(SQL)

	s.db, err = sql.Open("mymysql",
		"tcp:" + config.DBServer + "*" +
			config.DBDatabase + "/" + config.DBUsername + "/" + config.DBPassword)
// for Go-SQL-Driver:
//	s.db, err = sql.Open("mysql",
//		config.DBUsername + ":" + config.DBPassword + "@" +
//			"tcp(" +  config.DBServer + ")/" + config.DBDatabase + "?charset=utf8")
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

	s.getredirect, err = s.db.Prepare(
		`SELECT rd_title
			FROM wiki_redirect
			WHERE rd_from = ?
				AND rd_interwiki = "";`)	// don't cross sites
	if err != nil {
		log.Fatalf("could not prepare redirect query (for scan list): %v", err)
	}

	s.getcatlist, err = s.db.Prepare(
		`SELECT wiki_categorylinks.cl_to
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_page.page_namespace = 6
				AND wiki_page.page_title = ?
				AND wiki_categorylinks.cl_from = wiki_page.page_id;`)
	if err != nil {
		log.Fatalf("could not prepare category list query (for checking a scan): %v", err)
	}

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

	gl, err := s.getconsoles.Query()
	if err != nil {
		return nil, fmt.Errorf("could not run console list query: %v", err)
	}
	defer gl.Close()

	for gl.Next() {
		var b []byte

		err = gl.Scan(&b)
		if err != nil {
			return nil, fmt.Errorf("error reading entry in console list query: %v", err)
		}
		// TODO save the string conversion for later? or do we even need to convert to string...?
		c := string(b)
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

	gl, err := s.getgames.Query(canonicalize(console))
	if err != nil {
		return nil, fmt.Errorf("could not run game list query: %v", err)
	}
	defer gl.Close()

	// use sql.RawBytes to avoid a copy since we're going to be converting to string anyway
	// TODO or do we even need to convert to string...?
	var b sql.RawBytes

	for gl.Next() {
		err = gl.Scan(&b)
		if err != nil {
			return nil, fmt.Errorf("error reading entry in game list query: %v", err)
		}
		games = append(games, string(b))
	}
	return games, nil
}

func sql_getwikitext(page string) ([]byte, error) {
	return globsql.GetWikitext(page)
}

// get wikitext, following all redirects
func (s *SQL) GetWikitext(page string) ([]byte, error) {
	var wikitext []byte			// TODO make into a sql.RawBytes and then produce a copy at the end? but see the next comment
	var nextTitle []byte			// this should be sql.RawBytes but apparently I can't do that with sql.Stmt.QueryRow()

	curTitle := canonicalize(page)
	for {
		var id uint32

		err := s.getwikitext.QueryRow(curTitle).Scan(&wikitext, &id)
		if err != nil {
			return nil, fmt.Errorf("error running or reading entry in wikitext query (for scan list): %v", err)
		}

		err = s.getredirect.QueryRow(id).Scan(&nextTitle)
		if err == sql.ErrNoRows {			// no redirect, so finished
			break
		} else if err != nil {
			return nil, fmt.Errorf("error running or reading entry in redirect result rows query (for scan list): %v", err)
		}
		// TODO do we even need to convert to string...?
		curTitle = string(nextTitle)		// not finished; follow redirect
	}
	return wikitext, nil
}

func isfileincategorywithprefix(file string, prefix []byte) (bool, error) {
	return globsql.IsFileInCategoryWithPrefix(file, prefix)
}

func (s *SQL) IsFileInCategoryWithPrefix(file string, prefix []byte) (bool, error) {
	cl, err := s.getcatlist.Query(canonicalize(file))
	if err != nil {
		return false, fmt.Errorf("could not run category list query (for checking a scan): %v", err)
	}
	defer cl.Close()

	// use sql.RawBytes to avoid a copy since we aren't storing the bytes, only checking against them
	var b sql.RawBytes

	for cl.Next() {
		err = cl.Scan(&b)
		if err != nil {
			return false, fmt.Errorf("error reading entry in category list query (for checking a scan): %v", err)
		}
		if bytes.HasPrefix(b, prefix) {
			return true, nil
		}
	}
	return false, nil			// nope
}
