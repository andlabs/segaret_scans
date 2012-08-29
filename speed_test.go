// 27 august 2012
package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// console list
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