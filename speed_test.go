// 27 august 2012
package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"

	"strings"
	"sort"
)

func BenchmarkSpeed(b *testing.B) {
	b.StopTimer()
	w := httptest.NewRecorder()
	r := &http.Request{
		URL:		&url.URL{
			Path:		"/scans/",
		},
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getConsoleInfo(w, r)
	}
}

func BenchmarkGetConsoleList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := sql_getconsoles()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func _filter(s string) bool {
	return !strings.HasPrefix(s, "19") &&			// omit years
			!strings.HasPrefix(s, "20") &&
			!strings.HasSuffix(s, " action") &&		// omit genres
			!strings.HasSuffix(s, " adventure") &&	// TODO probably best to use a regexp
			!strings.HasSuffix(s, " educational") &&
			!strings.HasSuffix(s, " fighting") &&
			!strings.HasSuffix(s, " puzzle") &&
			!strings.HasSuffix(s, " racing") &&
			!strings.HasSuffix(s, " shoot-'em-up") &&
			!strings.HasSuffix(s, " shooting") &&
			!strings.HasSuffix(s, " simulation") &&
			!strings.HasSuffix(s, " sports") &&
			!strings.HasSuffix(s, " table") &&
			!strings.HasPrefix(s, "Unlicensed ") &&	// omit qualifiers
			!strings.HasPrefix(s, "Unreleased ") &&
			!strings.HasPrefix(s, "3D ") &&
			!strings.HasPrefix(s, "Big box ") &&
			!strings.HasPrefix(s, "US ") &&
			!strings.HasPrefix(s, "EU ") &&
			!strings.HasPrefix(s, "JP ") &&
			!strings.HasPrefix(s, "Homebrew ") &&
			!omitConsoles[s]
}

func BenchmarkGetConsoleList_Like(b *testing.B) {
	b.StopTimer()
	getconsoles, err := db.Prepare(
//		`SELECT cat_title, cat_pages
		`SELECT cat_title
			FROM wiki_category
			WHERE cat_title LIKE "%games"
				AND cat_pages > 0;`)
//			WHERE cat_title LIKE "%games";`)
//			ORDER BY cat_title ASC;`)
	if err != nil {
		b.Fatalf("could not prepare console list query: %v", err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
	var consoles []string
//	var nMembers []int32

	res, err := getconsoles.Run()
	if err != nil {
		b.Fatalf("could not run console list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		b.Fatalf("could not get console list result rows: %v", err)
	}
	nameField := res.Map("cat_title")
	if nameField < 0 {
		b.Fatalf("could not locate console names: %v", err)
	}
//	countField := res.Map("cat_pages")
//	if countField < 0 {
//		b.Fatalf("could not locate console game count: %v", err)
//	}
	for _, v := range gl {
		c := string(v[nameField].([]byte))
//		if strings.HasSuffix(c, "_games") {
			// make human readable and drop _games
			c = strings.Replace(c, "_", " ", -1)
			c = c[:len(c) - len(" games")]
			if _filter(c) {
				consoles = append(consoles, c)
			}
//			nMembers = append(nMembers, v[countField].(int32))
//		}
	}
//	_ = sort.Strings
	sort.Strings(consoles)
	}
}

// game list

const console = "Mega Drive"

func BenchmarkGetWikitext(b *testing.B) {
	b.StopTimer()
	games, err := GetGameList(console)
	if err != nil {
		b.Fatalf("error getting %s game list: %v", console, err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _, game := range games {
			_, err := sql_getwikitext(game)
			if err != nil {
				b.Fatalf("error retrieving game %s: %v", game, err)
			}
		}
	}
}

func BenchmarkGetWikitext_N(b *testing.B) {
	b.StopTimer()
	getgames, err := db.Prepare(
		`SELECT wiki_page.page_title
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_categorylinks.cl_to = ?
				AND wiki_page.page_id = wiki_categorylinks.cl_from
				AND wiki_page.page_namespace = 0
			ORDER BY wiki_page.page_title ASC;`)
	if err != nil {
		b.Fatalf("could not prepare game list query: %v", err)
	}
	getwikitext, err = db.Prepare(
		`SELECT wiki_text.old_text
			FROM wiki_page, wiki_revision, wiki_text
			WHERE wiki_page.page_namespace = 0
				AND wiki_page.page_title = ?
				AND wiki_page.page_latest = wiki_revision.rev_id
				AND wiki_revision.rev_text_id = wiki_text.old_id;`)
	if err != nil {
		b.Fatalf("could not prepare wikitext query (for scan list): %v", err)
	}
	var games [][]byte
	category := "Mega_Drive_games"
	res, err := getgames.Run(category)
	if err != nil {
		b.Fatalf("could not run game list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		b.Fatalf("could not get game list result rows: %v", err)
	}
	for _, v := range gl {
		games = append(games, v[0].([]byte))
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for _, game := range games {
			res, err := getwikitext.Run(game)
			if err != nil {
				b.Fatalf("could not run wikitext query (for scan list): %v", err)
			}
			wt, err := res.GetRows()
			if err != nil {
				b.Fatalf("could not get wikitext result rows (for scan list): %v", err)
			}
			_ = string(wt[0][0].([]byte))
		}
	}
}

func BenchmarkGetWikitext_ID(b *testing.B) {
	b.StopTimer()
	getgames, err := db.Prepare(
		`SELECT wiki_page.page_id
			FROM wiki_page, wiki_categorylinks
			WHERE wiki_categorylinks.cl_to = ?
				AND wiki_page.page_id = wiki_categorylinks.cl_from
				AND wiki_page.page_namespace = 0
			ORDER BY wiki_page.page_title ASC;`)
	if err != nil {
		b.Fatalf("could not prepare game list query: %v", err)
	}
	getwikitext, err = db.Prepare(
		`SELECT wiki_text.old_text
			FROM wiki_page, wiki_revision, wiki_text
			WHERE wiki_page.page_namespace = 0
				AND wiki_page.page_id = ?
				AND wiki_page.page_latest = wiki_revision.rev_id
				AND wiki_revision.rev_text_id = wiki_text.old_id;`)
	if err != nil {
		b.Fatalf("could not prepare wikitext query (for scan list): %v", err)
	}
	var games []uint32
	category := "Mega_Drive_games"
	res, err := getgames.Run(category)
	if err != nil {
		b.Fatalf("could not run game list query: %v", err)
	}
	gl, err := res.GetRows()
	if err != nil {
		b.Fatalf("could not get game list result rows: %v", err)
	}
	for _, v := range gl {
		games = append(games, v[0].(uint32))
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for _, game := range games {
			res, err := getwikitext.Run(game)
			if err != nil {
				b.Fatalf("could not run wikitext query (for scan list): %v", err)
			}
			wt, err := res.GetRows()
			if err != nil {
				b.Fatalf("could not get wikitext result rows (for scan list): %v", err)
			}
			_ = string(wt[0][0].([]byte))
		}
	}
}