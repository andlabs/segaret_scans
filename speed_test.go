// 27 august 2012
package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func BenchmarkSpeed(b *testing.B) {
	b.StopTimer()
	w := httptest.NewRecorder()
	r := &http.Request{
		URL:		&url.URL{
			Path:		"/scans/",
		},
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getConsoleInfo(w, r)
	}
}
