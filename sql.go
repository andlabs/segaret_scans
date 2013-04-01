// 23 august 2012
package main

import (
	"fmt"
	"strings"
	"bytes"
	"database/sql"
	_ "github.com/ziutek/mymysql/godrv"
//	_ "github.com/go-sql-driver/mysql"
	"log"
	"unicode"
)

type SQL struct {
	db			*sql.DB
	getgames		*sql.Stmt
	getcatlist		*sql.Stmt
	db_scanbox	*sql.DB		// TODO do I need a separate one?
	getscanboxes	*sql.Stmt
	getnoscans	*sql.Stmt
}

var globsql *SQL

func opendb(which string) (*sql.DB, error) {
	return sql.Open("mymysql",
		"tcp:" + config.DBServer + "*" +
			which + "/" + config.DBUsername + "/" + config.DBPassword)
// for Go-SQL-Driver:
//	return sql.Open("mysql",
//		config.DBUsername + ":" + config.DBPassword + "@" +
//			"tcp(" +  config.DBServer + ")/" + which + "?charset=utf8")
}

func NewSQL() *SQL {
	var err error

	s := new(SQL)

	s.db, err = opendb(config.DBDatabase)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
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

	s.getcatlist, err = s.db.Prepare(
		`SELECT wiki_categorylinks.cl_to
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_page.page_namespace = 6
				AND wiki_page.page_title = ?
				AND wiki_categorylinks.cl_from = wiki_page.page_id;`)
	if err != nil {
		log.Fatalf("could not prepare category list query (for checking a scan): %v", err)
	}

	s.db_scanbox, err = opendb(config.DBScanboxDatabase)
	if err != nil {
		log.Fatalf("could not connect to scanbox database: %v", err)
	}

	s.getscanboxes, err = s.db_scanbox.Prepare(
		`SELECT _page, console, region, cover, front, back, spine, spinemissing, square, spinecard, cart, disc, disk, manual, jewelcase, jewelcasefront, jewelcaseback, jewelcasespine, jewelcasespinemissing, item1, item2, item3, item4, item5, item6, item7, item8, item1name, item2name, item3name, item4name, item5name, item6name, item7name, item8name, spine2, top, bottom
			FROM Scanbox;`)
	if err != nil {
		log.Fatalf("could not prepare scanbox list query: %v", err)
	}

	s.getnoscans, err = s.db_scanbox.Prepare(
		`SELECT COUNT(*)
			FROM NoScans
			WHERE _page = ?
				AND console = ?;`)
	if err != nil {
		log.Fatalf("could not prepare noscans list query: %v", err)
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

// arguments to bytes.Replace() must be []byte
var (
	byteUnderscore = []byte("_")
	byteSpace = []byte(" ")
)

func decanonicalize(pageName sql.RawBytes) sql.RawBytes {
	return bytes.Replace(pageName, byteUnderscore, byteSpace, -1)
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
		b = decanonicalize(b)
		games = append(games, string(b))
	}
	return games, nil
}

func sql_getscanboxes() ([]*Scan, error) {
	return globsql.GetScanboxes()
}

const nScanboxFields = 38

func nsToString(_n interface{}) string {
	n := _n.(*sql.NullString)
	if n.Valid {
		return n.String
	}
	return ""
}

// get scanboxes
func (s *SQL) GetScanboxes() ([]*Scan, error) {
	scanboxes := make([]*Scan, 0)
	sbl, err := s.getscanboxes.Query()
	if err != nil {
		return nil, fmt.Errorf("could not run scanbox list query (for scan list): %v", err)
	}
	defer sbl.Close()

	// I cannot expand a slice into a variadic argument list so here goes complexity!
	sbf := make([]interface{}, nScanboxFields)
	for i := 0; i < len(sbf); i++ {
		sbf[i] = new(sql.NullString)
	}

	for sbl.Next() {
		var s Scan

		err := sbl.Scan(sbf...)
		if err != nil {
			return nil, fmt.Errorf("error reading entry in scanbox list query (for scan list): %v", err)
		}
		i := 0
		s.Name = nsToString(sbf[i]); i++
		s.Console = nsToString(sbf[i]); i++
		s.Region = nsToString(sbf[i]); i++
		s.Cover = nsToString(sbf[i]); i++
		s.Front = nsToString(sbf[i]); i++
		s.Back = nsToString(sbf[i]); i++
		s.Spine = nsToString(sbf[i]); i++
		s.DBSpineMissing = nsToString(sbf[i]); i++
		s.DBSquare = nsToString(sbf[i]); i++
		s.SpineCard = nsToString(sbf[i]); i++
		s.Cart = nsToString(sbf[i]); i++
		s.Disc = nsToString(sbf[i]); i++
		s.Disk = nsToString(sbf[i]); i++
		s.Manual = nsToString(sbf[i]); i++
		s.JewelCase = nsToString(sbf[i]); i++
		s.JewelCaseFront = nsToString(sbf[i]); i++
		s.JewelCaseBack = nsToString(sbf[i]); i++
		s.JewelCaseSpine = nsToString(sbf[i]); i++
		s.DBJCSM = nsToString(sbf[i]); i++
		s.Item1 = nsToString(sbf[i]); i++
		s.Item2 = nsToString(sbf[i]); i++
		s.Item3 = nsToString(sbf[i]); i++
		s.Item4 = nsToString(sbf[i]); i++
		s.Item5 = nsToString(sbf[i]); i++
		s.Item6 = nsToString(sbf[i]); i++
		s.Item7 = nsToString(sbf[i]); i++
		s.Item8 = nsToString(sbf[i]); i++
		s.Item1name = nsToString(sbf[i]); i++
		s.Item2name = nsToString(sbf[i]); i++
		s.Item3name = nsToString(sbf[i]); i++
		s.Item4name = nsToString(sbf[i]); i++
		s.Item5name = nsToString(sbf[i]); i++
		s.Item6name = nsToString(sbf[i]); i++
		s.Item7name = nsToString(sbf[i]); i++
		s.Item8name = nsToString(sbf[i]); i++
		s.Spine2 = nsToString(sbf[i]); i++
		s.Top = nsToString(sbf[i]); i++
		s.Bottom = nsToString(sbf[i]); i++
		scanboxes = append(scanboxes, &s)
	}

	return scanboxes, nil
}

func sql_getmarkednoscans(game string, console string) (bool, error) {
	return globsql.GetMarkedNoScans(game, console)
}

func (s *SQL) GetMarkedNoScans(game string, console string) (bool, error) {
	var n int

	err := s.getnoscans.QueryRow(game, console).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("could not run noscans list query (for scan list): %v", err)
	}
	
	if n == 0 {
		return false, nil
	}
	if n == 1 {
		return true, nil
	}
	return false, fmt.Errorf("sanity check fail: game %s console %s listed either more than once or negative times in NoScans table (listed %d times)", game, console, n)
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
