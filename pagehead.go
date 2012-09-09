// 8 september 2012
package main

import (
	"fmt"
	"image/color"
	"html/template"
	"log"
	"strings"			// filterRegion
)

var (
	// progress bar
//	color_red = color.RGBA{255, 0, 0, 255}
//	color_green = color.RGBA{0, 255, 0, 255}
//	color_yellow = color.RGBA{255, 255, 0, 255}

	// CSS from Scarred Sun
	color_bad = color.RGBA{0xCC, 0x00, 0x00, 255} // #C00
	color_good = color.RGBA{0x00, 0xCC, 0x00, 255} // #0C0
//	color_incomplete = color.RGBA{0x88, 0x88, 0x00, 255} // #888800

	color_incomplete = color.RGBA{0xCC, 0xCC, 0x00, 255} // #CCCC00
	color_missing = color.RGBA{0xCC, 0xCC, 0xCC, 255}
)

func toCSSColor(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X",
		c.R, c.G, c.B)
}

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
	</style>
</head>`

var pageTop_actual string // prepared

func init() {
	pageTop_actual = fmt.Sprintf(pageTop_form,
		toCSSColor(color_bad),
		toCSSColor(color_missing),
		toCSSColor(color_incomplete),
		toCSSColor(color_good))
}

// templates share functions
var tFunctions = template.FuncMap{
	"filterRegion":	filterRegion,
	"pageTop":	pageTop,
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

func NewTemplate(text string, forWhat string) *template.Template {
	t, err := template.New(forWhat).Funcs(tFunctions).Parse(text)
	if err != nil {
		log.Fatalf("could not prepare %s template: %v", forWhat, err)
	}
	return t
}
