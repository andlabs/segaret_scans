// 22 august 2012
package main

import (
	"fmt"
	"strings"
)

type ScanState struct {
	State		int
	Error		error
}

const (	// in sort order
	Error		= iota
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

func SS(x int) ScanState {
	return ScanState{
		State:	x,
	}
}

func (_s ScanState) Join(_s2 ScanState) ScanState {
	s := _s.State
	s2 := _s2.State

	if s == Error {
		return _s
	}
	if s2 == Error {
		return _s2
	}
	if s == Bad || s2 == Bad {
		return SS(Bad)
	}
	if s == Missing || s2 == Missing {
		// only return Missing if both are Missing; otherwise we mark it as Bad to signal that it's incomplete
		if s == Missing && s2 == Missing {
			return SS(Missing)
		}
		return SS(Incomplete)
	}
	return SS(Good)	// otherwise
}

func checkSingleState(what string) ScanState {
	if what == "" {
		return SS(Missing)
	}
	good, err := checkScanGood(what)
	if err != nil {
		return ScanState{
			State:	Error,
			Error:	err,
		}
	}
	if good {
		return SS(Good)
	}
	return SS(Bad)
}

func checkBoxSet(cover, front, back, spine string, spineMissing, square bool) ScanState {
	// if cover= is used then we have a single-image cover, so just check that
	if cover != "" {
		return checkSingleState(cover)
	}

	frontState := checkSingleState(front)
	backState := checkSingleState(back)
	spineState := checkSingleState(spine)

	// if the spine is missing but SpineMissing is not explicitly set, there is no spine
	if spineState.State == Missing && !spineMissing {
		return frontState.Join(backState)
	}

	return frontState.Join(backState).Join(spineState)
}

func (s Scan) BoxScanState() ScanState {
	baseState := checkBoxSet(s.Cover, s.Front, s.Back, s.Spine, s.SpineMissing, s.Square)
	if s.HasJewelCase {
		baseState = baseState.Join(
			checkBoxSet("", s.JewelCaseFront, s.JewelCaseBack,
				s.JewelCaseSpine, s.JewelCaseSpineMissing,
				true))		// jewel cases are always square
	}
	if s.Spine2 != "" {			// check Spine2, Top, Bottom if we have them
		baseState = baseState.Join(checkSingleState(s.Spine2))
	}
	if s.Top != "" {
		baseState = baseState.Join(checkSingleState(s.Top))
	}
	if s.Bottom != "" {
		baseState = baseState.Join(checkSingleState(s.Bottom))
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
		return SS(Missing)
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

func (s ScanState) String() string {
	switch s.State {
	case Missing:
		return "Missing"
	case Bad:
		return "Bad"
	case Incomplete:
		return "Incomplete"
	case Good:
		return "Good"
	case Error:
		return "Error: " + s.Error.Error()
	}
	panic(fmt.Sprintf("invalid value %d for scan state", int(s.State)))
}

func (s ScanState) TypeString() string {
	switch s.State {
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
	panic(fmt.Sprintf("invalid value %d for scan state", int(s.State)))
}
