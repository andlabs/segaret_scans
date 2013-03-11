// 1 september 2011
package main

import (
	"strings"
)

var omitConsoles = map[string]bool{
	// TODO
//	"Aurora":						true,		// not sure what it uses (need to ask Nik)
//	"G80":						true,		// used CPU boards
//	"Europa-R":					true,		// not sure what it uses (too new)
//	"Hikaru":						true,		// not sure what type of ROM board it uses
//	"System E":					true,		// not sure what removable media it used but it definitely used removable media
}

func filterConsole(s string) bool {
	return !strings.HasPrefix(s, "19") &&			// omit years
		!strings.HasPrefix(s, "20") &&
		!strings.HasSuffix(s, " action") &&		// omit genres
		!strings.HasSuffix(s, " adventure") &&
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
		!omitConsoles[s]
}

func GetConsoleList() ([]string, error) {
	return sql_getconsoles(filterConsole)
}
