// 22-24 august 2012
package main

import (
	"fmt"
	"net/http"
	"strings"
)

type GameRegionScan struct {
	Region		string
	BoxState		ScanState
	MediaState	ScanState
}

type GameScanSet struct {
	Name	string
	Scans	[]GameRegionScan
	Error		error
}

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


func GetConsoleInfo(console string) ([]GameScanSet, error) {
	var gameScans []GameScanSet

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
			gameScans = append(gameScans, GameScanSet{
				Name:	game,
				Error:	err,
			})
			continue
		}
		gameEntry := GameScanSet{
			Name:	game,
		}
		for _, scan := range scans {
			var mediaState ScanState

			if scan.Console != console {	// omit scans from other consoles
				continue
			}
			boxState := scan.BoxScanState()
			mediaState = getMediaState(scan)
			gameEntry.Scans = append(gameEntry.Scans, GameRegionScan{
				Region:		scan.Region,
				BoxState:		boxState,
				MediaState:	mediaState,
			})
		}
		gameScans = append(gameScans, gameEntry)
	}
	return gameScans, nil
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
	<table>
		<tr>
			<th colspan=2>Game</th>
			<th>Box</th>
			<th>Cart/Disc</th>
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

var report_bottom = `
	</table>
`

func generateConsoleInfo(console string, w http.ResponseWriter) {
	fmt.Fprintf(w, top, console, console)
	games, err := GetConsoleInfo(console)
	if err != nil {
		fmt.Fprintf(w, report_bottom + "\n<p>Error getting %s game list: %v</p>\n", console, err)
		return
	}
	for _, game := range games {
		if game.Error != nil {
			fmt.Fprintf(w, gameStart, game.Name, game.Name)
			fmt.Fprintf(w, gameError, game.Error)
			continue
		}
		for _, scan := range game.Scans {
			fmt.Fprintf(w, gameStart, game.Name, game.Name)
			fmt.Fprintf(w, gameEntry,
				scan.Region,
				scan.BoxState, scan.BoxState,
				scan.MediaState, scan.MediaState)
		}
		if len(game.Scans) == 0 {
			fmt.Fprintf(w, gameStart, game.Name, game.Name)
			fmt.Fprintf(w, gameNoScans)
		}
	}
	fmt.Fprintf(w, report_bottom)
}
