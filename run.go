// 22-24 august 2012
package main

import (
	"fmt"
	"sort"
)

type SortOrder int
const (
	SortByName SortOrder = iota
	SortByRegion
	SortByBoxState
	SortByMediaState
	SortByManualState
)

type GameScan struct {
	Name		string
	HasNoScans	bool			// for whole games
	Region		string
	BoxState		ScanState
	MediaState	ScanState
	ManualState	ScanState
	Error			error
}

type ScanSet []*GameScan
type ScanSets map[string]ScanSet

func RunOne(sql *SQL, category string) (ScanSet, error) {
	if console, ok := config.Consoles[category]; ok {
		m := Consoles{
			category:		console,
		}
		ss, err := Run(sql, m)
		if err != nil {
			return nil, err
		}
		return ss[category], nil
	}
	return nil, fmt.Errorf("unknown category %s; is it in the configuration file? is there a typo?", category)
}

func Run(sql *SQL, consoles Consoles) (ScanSets, error) {
	var gameScans = ScanSets{}
	var gameLists = map[string][]string{}
	var tocover = map[string]map[string]struct{}{}		// tocover[console][game] exists if the game was not yet covered
	var expected = map[string]map[string]string{}		// expected[console][game] becomes the appropriate category name
	var scanboxes []*Scan
	var err error

	// 1) populate game lists
	for category, console := range consoles {
		games, err := sql.GetGameList(category)
		if err != nil {
			return nil, fmt.Errorf("error getting %s list: %v", category, err)
		}
		gameScans[category] = nil			// even if there's nothing in the category, make the entry present in the map
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
	scanboxes, err = GetAllScanboxes(sql)
	if err != nil {
		return nil, fmt.Errorf("error getting scanboxes: %v", err)		// TODO make a specific error message?
	}

	// 3) get all good scans
	goodscans, err := sql.GetAllGoodScans()
	if err != nil {
		return nil, fmt.Errorf("error getting all good scans: %v", err)	// TODO make a specific error message?
	}

	// 3) get all the scan states for all the known games
	for _, scan := range scanboxes {
		if _, ok := expected[scan.Console][scan.Name]; !ok {			// not expected
			continue
		}
		boxState := scan.BoxScanState(goodscans)
		mediaState := scan.MediaScanState(goodscans)
		manualState := scan.ManualScanState(goodscans)
		category := expected[scan.Console][scan.Name]
		gameScans[category] = append(gameScans[category], &GameScan{
			Name:		scan.Name,
			Region:		scan.Region,
			BoxState:		boxState,
			MediaState:	mediaState,
			ManualState:	manualState,
		})
		delete(tocover[scan.Console], scan.Name)
	}

	// 4) check what's left to see if they lack scans or are marked as not having them
	for console, games := range tocover {
		for game := range games {
			markedNoScans, err := sql.GetMarkedNoScans(game, console)
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
	// if we sort by name, we ALWAYS sort by name, even if there are errors or no scans
	// TODO make sure this behaves properly in the face of multiple game names with different region values (previously it just reported regions in whatever order they were in on the page...)
	if s.sortOrder == SortByName {
		return scans[i].Name < scans[j].Name
	}
	// the other sort orders make no sense if there is either an error or no scans for a game, so handle those cases first
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
		if scans[i].Region == scans[j].Region {				// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].Region < scans[j].Region				// finally
	case SortByBoxState:
		if scans[i].BoxState == scans[j].BoxState {				// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].BoxState < scans[j].BoxState			// finally
	case SortByMediaState:
		if scans[i].MediaState == scans[j].MediaState {			// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].MediaState < scans[j].MediaState		// finally
	case SortByManualState:
		if scans[i].ManualState == scans[j].ManualState {		// then if they have the same region, alphabetically
			return scans[i].Name < scans[j].Name
		}
		return scans[i].ManualState < scans[j].ManualState		// finally
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
