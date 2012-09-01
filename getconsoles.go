// 1 september 2011
package main

import (
	"strings"
)

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

	// arcade systems that don't use removable media
	"AS-1":						true,
	"Model 2":					true,
	"Model 3":					true,
	"System 1":					true,
	"System 2":					true,
}

func filterConsole(s string) bool {
	return !strings.HasPrefix(s, "19") &&			// omit years
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
		!omitConsoles[s]
}

func GetConsoleList() ([]string, error) {
	return sql_getconsoles(filterConsole)
}
