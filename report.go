// 22-24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

var gameStats = `
	<table>
		<tr>
			<th rowspan=3 valign=top align=right>Box</th>
			<td>We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td>%d (%.2f%%) of them are good (%.2f%% overall)</td></tr>
		<tr><td>%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
		<tr>
			<th rowspan=3 valign=top align=right>Media</th>
			<td>We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td>%d (%.2f%%) of them are good (%.2f%% overall)</td>
		<tr><td>%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
	</table>
	<br>
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

func urlSort(url url.URL, order string) string {
	q := url.Query()
	q.Del("sort")
	q.Add("sort", order)
	url.RawQuery = q.Encode()
	return url.String()
}

func urlNoSort(url url.URL) string {
	q := url.Query()
	q.Del("sort")
	url.RawQuery = q.Encode()
	return url.String()
}

func generateConsoleReport(console string, w http.ResponseWriter, url url.URL) {
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
	fmt.Fprintf(w, gameStats,
		stats.nBoxHave, stats.nBoxScans, stats.pBoxHave,
		stats.nBoxGood, stats.pBoxGood, stats.pBoxGoodAll,
		stats.nBoxBad, stats.pBoxBad, stats.pBoxBadAll,
		stats.nMediaHave, stats.nMediaScans, stats.pMediaHave,
		stats.nMediaGood, stats.pMediaGood, stats.pMediaGoodAll,
		stats.nMediaBad, stats.pMediaBad, stats.pMediaBadAll)
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
