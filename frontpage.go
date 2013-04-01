// 24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"html/template"
	"sort"
)

var frontpage_text = `{{define "pageTitle"}}{{siteName}}{{end}}
{{define "pageContent"}}
	<p>Welcome to the scan information page. Please enter the console to look at in the URL, or click on one of the following links to go to that console's page. On a console page, you can filter results by region and sort the results.</p>

	<table style="border: none" cellspacing=0><tr><td style="border: none">
	<b>Overall Status:</b><br>
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
			<td><a href="{{reportpage .Console}}">{{.Console}}</a></td>
{{if .Error}}
			<td colspan=2 class="Error">Error grabbing progress: {{.Error}}</td>
{{else}}
			<td>{{.BoxBar}}</td>
			<td>{{.MediaBar}}</td>
{{end}}
		</tr>
{{end}}
	</table>
{{end}}`

var frontpage_template *template.Template

type FrontPageContents struct {
	Stats			Stats
	Entries		[]ConsoleTableEntry
}

// TODO change Console to Category everywhere?

type ConsoleTableEntry struct {
	Console		string
	Error			error
	BoxBar		template.HTML
	MediaBar		template.HTML
}

// for sorting
type ConsoleTableEntries []ConsoleTableEntry

func frontpage_init() {
	frontpage_template = NewTemplate(frontpage_text, "front page")
}

func init() {
	addInit(frontpage_init)
}

func generateFrontPage(sql *SQL, w http.ResponseWriter, url url.URL) error {
	overallStats := Stats{}
	consoleEntries := ConsoleTableEntries{}

	sets, err := Run(sql, config.Consoles)
	if err != nil {
		return fmt.Errorf("error getting scan information: %v", err)
	}
	for category, ss := range sets {
		stats := ss.GetStats("")
		boxes := stats.BoxProgressBar()
		media := stats.MediaProgressBar()
		consoleEntries = append(consoleEntries, ConsoleTableEntry{
			Console:		category,
			BoxBar:		boxes,
			MediaBar:		media,
		})
		overallStats.Add(stats)
	}
	sort.Sort(consoleEntries)
	frontpage_template.Execute(w, FrontPageContents{
		Stats:	overallStats,
		Entries:	consoleEntries,
	})
	return nil
}

func (c ConsoleTableEntries) Len() int {
	return len(c)
}

func (c ConsoleTableEntries) Less(i, j int) bool {
	return c[i].Console < c[j].Console
}

func (c ConsoleTableEntries) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
