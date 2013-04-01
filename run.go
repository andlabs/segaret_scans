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
type ScanSets map[string]ScanSet

func Run(consoles Consoles) (ScanSets, error) {
	var gameScans = ScanSets{}
	var gameLists = map[string][]string{}
	var tocover = map[string]map[string]struct{}{}		// tocover[console][game] exists if the game was not yet covered
	var expected = map[string]map[string]string{}		// expected[console][game] becomes the appropriate category name
	var scanboxes []*Scan
	var err error

	// 1) populate game lists
	for category, console := range consoles {
		games, err := GetGameList(category)
		if err != nil {
			return nil, fmt.Errorf("error getting %s list: %v", category, err)
		}
		gameLists[console] = games
		if tocover[console] == nil {			// initialize the first time
			tocover[console] = map[string]struct{}{}
			expected[console] = map[string]string{}
		}
		for _, g := range games {
			tocover[console][g] = struct{}{}
			if expected[console][g] != "" {		// sanity check
				panic(fmt.Sprintf("%s:%s already in %s, want to add to %s",
					console, g, expected[console][g], category))
			}
			expected[console][g] = category
		}
	}

	// 2) get either one console's scanboxes or all scanboxes (the former is an optimization)
	// TODO specialize this to allow only returning the scanboxes for one console
	scanboxes, err = GetAllScanboxes()
	if err != nil {
		return nil, fmt.Errorf("error getting scanboxes: %v", err)		// TODO make a specific error message?
	}

	// 3) get all the scan states for all the known games
	for _, scan := range scanboxes {
		if _, ok := expected[scan.Console][scan.Name]; !ok {			// not expected
			continue
		}
		boxState := scan.BoxScanState()
		mediaState := scan.MediaScanState()
		category := expected[scan.Console][scan.Name]
		gameScans[category] = append(gameScans[category], &GameScan{
			Name:		scan.Name,
			Region:		scan.Region,
			BoxState:		boxState,
			MediaState:	mediaState,
		})
		delete(tocover[scan.Console], scan.Name)
	}

	// 4) check what's left to see if they lack scans or are marked as not having them
	for console, games := range tocover {
		for game := range games {
			markedNoScans, err := sql_getmarkednoscans(game, console)
			if err != nil {
				gameScans[console] = append(gameScans[console], &GameScan{
					Name:	game,
					Error:	err,
				})
				continue
			}
			if !markedNoScans {				// if not marked as no scans, inform viewer we don't have scans
				category := expected[console][game]
				gameScans[category] = append(gameScans[category], &GameScan{
					Name:		game,
					HasNoScans:	true,
				})
				continue
			}
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
