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
	for _, v := range config.ConsolePrefixesToOmit {
		if strings.HasPrefix(s, v) {
			return false
		}
	}
	for _, v := range config.ConsoleSuffixesToOmit {
		if strings.HasSuffix(s, v) {
			return false
		}
	}
	return !omitConsoles[s]
}

func GetConsoleList() ([]string, error) {
	return sql_getconsoles(filterConsole)
}
