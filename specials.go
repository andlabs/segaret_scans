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

func showAllInvalid(w http.ResponseWriter, r *http.Request) error {
	consoles, err := sql_getconsoles(filterConsole)
	if err != nil {
		return fmt.Errorf("Error getting list of consoles: %v", err)
	}
	fmt.Fprintln(w, "<html><head><title>[invalid scanboxes]</title><body>")
	for _, s := range consoles {
		games, err := GetGameList(s)
		if err != nil {
			fmt.Fprintf(w,  "<p>Error getting game list: %v</p>\n", err)
			continue
		}
		for _, g := range games {
			scans, err := GetScans(g, s)
			if err == ErrGameNoScans {		// omit games for this console that will not have scans
				continue
			}
			if err != nil {
				fmt.Fprintf(w,  "<p>Error getting scans for %s: %v</p>\n", g, err)
				continue
			}
			fmt.Fprintf(w, "<ul>\n")
			for _, v := range scans {
				if v.Console == "" || v.Region == "" {
					fmt.Fprintf(w, `<li><a href="http://segaretro.org/%s">%s</a></li>`, g, g)
					break
				}
			}
			fmt.Fprintf(w, "</ul>\n")
		}
	}
	return nil
}
