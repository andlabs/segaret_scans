// 22-24 august 2012
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"html/template"
)

var report_text = `{{define "pageTitle"}}{{siteName}}: {{.Console}}{{end}}
{{define "pageContent"}}
	<table>
		<tr><td>{{.Stats.HTML}}</td>
		<td valign=top><table>
			<tr><td><b>Filter by region:</b> Enter a region string (such as the two-letter code or a collection name) to filter by. Leave the field blank to remove the filter. The statistics to the left will change to reflect the filter.</td></tr>
			<tr><td><form action="/scans/?special=filter" method=POST>
				<input type=text name=region>
				<input type=submit value=Apply>
			</form></td></tr>
		</table></td></tr>
	</table>
	<br>
	<table>
		<tr>
			<th><a href="{{.URL_NoSort}}">Game</a></th>
			<th><a href="{{.URL_SortRegion}}">Region</a></th>
			<th><a href="{{.URL_SortBox}}">Box</a></th>
			<th><a href="{{.URL_SortMedia}}">Media</a></th>
		</tr>
{{$filter := .FilterRegion}}{{range .Scans}}
{{if .Error}}
		<tr>
			<td><a href="{{wikipage .Name}}">{{.Name}}</a></td>
			<td colspan=3 class=Error>Error: {{.Error}}</td>
		</tr>
{{else}}{{if .HasNoScans}}
		<tr>
			<td><a href="{{wikipage .Name}}">{{.Name}}</a></td>
			<td colspan=3 class=Missing>No scans</td>
		</tr>
{{else}}{{if filterRegion .Region $filter}}
		<tr>
			<td><a href="{{wikipage .Name}}">{{.Name}}</a></td>
			<td>{{.Region}}</td>
			<td class={{.BoxState.TypeString}}>{{.BoxState}}</td>
			<td class={{.MediaState.TypeString}}>{{.MediaState}}</td>
		</tr>
{{end}}{{end}}{{end}}
{{end}}
	</table>
{{end}}`

type ReportPageContents struct {
	Console			string
	Stats				Stats
	FilterRegion		string
	URL_NoSort		string
	URL_SortRegion	string
	URL_SortBox		string
	URL_SortMedia		string
	Scans			ScanSet
}

var report_template *template.Template

func report_init() {
	report_template = NewTemplate(report_text, "report page")
}

func init() {
	addInit(report_init)
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

func generateConsoleReport(category string, w http.ResponseWriter, url url.URL) error {
	var filterRegion string
	var scans ScanSet
	var err error

	scans, err = RunOne(category)
	if err != nil {
		return fmt.Errorf("Error getting %s scan info: %v", category, err)
	}
	query := url.Query()
	if x, ok := query[filterRegionName]; ok && len(x) > 0 {	// filter by region if supplied
		filterRegion = x[0]
	}
	so := SortByName								// default sort order
	if x, ok := query[sortOrderName]; ok && len(x) > 0 {		// sort differently if asked
		if y, ok := sortOrders[x[0]]; ok {				// but only if we passed a valid sort order
			so = y
		}
	}
	scans.Sort(so)
	stats := scans.GetStats(filterRegion)
	report_template.Execute(w, ReportPageContents{
		Console:			category,
		Stats:			stats,
		FilterRegion:		filterRegion,
		URL_NoSort:		urlNoSort(url),
		URL_SortRegion:	urlSort(url, "region"),
		URL_SortBox:		urlSort(url, "box"),
		URL_SortMedia:	urlSort(url, "media"),
		Scans:			scans,
	})
	return nil
}

func applyFilter(w http.ResponseWriter, r *http.Request) error {
	newURL, err := url.Parse(r.Referer())
	if err != nil {
		return err
	}
	filterRegion := r.FormValue("region")
	query := newURL.Query()
	query.Del("region")
	if filterRegion != "" {		// want a filter
		query.Add("region", filterRegion)
	}
	newURL.RawQuery = query.Encode()
	http.Redirect(w, r, newURL.String(), http.StatusFound)
	return nil
}
