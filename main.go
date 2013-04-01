// 22 august 2012
package main

import (
	"fmt"
	"net/http"
	"time"
	"runtime/debug"
	"flag"			// command line parsing
	"os"
)

var bottom = `
	<p>Page generated in %v.</p>
</body>
</html>
`

var specials = map[string]func(sql *SQL, w http.ResponseWriter, r *http.Request) error{
	"filter":		applyFilter,
//	"missing":		showAllMissing,
//	"invalid":		showAllInvalid,
	"listcompare":	listcompare,
	"listotherconsoles":	listotherconsoles,
}

func do(w http.ResponseWriter, r *http.Request) {
	var err error
	var console string

	defer func() {
		err := recover()
		if err != nil {
			http.Error(w,
				fmt.Sprintf("runtime panic: %v\nstack trace:\n%s\n",
					err, debug.Stack()),
				http.StatusInternalServerError)
		}
	}()

	startTime := time.Now()
	if len(r.URL.Path) < 7 {	// no trailing /
		http.Redirect(w, r, "/scans/", http.StatusFound)
		return
	}
	sql, err := NewSQL()
	if err != nil {
		http.Error(w,
			fmt.Sprintf("error creating new SQL connection: %v", err),
			http.StatusInternalServerError)
		return
	}
	defer sql.Close()
	console = r.URL.Path[7:]
	if console == "" {
//		fmt.Fprintln(w, "Server up. Specify the console in the URL.")
		special := r.URL.Query().Get("special")
		if f, ok := specials[special]; ok && f != nil {
			err = f(sql, w, r)
		} else {
			err = generateFrontPage(sql, w, *r.URL)
		}
	} else {
		err = generateConsoleReport(console, sql, w, *r.URL)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, bottom, time.Now().Sub(startTime))
}

func main() {
	flag.Usage = func() {
		fmt.Printf("usage: %s [options] config-file\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	configFile := flag.Arg(0)
	if *configFlag {
		makeConfig(configFile)
		return
	}

	// otherwise, run the server
	loadConfig(configFile)
	http.HandleFunc("/", do)
	http.ListenAndServe("127.0.0.1:6060", nil)
}
