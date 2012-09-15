// 24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"html/template"
)

var frontpage_text = `{{pageTop "Sega Retro Scan Information"}}
<body>
	<h1>Sega Retro Scan Information</h1>
	<p>Welcome to the scan information page. Please enter the console to look at in the URL, or click on one of the following links to go to that console's page. On a console page, you can filter results by region and sort the results.</p>

	<table style="border: none" cellspacing=0><tr><td style="border: none">
	<b>Overall Status:</h3></b><br>
	{{.Stats.HTML}}
	</td><td style="border: none" valign=top>
	<b>Legend:</b><br>
	<span class=Missing>Missing</span>: we have nothing for this particular scan<br>
	<span class=Bad>Bad</span>: at least one part of the scan is bad<br>
	<span class=Incomplete>Incomplete</span>: we are missing some parts but the parts we do have are good<br>
	<span class=Good>Good</span>: we have everything and everything is up to standard</td></tr></table><br>

	<table>
		<tr>
			<th>Console</th>
			<th>Box Scan Progress</th>
			<th>Media Scan Progress</th>
		</tr>
{{range .Entries}}
		<tr>
			<td><a href="http://andlabs.sonicretro.org/scans/{{.Console}}">{{.Console}}</a></td>
{{if .Error}}
			<td colspan=2 class="Error">Error grabbing progress: {{.Error}}</td>
{{else}}
			<td><img src="data:image/png;base64,{{.BoxBar}}"></td>
			<td><img src="data:image/png;base64,{{.MediaBar}}"></td>
{{end}}
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
	Error			error
	BoxBar		string
	MediaBar		string
}

func init() {
	frontpage_template = NewTemplate(frontpage_text, "front page")
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
		ss, err := GetConsoleScans(s)
		if err == nil {
			stats := ss.GetStats("")
			boxes := stats.BoxProgressBar()
			media  := stats.MediaProgressBar()
			consoleEntries = append(consoleEntries, ConsoleTableEntry{
				Console:		s,
				BoxBar:		boxes,
				MediaBar:		media,
			})
			overallStats.Add(stats)
		} else {
			consoleEntries = append(consoleEntries, ConsoleTableEntry{
				Console:		s,
				Error:		err,
			})
		}
	}
	frontpage_template.Execute(w, FrontPageContents{
		Stats:	overallStats,
		Entries:	consoleEntries,
	})
	return nil
}
