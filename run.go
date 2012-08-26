// 22-24 august 2012
package main

import (
	"fmt"
	"strings"
	"sort"
)

type SortOrder int
const (
	SortByRegion SortOrder = iota
	SortByBoxState
	SortByMediaState
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
		scans, err := GetScans(game, console)
		if err == ErrGameNoScans {		// omit games for this console that will not have scans
			continue
		}
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

type sorter struct {
	scans		ScanSet
	sortOrder		SortOrder
}

// for sort.Interface
func (s sorter) Len() int {
	return len(s.scans)
}

func (s sorter) Less(i, j int) bool {
	scans := s.scans
	// the sort orders make no sense if there is either an error or no scans for a game, so handle those cases first
	if scans[i].Error != nil && scans[j].Error != nil {		// errors go first
		return scans[i].Name < scans[j].Name		// by title if they both error
	}
	if scans[i].Error != nil {
		return true
	}
	if scans[j].Error != nil {
		return false
	}
	if scans[i].HasNoScans && scans[j].HasNoScans {	// then lack of scans
		return scans[i].Name < scans[j].Name		// by title if they both lack scans
	}
	if scans[i].HasNoScans {
		return true
	}
	if scans[j].HasNoScans {
		return false
	}
	switch s.sortOrder {
	case SortByRegion:		// sort by region, then by name
		if scans[i].Region == scans[j].Region {			// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].Region < scans[j].Region			// finally
	case SortByBoxState:
		if scans[i].BoxState == scans[j].BoxState {			// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].BoxState < scans[j].BoxState		// finally
	case SortByMediaState:
		if scans[i].MediaState == scans[j].MediaState {		// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].MediaState < scans[j].MediaState	// finally
	}
	panic(fmt.Sprintf("invalid sort order %d", int(s.sortOrder)))
}

func (s sorter) Swap(i, j int) {
	s.scans[i], s.scans[j] = s.scans[j], s.scans[i]
}

func (scans ScanSet) Sort(so SortOrder) {
	sort.Sort(sorter{
		scans:		scans,
		sortOrder:		so,
	})
}

type Stats struct {
	nBoxScans		int
	nBoxHave			int
	nBoxGood		int
	nBoxBad			int
	pBoxHave			float64
	pBoxGood		float64
	pBoxGoodAll		float64
	pBoxBad			float64
	pBoxBadAll		float64
	nMediaScans		int
	nMediaHave		int
	nMediaGood		int
	nMediaBad		int
	pMediaHave		float64
	pMediaGood		float64
	pMediaGoodAll		float64
	pMediaBad		float64
	pMediaBadAll		float64
}

func pcnt(_a, _b int) float64 {
	if _a != 0 && _b == 0 {	// sanity check
		panic("we somehow have scans where none are expected")
	}
	if _b == 0 {
		return 0.0
	}
	a, b := float64(_a), float64(_b)
	return (a / b) * 100.0
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
			stats.nBoxHave++
		case Bad:
			stats.nBoxBad++
			fallthrough
		case Incomplete:
			stats.nBoxHave++
		}
		stats.nMediaScans++
		switch scan.MediaState {
		case Good:
			stats.nMediaGood++
			stats.nMediaHave++
		case Bad:
			stats.nMediaBad++
			fallthrough
		case Incomplete:
			stats.nMediaHave++
		}
	}
	stats.pBoxHave = pcnt(stats.nBoxHave, stats.nBoxScans)
	stats.pBoxGood = pcnt(stats.nBoxGood, stats.nBoxHave)
	stats.pBoxGoodAll = pcnt(stats.nBoxGood, stats.nBoxScans)
	stats.pBoxBad = pcnt(stats.nBoxBad, stats.nBoxHave)
	stats.pBoxBadAll = pcnt(stats.nBoxBad, stats.nBoxScans)
	stats.pMediaHave = pcnt(stats.nMediaHave, stats.nMediaScans)
	stats.pMediaGood = pcnt(stats.nMediaGood, stats.nMediaHave)
	stats.pMediaGoodAll = pcnt(stats.nMediaGood, stats.nMediaScans)
	stats.pMediaBad = pcnt(stats.nMediaBad, stats.nMediaHave)
	stats.pMediaBadAll = pcnt(stats.nMediaBad, stats.nMediaScans)
	return
}
