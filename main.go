// 22 august 2012
package main

import (
	"fmt"
	"net/http"
	"time"
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
	<table>
		<tr>
			<th colspan=2>Game</th>
			<th>Box</th>
			<th>Cart/Disc</th>
		</tr>
`

var gameEntry = `
		<tr>
			<td>%s</td>
			<td>%s</td>
			<td class=%v>%v</td>
			<td class=%v>%v</td>
		</tr>
`

var gameError = `
		<tr>
			<td>%s</td>
			<td colspan=3 class=Error>Error: %v</td>
		</tr>
`

var gameNoScans = `
		<tr>
			<td>%s</td>
			<td colspan=3 class=Missing>No scans</td>
		</tr>
`

var bottom = `
	</table>
	<p>Page generated in %v.</p>
</body>
</html>
`

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

func getConsoleInfo(w http.ResponseWriter, r *http.Request) {
	console := r.URL.Path[1:]
	if console == "" {
		fmt.Fprintln(w, "Server up. Specify the console in the URL.")
		return
	}
	startTime := time.Now()
	games, err := GetGameList(console)
	if err != nil {
		fmt.Fprintf(w, "Error getting %s game list: %v\n", console, err)
		return
	}
	fmt.Fprintf(w, top, console, console)
	for _, game := range games {
fmt.Println(game)
		scans, err := GetScans(game)
		if err != nil {
			fmt.Fprintf(w, gameError, game, err)
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
			fmt.Fprintf(w, gameEntry,
				game,
				scan.Region,
				boxState, boxState,
				mediaState, mediaState)
		}
		if nScans == 0 {
			fmt.Fprintf(w, gameNoScans, game)
		}
	}
	fmt.Fprintf(w, bottom, time.Now().Sub(startTime))
}

func main() {
	http.HandleFunc("/", getConsoleInfo)
	http.ListenAndServe(":6060", nil)
}
