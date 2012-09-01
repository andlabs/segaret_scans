// 22 august 2012
package main

import (
	"fmt"
	"net/http"
	"time"
	"runtime/debug"
)

var bottom = `
	<p>Page generated in %v.</p>
</body>
</html>
`

var specials = map[string]func(w http.ResponseWriter, r *http.Request) error{
	"missing":		showAllMissing,
	"filter":		applyFilter,
}

func do(w http.ResponseWriter, r *http.Request) {
	var err error

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
	console := r.URL.Path[7:]
	if console == "" {
//		fmt.Fprintln(w, "Server up. Specify the console in the URL.")
		special := r.URL.Query().Get("special")
		if f, ok := specials[special]; ok && f != nil {
			err = f(w, r)
		} else {
			err = generateFrontPage(w, *r.URL)
		}
	} else {
		err = generateConsoleReport(console, w, *r.URL)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, bottom, time.Now().Sub(startTime))
}

func main() {
	http.HandleFunc("/", do)
	http.ListenAndServe("127.0.0.1:6060", nil)
}
