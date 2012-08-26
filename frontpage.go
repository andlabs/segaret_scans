// 24 august 2012
package main

import (
	"fmt"
	"net/http"
	"strings"
)

var frontpage_top = `<html>
<head>
	<title>Sega Retro Scan Information</title>
</head>
<body>
	<h1>Sega Retro Scan Information</h1>
	<p>Welcome to the scan information page. Please enter the console to look at in the URL, or click on one of the following links to go to that console's page.</p>
	<p>Once on a console's page, you can filter by region by adding <tt>?region=(two-letter region code)</tt> to the end of the URL. For instance, to show only American games, add <tt>?region=US</tt>.</p>
	<table>
		<tr>
			<th>Console</th>
			<th>(statistics go here)</th>
		</tr>
`

var frontpage_console = `
		<tr>
			<td><a href="http://andlabs.sonicretro.org/scans/%s">%s</a></td>
		</tr>
`

var frontpage_bottom = `
	</table>
`

var omitConsoles = map[string]bool{
	// genres, not consoles
	"Action":						true,
	"Adventure":					true,

	// qualifiers, not consoles
	"Unlicensed":					true,
	"Unreleased":					true,

	// these are download only and thus won't have scans OR are services and thus have dupes
	"Android":					true,
	"Game Toshokan":				true,
	"IOS":						true,
	"Java":						true,
	"Meganet":					true,
	"PlayStation Network":			true,
	"Sega Channel":				true,
	"Steam":						true,
	"Tectoy Mega Net":				true,
	"Virtual Console":				true,
	"WiiWare":					true,
	"Xbox Live Arcade":				true,

	// variants on arcade boards
	"Model 2A CRX":				true,
	"Model 2B CRX":				true,
	"Model 2C CRX":				true,
	"Model 3 Step 2.1":				true,
	"NAOMI 2 Satellite Terminal":		true,
	"NAOMI GD-ROM":				true,
}

func generateFrontPage(w http.ResponseWriter) {
	fmt.Fprintf(w, frontpage_top)
	consoles, nGames, err := sql_getconsoles()
	if err != nil {
		fmt.Fprintf(w, frontpage_bottom + "\n<p><b>Error: %s</p>\n", err)
		return
	}
	for i, s := range consoles {
		if nGames[i] != 0 &&						// omit empty categories
			!strings.HasPrefix(s, "19") &&			// omit years
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
			!omitConsoles[s] {					// explicitly omitted
				fmt.Fprintf(w, frontpage_console, s, s)
		}
	}
	fmt.Fprintf(w, frontpage_bottom)
}
