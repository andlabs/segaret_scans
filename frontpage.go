// 24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"html/template"
	"log"
)

var frontpage_text = `<html>
<head>
	<title>Sega Retro Scan Information</title>
</head>
<body>
	<h1>Sega Retro Scan Information</h1>
	<p>Welcome to the scan information page. Please enter the console to look at in the URL, or click on one of the following links to go to that console's page.</p>
	<p>Once on a console's page, you can filter by region by adding <tt>?region=(two-letter region code)</tt> to the end of the URL. For instance, to show only American games, add <tt>?region=US</tt>. You can also provide an optional sort order, by region, box quality, or media quality, with <tt>?sort=(region|box|media)</tt>.</p>

	<p><b>Overall Status:</h3></b><br>
	{{.Stats.HTML}}
	</p>

	<table>
		<tr>
			<th>Console</th>
			<th>Box Scan Progress</th>
			<th>Media Scan Progress</th>
		</tr>
{{range .Entries}}
		<tr>
			<td><a href="http://andlabs.sonicretro.org/scans/{{.Console}}">{{.Console}}</a></td>
			<td><img src="data:image/png;base64,{{.BoxBar}}"></td>
			<td><img src="data:image/png;base64,{{.MediaBar}}"></td>
<td>{{.Gentime}}</td>
		</tr>
{{end}}
	</table>
`

var frontpage_template *template.Template

type FrontPageContents struct {
	Stats			Stats
	Entries		[]ConsoleTableEntry
}

type ConsoleTableEntry struct {
	Console		string
	BoxBar		string
	MediaBar		string
	Gentime		string
}

func init() {
	var err error

	frontpage_template, err = template.New("frontpage").Parse(frontpage_text)
	if err != nil {
		log.Fatalf("could not prepare front page template: %v", err)
	}
}

func generateFrontPage(w http.ResponseWriter, url url.URL) error {
	overallStats := Stats{}
	consoleEntries := []ConsoleTableEntry{}

//	fmt.Fprintf(w, frontpage_top)
	consoles, err := GetConsoleList()
	if err != nil {
		return fmt.Errorf("Error getting list of consoles: %v", err)
	}
	for _, s := range consoles {
		start := time.Now()
		ss, err := GetConsoleScans(s)
		gentime := time.Now().Sub(start).String()
		if err != nil {
			panic(err)			// TODO
		}
		stats := ss.GetStats("")
		boxes := stats.BoxProgressBar()
		media  := stats.MediaProgressBar()
		consoleEntries = append(consoleEntries, ConsoleTableEntry{
			Console:		s,
			BoxBar:		boxes,
			MediaBar:		media,
			Gentime:		gentime,
		})
		overallStats.Add(stats)
	}
	frontpage_template.Execute(w, FrontPageContents{
		Stats:	overallStats,
		Entries:	consoleEntries,
	})
	return nil
}
