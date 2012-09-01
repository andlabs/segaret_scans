// 1 september 2012
package main

import (
	"fmt"
	"net/http"
)

func showAllMissing(w http.ResponseWriter, r *http.Request) error {
	consoles, err := sql_getconsoles(filterConsole)
	if err != nil {
		return fmt.Errorf("Error getting list of consoles: %v", err)
	}
	fmt.Fprintln(w, "<html><head><title>[missing pages]</title><body>")
	for _, s := range consoles {
		ss, err := GetConsoleScans(s)
		fmt.Fprintf(w, "<h1>%s</h1>", s)
		if err != nil {
			fmt.Fprintf(w,  "<p>Error: %v</p>\n", err)
			continue
		}
		fmt.Fprintf(w,  "<ul>\n")
		for _, g := range ss {
			if g.HasNoScans {
				fmt.Fprintf(w, `<li><a href="http://segaretro.org/%s">%s</a>`, g.Name, g.Name)
			}
		}
		fmt.Fprintf(w, "</ul>\n")
	}
	return nil
}
