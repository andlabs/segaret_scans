// 22-24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"html/template"
)

var report_text = `<html>
<head>
	<title>Sega Retro Scan Information: {{.Console}}</title>
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
	<h1>Sega Retro Scan Information: {{.Console}}</h1>
	{{.Stats.HTML}}
	<br>
	<table>
		<tr>
			<th><a href="{{.URL}}">Game</a></th>
			<th><a href="{{.URL_SortRegion}}">Region</a></th>
			<th><a href="{{.URL_SortBox}}">Box</a></th>
			<th><a href="{{.URL_SortMedia}}">Media</a></th>
		</tr>
{{$filter := .FilterRegion}}{{range .Scans}}
{{if .Error}}
		<tr>
			<td><a href="http://segaretro.org/{{.Name}}">{{.Name}}</a></td>
			<td colspan=3 class=Error>Error: {{.Error.String}}</td>
		</tr>
{{else}}{{if .HasNoScans}}
		<tr>
			<td><a href="http://segaretro.org/{{.Name}}">{{.Name}}</a></td>
			<td colspan=3 class=Missing>No scans</td>
		</tr>
{{else}}{{if filterRegion .Region $filter}}
		<tr>
			<td><a href="http://segaretro.org/{{.Name}}">{{.Name}}</a></td>
			<td>{{.Region}}</td>
			<td class={{.BoxState}}>{{.BoxState}}</td>
			<td class={{.MediaState}}>{{.MediaState}}</td>
		</tr>
{{end}}{{end}}{{end}}
{{end}}
	</table>
`

type ReportPageContents struct {
	Console			string
	Stats				Stats
	FilterRegion		string
	URL				string
	URL_SortRegion	string
	URL_SortBox		string
	URL_SortMedia		string
	Scans			ScanSet
}

var report_template *template.Template

func init() {
	var err error

	report_template = template.New("report")
	report_template = report_template.Funcs(template.FuncMap{
		"filterRegion":	func(r, what string) bool {
			if what == "" {		// no filter
				return true
			}
			return strings.HasPrefix(r, what)
		},
	})
	report_template, err = report_template.Parse(report_text)
	if err != nil {
		panic(err)
	}
}

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

//	fmt.Fprintf(w, top, console, console)
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
	report_template.Execute(w, ReportPageContents{
		Console:			console,
		Stats:			stats,
		FilterRegion:		filterRegion,
		URL:				urlNoSort(url),
		URL_SortRegion:	urlSort(url, "region"),
		URL_SortBox:		urlSort(url, "box"),
		URL_SortMedia:	urlSort(url, "media"),
		Scans:			scans,
	})
}
