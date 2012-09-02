// 22 august 2012
package main

import (
	"fmt"
	"strings"
	"log"
)

type ScanState int
const (	// in sort order
	Error ScanState = iota
	Missing
	Bad
	Incomplete
	Good
)

func checkScanGood(scan string) (bool, error) {
	categories, err := sql_getcatlist(scan)
	if err != nil {
		return false, fmt.Errorf("error processing scan: %v", err)
	}
	for _, v := range categories {
		if strings.HasPrefix(v, "Good") {
			return true, nil
		}
	}
	return false, nil
}

// TODO better name
func (s ScanState) Join(s2 ScanState) ScanState {
	if s == Error || s2 == Error {
		return Error
	}
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
	return Good		// otherwise
}

func checkSingleState(what string) ScanState {
	if what == "" {
		return Missing
	}
	good, err := checkScanGood(what)
	if err != nil {
log.Println(err)
		return Error	// TODO provide a way to show error information
	}
	if good {
		return Good
	}
	return Bad
}

func checkBoxSet(front, back, spine string, spineMissing, square bool) ScanState {
	// if there is no back or spine and the cover is not square (like jewel cases), then the cover is a single piece cover (like clamshell Mega Drive games)
	if back == "" && spine == "" && !square {
		return checkSingleState(front)
	}

	frontState := checkSingleState(front)
	backState := checkSingleState(back)
	spineState := checkSingleState(spine)

	// if the spine is missing but SpineMissing is not explicitly set, there is no spine
	if spineState == Missing && !spineMissing {
		return frontState.Join(backState)
	}

	return frontState.Join(backState).Join(spineState)
}

func (s Scan) BoxScanState() ScanState {
	baseState := checkBoxSet(s.Front, s.Back, s.Spine, s.SpineMissing, s.Square)
	if s.HasJewelCase {
		baseState = baseState.Join(
			checkBoxSet(s.JewelCaseFront, s.JewelCaseBack,
				s.JewelCaseSpine, s.JewelCaseSpineMissing,
				true))		// jewel cases are always square
	}
	return baseState
}

func (s Scan) MediaScanState() ScanState {
	itemsState := func() ScanState {
		state := checkSingleState(s.Items[0])
		for i := 1; i < len(s.Items); i++ {
			state = state.Join(checkSingleState(s.Items[i]))
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
		discState := checkSingleState(s.Disc)
		if len(s.Items) > 0 {
			return discState.Join(itemsState())
		}
		return discState
	}

	// no disc
	if s.Disc == "" {
		cartState := checkSingleState(s.Cart)
		if len(s.Items) > 0 {
			return cartState.Join(itemsState())
		}
		return cartState
	}

	// both cart and disc
	state := checkSingleState(s.Cart)
	state = state.Join(checkSingleState(s.Disc))
	if len(s.Items) > 0 {
		return state.Join(itemsState())
	}
	return state
}

// for testing
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
	case Error:
		return "Error"
	}
	panic(fmt.Sprintf("invalid value %d for scan state", int(s)))
}
