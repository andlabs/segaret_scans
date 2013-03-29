// 22-24 august 2012
package main

import (
	"fmt"
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

func GetConsoleScans(console string) (ScanSet, error) {
	return Run(console + " games", console)
}

func GetAlbumScans() (ScanSet, error) {
	return Run("Albums", "CD")
}

func Run(category string, console string) (ScanSet, error) {
	var gameScans ScanSet

	games, err := GetGameList(category)
	if err != nil {
		return nil, fmt.Errorf("error getting %s list: %v", category, err)
	}
	for _, game := range games {
//fmt.Println(game)
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

			nScans++
			boxState := scan.BoxScanState()
			mediaState = scan.MediaScanState()
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
		if scans[i].Region == scans[j].Region {					// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].Region < scans[j].Region					// finally
	case SortByBoxState:
		if scans[i].BoxState.State == scans[j].BoxState.State {		// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].BoxState.State < scans[j].BoxState.State		// finally
	case SortByMediaState:
		if scans[i].MediaState.State == scans[j].MediaState.State {		// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].MediaState.State < scans[j].MediaState.State	// finally
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
