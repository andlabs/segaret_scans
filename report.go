// 22-24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GameScan struct {
	Name		string
	HasNoScans	bool			// for whole games
	Region		string
	BoxState		ScanState
	MediaState	ScanState
	Error			error
}

type ScanSet []*GameScan

func getMediaState(scan Scan) ScanState {
	if scan.Cart == "" && scan.Disc == "" {
		return Missing
	}
	if scan.Cart != "" && scan.Disc == "" {
		return scan.CartScanState()
	}
	if scan.Cart == "" && scan.Disc != "" {
		return scan.DiscScanState()
	}
	return scan.CartScanState().Join(scan.DiscScanState())	// else
}


func GetConsoleScans(console string) (ScanSet, error) {
	var gameScans ScanSet

	games, err := GetGameList(console)
	if err != nil {
		return nil, fmt.Errorf("error getting %s game list: %v", console, err)
	}
	for _, game := range games {
//fmt.Println(game)
		if strings.HasPrefix(game, "List of " + console + " games") {	// omit list from report
			continue
		}
		scans, err := GetScans(game)
		if err != nil {
			gameScans = append(gameScans, &GameScan{
				Name:	game,
				Error:	err,
			})
			continue
		}
		if len(scans) == 0 {				// there are no scans at all
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				HasNoScans:	true,
			})
			continue
		}
		nScans := 0
		for _, scan := range scans {
			var mediaState ScanState

			if scan.Console != console {	// omit scans from other consoles
				continue
			}
			nScans++
			boxState := scan.BoxScanState()
			mediaState = getMediaState(scan)
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				Region:		scan.Region,
				BoxState:		boxState,
				MediaState:	mediaState,
			})
		}
		if nScans == 0 {					// there are no scans for the specified console
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				HasNoScans:	true,
			})
			continue
		}
	}
	return gameScans, nil
}

type Stats struct {
	nBoxScans	int
	nBoxHave		int
	nBoxGood	int
	nMediaScans	int
	nMediaHave	int
	nMediaGood	int
}

func (scans ScanSet) GetStats(filterRegion string) (stats Stats) {
	for _, scan := range scans {
		if scan.Error != nil || scan.HasNoScans {		// TODO really skip entries without scans?
			continue
		}
		if filterRegion != "" &&
			!strings.HasPrefix(scan.Region, filterRegion) {
			continue
		}
		stats.nBoxScans++
		switch scan.BoxState {
		case Good:
			stats.nBoxGood++
			fallthrough
		case Bad, Incomplete:
			stats.nBoxHave++
		}
		stats.nMediaScans++
		switch scan.MediaState {
		case Good:
			stats.nMediaGood++
			fallthrough
		case Bad, Incomplete:
			stats.nMediaHave++
		}
	}
	return
}

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
			<th align=right>Box</th>
			<td style="border-left: 1px solid">%d have/%d total (%.2f%%)</td>
			<td style="border-left: 1px solid">%d good/%d total (%.2f%%)</td>
		</tr>
		<tr>
			<th align=right>Media</th>
			<td style="border-left: 1px solid">%d have/%d total (%.2f%%)</td>
			<td style="border-left: 1px solid">%d good/%d total (%.2f%%)</td>
		</tr>
	</table>
	<br>
`

var beginTable = `
	<table>
		<tr>
			<th colspan=2>Game</th>
			<th>Box</th>
			<th>Media</th>
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

func pcnt(_a, _b int) float64 {
	a, b := float64(_a), float64(_b)
	return (a / b) * 100.0
}

func generateConsoleReport(console string, w http.ResponseWriter, query url.Values) {
	var filterRegion string

	fmt.Fprintf(w, top, console, console)
	scans, err := GetConsoleScans(console)
	if err != nil {
		fmt.Fprintf(w, "<p>Error getting %s scan info: %v</p>\n", console, err)
		return
	}
	if x, ok := query[filterRegionName]; ok && len(x) > 0 {	// filter by region if supplied
		filterRegion = x[0]
	}
	stats := scans.GetStats(filterRegion)
	fmt.Fprintf(w, gameStats,
		stats.nBoxHave, stats.nBoxScans, pcnt(stats.nBoxHave, stats.nBoxScans),
		stats.nBoxGood, stats.nBoxScans, pcnt(stats.nBoxGood, stats.nBoxScans),
		stats.nMediaHave, stats.nMediaScans, pcnt(stats.nMediaHave, stats.nMediaScans),
		stats.nMediaGood, stats.nMediaScans, pcnt(stats.nMediaGood, stats.nMediaScans))
	fmt.Fprintf(w, beginTable)
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
