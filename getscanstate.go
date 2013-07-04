// 22 august 2012
package main

import (
	"fmt"
)

type ScanState int

const (	// in sort order
	Missing ScanState = iota
	Bad
	Incomplete
	Good
)

var goodScanPrefix = []byte("Good")

func checkScanGood(goodscans *GoodScansList, scan string) bool {
	return goodscans.IsGood(scan)
}

func (s ScanState) Join(s2 ScanState) ScanState {
	if s == Bad || s2 == Bad {
		return Bad
	}
	if s == Missing || s2 == Missing {
		// only return Missing if both are Missing; otherwise we mark it as Bad to signal that it's incomplete
		if s == Missing && s2 == Missing {
			return Missing
		}
		return Incomplete
	}
	return Good	// otherwise
}

func checkSingleState(goodscans *GoodScansList, what string) ScanState {
	if what == "" {
		return Missing
	}
	good := checkScanGood(goodscans, what)
	if good {
		return Good
	}
	return Bad
}

func checkBoxSet(goodscans *GoodScansList, cover, front, back, spine string, spineMissing, square bool) ScanState {
	// if cover= is used then we have a single-image cover, so just check that
	if cover != "" {
		return checkSingleState(goodscans, cover)
	}

	frontState := checkSingleState(goodscans, front)
	backState := checkSingleState(goodscans, back)
	spineState := checkSingleState(goodscans, spine)

	// if the spine is missing but SpineMissing is not explicitly set, there is no spine
	if spineState == Missing && !spineMissing {
		return frontState.Join(backState)
	}

	return frontState.Join(backState).Join(spineState)
}

func (s Scan) BoxScanState(goodscans *GoodScansList) ScanState {
	baseState := checkBoxSet(goodscans, s.Cover, s.Front, s.Back, s.Spine, s.SpineMissing, s.Square)
	if s.HasJewelCase {
		baseState = baseState.Join(
			checkBoxSet(goodscans, "", s.JewelCaseFront, s.JewelCaseBack,
				s.JewelCaseSpine, s.JewelCaseSpineMissing,
				true))		// jewel cases are always square
	}
	if s.Spine2 != "" {			// check Spine2, Top, Bottom if we have them
		baseState = baseState.Join(checkSingleState(goodscans, s.Spine2))
	}
	if s.Top != "" {
		baseState = baseState.Join(checkSingleState(goodscans, s.Top))
	}
	if s.Bottom != "" {
		baseState = baseState.Join(checkSingleState(goodscans, s.Bottom))
	}
	return baseState
}

func (s Scan) MediaScanState(goodscans *GoodScansList) ScanState {
	itemsState := func() ScanState {
		state := checkSingleState(goodscans, s.Items[0])
		for i := 1; i < len(s.Items); i++ {
			state = state.Join(checkSingleState(goodscans, s.Items[i]))
		}
		return state
	}

	// neither cart nor disc
	if s.Cart == "" && s.Disc == "" {
		if len(s.Items) > 0 {
			return itemsState()
		}
		return Missing
	}

	// no cart
	if s.Cart == "" {
		discState := checkSingleState(goodscans, s.Disc)
		if len(s.Items) > 0 {
			return discState.Join(itemsState())
		}
		return discState
	}

	// no disc
	if s.Disc == "" {
		cartState := checkSingleState(goodscans, s.Cart)
		if len(s.Items) > 0 {
			return cartState.Join(itemsState())
		}
		return cartState
	}

	// both cart and disc
	state := checkSingleState(goodscans, s.Cart)
	state = state.Join(checkSingleState(goodscans, s.Disc))
	if len(s.Items) > 0 {
		return state.Join(itemsState())
	}
	return state
}

func (s Scan) ManualScanState(goodscans *GoodScansList) ScanState {
	if s.Manual == "" {
		return Missing
	}
	return checkSingleState(goodscans, s.Manual)
}

func (s ScanState) String() string {
	switch s {
	case Missing:
		return "Missing"
	case Bad:
		return "Bad"
	case Incomplete:
		return "Incomplete"
	case Good:
		return "Good"
	}
	panic(fmt.Sprintf("invalid value %d for scan state", int(s)))
}
