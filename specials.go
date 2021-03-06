// 1 september 2012
package main

// TODO all this needs to be rewritten now that the semantics of grabbing scans has changed

import (
	"fmt"
	"net/http"
	"strings"
)

// TODO NoScans
func listotherconsoles(sql *SQL, w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintln(w, "<html><head><title>[extra consoles]</title><body>")

	var consoles = map[string]bool{}
	for _, c := range config.Consoles {
		consoles[c] = true
	}

	sbl, err := sql.db_scanbox.Query(
		`SELECT _page, console
			FROM Scanbox;`)
	if err != nil { panic(err) }
	defer sbl.Close()

	var n = map[string]int{}
	var pg = map[string][]string{}

	for sbl.Next() {
		var page string
		var console string

		err = sbl.Scan(&page, &console)
		if err != nil { panic(err) }
		if consoles[console] {		// skip consoles we have
			continue
		}
		what := `<a href="http://segaretro.org/` + page + `">` + page + `</a>`
		in := false
		for i := 0; i < len(pg[console]); i++ {
			if pg[console][i] == what {
				in = true
			}
		}
		if !in {
			n[console]++
			if len(pg[console]) < 5 {
				pg[console] = append(pg[console], what)
			}
		}
	}

	fmt.Fprintln(w, "<pre>")
	for console := range pg {
		fmt.Fprintf(w, "%20s %s", console, strings.Join(pg[console], ", "))
		if n[console] > 5 {
			fmt.Fprintf(w, ", %d more", n[console] - 5)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "</pre>")
	return nil
}

func listcompare(sql *SQL, w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintln(w, "<html><head><title>[missing pages]</title><body>")

	p := func(f string, a ...interface{}){panic(fmt.Sprintf(f,a...))}

	type S struct{}
	var s = S(struct{}{})

	categorylist := map[string]S{}
	clscan := map[string]S{}
	for category := range config.Consoles {
		games, err := sql.GetGameList(category)
		if err != nil {
			p("error getting %s list: %v", category, err)
		}
		for _, g := range games {
			categorylist[g] = s
			clscan[g] = s
		}
	}

	scanboxlist := map[string]S{}
	sbl, err := sql.db_scanbox.Query(
		`SELECT _page
			FROM Scanbox
		UNION SELECT _page
			FROM NoScans;`)
	if err != nil {
		p("could not run scanbox list query (for scan list): %v", err)
	}
	defer sbl.Close()

	for sbl.Next() {
		var d string

		err := sbl.Scan(&d)
		if err != nil {
			p("error reading entry in scanbox list query (for scan list): %v", err)
		}
		scanboxlist[d] = s
	}

	for g := range clscan {
		if _, ok := scanboxlist[g]; ok {
			delete(scanboxlist, g)
			delete(categorylist, g)
		}
	}

	fmt.Fprintln(w, `<pre>Only in category list:`)
	for g := range categorylist {
		fmt.Fprintf(w, "<a href=\"http://segaretro.org/%s\">%s</a>\n", g, g)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, `Only in scanbox db:`)
	for g := range scanboxlist {
		fmt.Fprintln(w, g)
	}
	fmt.Fprintln(w, "</pre>")

	return nil
}

/*
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

*/
