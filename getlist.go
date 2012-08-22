// 22 august 2012
package main

import (
	"fmt"
	"encoding/json"
	"net/url"		// for urlForConsole()

	// getWikiAPIData
	// "fmt"
	"io/ioutil"
	"net/http"
)

type GameListEntry struct {
//	PageID	string	`json:"pageid"`
	Name	string	`json:"title"`
}

type gameslistcontval struct {
	Cont		string			`json:"cmcontinue"`
}

type gamelistcont struct {
	Cont		gameslistcontval	`json:"categorymembers"`
}

type gamelistquery struct {
	Games	[]GameListEntry	`json:"categorymembers"`
}

type gamelistres struct {
	Games	gamelistquery		`json:"query"`
	Cont		gamelistcont		`json:"query-continue"`
}

func urlForConsole(console string) string {
	return "http://segaretro.org/api.php?format=json&action=query&list=categorymembers&cmtitle=Category:" + url.QueryEscape(console) + "_games&cmlimit=max"
}

func (g gamelistres) urlForContinue(baseURL string) string {
	return baseURL + "&cmcontinue=" + g.Cont.Cont.Cont
}

func (g gamelistres) mustContinue() bool {
	return g.Cont.Cont.Cont != ""		// oi
}

func (g *gamelistres) unsetContinueFlag() {
	g.Cont.Cont.Cont = ""
}

func getGameList(console string) ([]GameListEntry, error) {
	var g gamelistres
	var list []GameListEntry

	baseURL := urlForConsole(console)
	r, err := getWikiAPIData(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of games: %v", err)
	}
	err = json.Unmarshal(r, &g)
	if err != nil {
		return nil, fmt.Errorf("error processing list of games: %v\ndata: %s", err, r)
	}
	list = append(list, g.Games.Games...)
	for g.mustContinue() {
		r, err = getWikiAPIData(g.urlForContinue(baseURL))
		if err != nil {
			return nil, fmt.Errorf("error retrieving partial list of games: %v", err)
		}
		g.unsetContinueFlag()		// unmark flag as json.Unmarshal() won't overwrite it when we're done
		err = json.Unmarshal(r, &g)
		if err != nil {
			return nil, fmt.Errorf("error processing partial list of games: %v\ndata: %s", err, r)
		}
		list = append(list, g.Games.Games...)
	}
	return list, nil
}

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

// test
func main() {
	l, err := getGameList("Mega Drive")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	for _, v := range l {
		fmt.Println(v.Name)
	}
}
