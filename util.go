// 22 august 2012
package main

import (
	// getWikiAPIData
	"fmt"
	"io/ioutil"
	"net/http"
)

const ServerIP = "208.94.244.139"

// TODO needs a better name
func getWikiAPIData(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", "http://" + ServerIP + url, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request for http://segaretro.org%s: %v", url, err)
	}
	req.Host = "segaretro.org"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to http://segaretro.org%s: %v", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading from http://segaretro.org%s: %v", url, err)
	}
	return b, nil
}
