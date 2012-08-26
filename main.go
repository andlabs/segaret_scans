// 22 august 2012
package main

import (
	"fmt"
	"net/http"
	"time"
)

var bottom = `
	<p>Page generated in %v.</p>
</body>
</html>
`

func getConsoleInfo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	console := r.URL.Path[7:]
	if console == "" {
//		fmt.Fprintln(w, "Server up. Specify the console in the URL.")
		generateFrontPage(w)
	} else {
		generateConsoleReport(console, w, r.URL.Query())
	}
	fmt.Fprintf(w, bottom, time.Now().Sub(startTime))
}

func main() {
	http.HandleFunc("/", getConsoleInfo)
	http.ListenAndServe("127.0.0.1:6060", nil)
}
