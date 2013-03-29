// 8 september 2012
package main

import (
	"fmt"
	"html/template"
	"log"
	"strings"			// filterRegion
)

const (
	// progress bar
//	color_red = color.RGBA{255, 0, 0, 255}
//	color_green = color.RGBA{0, 255, 0, 255}
//	color_yellow = color.RGBA{255, 255, 0, 255}

	// CSS from Scarred Sun
	color_bad = "#C00"
	color_good = "#0C0"
//	color_incomplete = "#888800"

	color_incomplete = "#CCCC00"
	color_missing = "#CCCCCC"
)

const pageTop_form = `<html>
<head>
	<title>%%s</title>
	<style type="text/css">
		.Bad {
			background-color: %s;
		}
		.Missing {
			background-color: %s;
		}
		.Incomplete {
			background-color: %s;
		}
		.Good {
			background-color: %s;
		}
		.Error {
			background-color: #000000;
			color: #FFFFFF;
		}
		table {
			border-collapse:collapse;
		}
		td {
			border:1px solid #000;
			padding:2px;
			font-size:0.9em;
		}
		body {
			font-family:Verdana,Helvetica,DejaVu Sans,sans-serif;
			font-size:0.8em;
		}
		th {
			background-color: #999;
			border:1px solid #000;
		}
		th a {
			text-decoration: none !important;
			color:#006;
		}

		%s
	</style>
</head>`

var pageTop_actual string // prepared

func pagehead_init() {
	pageTop_actual = fmt.Sprintf(pageTop_form,
		color_bad,
		color_missing,
		color_incomplete,
		color_good,
		pbarCSS)
}

func init() {
	addInit(pagehead_init)
}

// templates share functions
var tFunctions = template.FuncMap{
	"filterRegion":		filterRegion,
	"pageTop":		pageTop,
	"makeTitle":		makeTitle,
	"siteBaseURL":		siteBaseURL,
	"wikiBaseURL":		wikiBaseURL,
	"toURL":			toURL,
}

// template function {{filterRegion given_region region_to_filter}}
func filterRegion(r, what string) bool {
	if what == "" {		// no filter
		return true
	}
	return strings.Contains(strings.ToLower(r), strings.ToLower(what))
}

// template function {{pageTop page_title}}
func pageTop(title string) template.HTML {
	return template.HTML(fmt.Sprintf(pageTop_actual, title))
}

// template function {{makeTitle title}}
func makeTitle(title string) string {
	if title == "" {
		return config.SiteName
	}
	return config.SiteName + ": " + title
}

func siteBaseURL() string {
	return config.SiteBaseURL
}

func wikiBaseURL() string {
	return config.WikiBaseURL
}

func toURL(pageName string) template.URL {
	return template.URL(config.WikiBaseURL + pageName)
}

func NewTemplate(text string, forWhat string) *template.Template {
	t, err := template.New(forWhat).Funcs(tFunctions).Parse(text)
	if err != nil {
		log.Fatalf("could not prepare %s template: %v", forWhat, err)
	}
	return t
}
