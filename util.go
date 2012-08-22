// 22 august 2012
package main

import (
	// getWikiAPIData
	"fmt"
	"io/ioutil"
	"net/http"
)

// TODO needs a better name
func getWikiAPIData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading from %s: %v", url, err)
	}
	return b, nil
}
