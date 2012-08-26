// 22-24 august 2012
package main

import (
	"fmt"
	"strings"
)

type GameScan struct {
	Name		string
	HasNoScans	bool			// for whole games
	Region		string
	BoxState		ScanState
	MediaState	ScanState
	Error			error
}

type ScanSet []*GameScan

func getMediaState(scan Scan) ScanState {
	if scan.Cart == "" && scan.Disc == "" {
		return Missing
	}
	if scan.Cart != "" && scan.Disc == "" {
		return scan.CartScanState()
	}
	if scan.Cart == "" && scan.Disc != "" {
		return scan.DiscScanState()
	}
	return scan.CartScanState().Join(scan.DiscScanState())	// else
}


func GetConsoleScans(console string) (ScanSet, error) {
	var gameScans ScanSet

	games, err := GetGameList(console)
	if err != nil {
		return nil, fmt.Errorf("error getting %s game list: %v", console, err)
	}
	for _, game := range games {
//fmt.Println(game)
		if strings.HasPrefix(game, "List of " + console + " games") {	// omit list from report
			continue
		}
		scans, err := GetScans(game)
		if err != nil {
			gameScans = append(gameScans, &GameScan{
				Name:	game,
				Error:	err,
			})
			continue
		}
		if len(scans) == 0 {				// there are no scans at all
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				HasNoScans:	true,
			})
			continue
		}
		nScans := 0
		for _, scan := range scans {
			var mediaState ScanState

			if scan.Console != console {	// omit scans from other consoles
				continue
			}
			nScans++
			boxState := scan.BoxScanState()
			mediaState = getMediaState(scan)
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				Region:		scan.Region,
				BoxState:		boxState,
				MediaState:	mediaState,
			})
		}
		if nScans == 0 {					// there are no scans for the specified console
			gameScans = append(gameScans, &GameScan{
				Name:		game,
				HasNoScans:	true,
			})
			continue
		}
	}
	return gameScans, nil
}

type Stats struct {
	nBoxScans	int
	nBoxHave		int
	nBoxGood	int
	nMediaScans	int
	nMediaHave	int
	nMediaGood	int
}

func (scans ScanSet) GetStats(filterRegion string) (stats Stats) {
	for _, scan := range scans {
		if scan.Error != nil || scan.HasNoScans {		// TODO really skip entries without scans?
			continue
		}
		if filterRegion != "" &&
			!strings.HasPrefix(scan.Region, filterRegion) {
			continue
		}
		stats.nBoxScans++
		switch scan.BoxState {
		case Good:
			stats.nBoxGood++
			fallthrough
		case Bad, Incomplete:
			stats.nBoxHave++
		}
		stats.nMediaScans++
		switch scan.MediaState {
		case Good:
			stats.nMediaGood++
			fallthrough
		case Bad, Incomplete:
			stats.nMediaHave++
		}
	}
	return
}
