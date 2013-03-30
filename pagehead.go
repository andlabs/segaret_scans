// 8 september 2012
package main

import (
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

var pageTop = `<html>
<head>
	<title>{{template "pageTitle" .}}</title>
	<style type="text/css">
		.Bad {
			background-color: {{badcolor}};
		}
		.Missing {
			background-color: {{missingcolor}};
		}
		.Incomplete {
			background-color: {{incompletecolor}};
		}
		.Good {
			background-color: {{goodcolor}};
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

` + pbarCSS + `
	</style>
</head>
<body>
	<h1>{{template "pageTitle" .}}</h1>
	{{template "pageContent" .}}`

// templates share functions
var tFunctions = template.FuncMap{
	"badcolor":		func() string { return color_bad },
	"missingcolor":		func() string { return color_missing },
	"incompletecolor":	func() string { return color_incomplete },
	"goodcolor":		func() string { return color_good },

	"siteName":		func() string { return config.SiteName },

	"filterRegion":		filterRegion,
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
	t, err := template.New(forWhat).Funcs(tFunctions).Parse(pageTop + text)
	if err != nil {
		log.Fatalf("could not prepare %s template: %v", forWhat, err)
	}
	return t
}
