// 27 august 2012
package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"

	"strings"
	"sort"
	"fmt"
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
//		_, _, err := sql_getconsoles()
		_, err := sql_getconsoles(_filter)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func _filter(s string) bool {
	return !strings.HasPrefix(s, "19") &&			// omit years
			!strings.HasPrefix(s, "20") &&
			!strings.HasSuffix(s, " action") &&		// omit genres
			!strings.HasSuffix(s, " adventure") &&
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

func BenchmarkConsolePage_NewCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		url, _ := url.Parse("http://andlabs.sonicretro.org/scans/Mega Drive")
		generateConsoleReport(console,
			httptest.NewRecorder(),
			*url)
	}
}

func BenchmarkConsolePage_OldCode(b *testing.B) {
	b.StopTimer()

var top = `<html>
<head>
	<title>Sega Retro Scan Information: %s</title>
	<style type="text/css">
	.Bad {
		background-color: #888800;
	}
	.Missing {
		background-color: #880000;
	}
	.Incomplete {
		background-color: #888800;
	}
	.Good {
		background-color: #008800;
	}
	.Error {
		background-color: #000000;
		color: #FFFFFF;
	}
	</style>
</head>
<body>
	<h1>Sega Retro Scan Information: %s</h1>
`

var beginTable = `
	<table>
		<tr>
			<th><a href="%s">Game</a></th>
			<th><a href="%s">Region</a></th>
			<th><a href="%s">Box</a></th>
			<th><a href="%s">Media</a></th>
		</tr>
`

var gameStart = `
		<tr>
			<td><a href="http://segaretro.org/%s">%s</a></td>
`

var gameEntry = `
			<td>%s</td>
			<td class=%v>%v</td>
			<td class=%v>%v</td>
		</tr>
`

var gameError = `
			<td colspan=3 class=Error>Error: %v</td>
		</tr>
`

var gameNoScans = `
			<td colspan=3 class=Missing>No scans</td>
		</tr>
`

var endTable = `
	</table>
`

const filterRegionName = "region"
const sortOrderName = "sort"
var sortOrders = map[string]SortOrder{
	"region":		SortByRegion,
	"box":		SortByBoxState,
	"media":		SortByMediaState,
}

urlSort := func(url url.URL, order string) string {
	q := url.Query()
	q.Del("sort")
	q.Add("sort", order)
	url.RawQuery = q.Encode()
	return url.String()
}

urlNoSort := func (url url.URL) string {
	q := url.Query()
	q.Del("sort")
	url.RawQuery = q.Encode()
	return url.String()
}

b.StartTimer()
for i := 0; i < b.N; i++ {
_url, _ := url.Parse("http://andlabs.sonicretro.org/scans/Mega Drive")
url := *_url
w := httptest.NewRecorder()
	var filterRegion string

	fmt.Fprintf(w, top, console, console)
	scans, err := GetConsoleScans(console)
	if err != nil {
		fmt.Fprintf(w, "<p>Error getting %s scan info: %v</p>\n", console, err)
		return
	}
	query := url.Query()
	if x, ok := query[filterRegionName]; ok && len(x) > 0 {	// filter by region if supplied
		filterRegion = x[0]
	}
	if x, ok := query[sortOrderName]; ok && len(x) > 0 {		// sort differently if asked
		if so, ok := sortOrders[x[0]]; ok {				// but only if we passed a valid sort order
			scans.Sort(so)
		}
	}
	stats := scans.GetStats(filterRegion)
	fmt.Fprintf(w, "%s\n<br>", stats.HTML())
	fmt.Fprintf(w, beginTable,
		urlNoSort(url), urlSort(url, "region"),
		urlSort(url, "box"), urlSort(url, "media"))
	for _, scan := range scans {
		if scan.Error != nil {
			fmt.Fprintf(w, gameStart, scan.Name, scan.Name)
			fmt.Fprintf(w, gameError, scan.Error)
			continue
		}
		if scan.HasNoScans {
			fmt.Fprintf(w, gameStart, scan.Name, scan.Name)
			fmt.Fprintf(w, gameNoScans)
			continue
		}
		if filterRegion != "" &&		// filter by region if supplied
			!strings.HasPrefix(scan.Region, filterRegion) {
			continue
		}
		fmt.Fprintf(w, gameStart, scan.Name, scan.Name)
		fmt.Fprintf(w, gameEntry,
			scan.Region,
			scan.BoxState, scan.BoxState,
			scan.MediaState, scan.MediaState)
	}
	fmt.Fprintf(w, endTable)
}
}
