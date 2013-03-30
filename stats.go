// split out 12 september 2012
package main

import (
	"fmt"

	// stats report
	"html/template"
)

type Stats struct {
	nBoxScans			int
	nBoxHave				int
	nBoxGood			int
	nBoxBad				int
	nBoxIncomplete		int
	pBoxHave				float64
	pBoxGood			float64
	pBoxGoodAll			float64
	pBoxBad				float64
	pBoxBadAll			float64
	pBoxIncomplete		float64
	pBoxIncompleteAll		float64
	nMediaScans			int
	nMediaHave			int
	nMediaGood			int
	nMediaBad			int
	nMediaIncomplete		int
	pMediaHave			float64
	pMediaGood			float64
	pMediaGoodAll			float64
	pMediaBad			float64
	pMediaBadAll			float64
	pMediaIncomplete		float64
	pMediaIncompleteAll	float64
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

func (scans ScanSet) GetStats(_filterRegion string) (stats Stats) {
	for _, scan := range scans {
		if scan.Error != nil || scan.HasNoScans {		// we can't really count games without known scans since the stats are for known scans
			continue
		}
		if !filterRegion(scan.Region, _filterRegion) {
			continue
		}
		stats.nBoxScans++
		switch scan.BoxState.State {
		case Good:
			stats.nBoxGood++
			stats.nBoxHave++
		case Bad:
			stats.nBoxBad++
			stats.nBoxHave++
		case Incomplete:
			stats.nBoxIncomplete++
			stats.nBoxHave++
		}
		stats.nMediaScans++
		switch scan.MediaState.State {
		case Good:
			stats.nMediaGood++
			stats.nMediaHave++
		case Bad:
			stats.nMediaBad++
			stats.nMediaHave++
		case Incomplete:
			stats.nMediaIncomplete++
			stats.nMediaHave++
		}
	}
	stats.CalculatePercents()
	return
}

func (stats *Stats) Add(stats2 Stats) {
	stats.nBoxScans += stats2.nBoxScans
	stats.nBoxHave += stats2.nBoxHave
	stats.nBoxGood += stats2.nBoxGood
	stats.nBoxBad += stats2.nBoxBad
	stats.nMediaScans += stats2.nMediaScans
	stats.nMediaHave += stats2.nMediaHave
	stats.nMediaGood += stats2.nMediaGood
	stats.nMediaBad += stats2.nMediaBad
	stats.CalculatePercents()
}

func (stats *Stats) CalculatePercents() {
	stats.pBoxHave = pcnt(stats.nBoxHave, stats.nBoxScans)
	stats.pBoxGood = pcnt(stats.nBoxGood, stats.nBoxHave)
	stats.pBoxGoodAll = pcnt(stats.nBoxGood, stats.nBoxScans)
	stats.pBoxBad = pcnt(stats.nBoxBad, stats.nBoxHave)
	stats.pBoxBadAll = pcnt(stats.nBoxBad, stats.nBoxScans)
	stats.pBoxIncomplete = pcnt(stats.nBoxIncomplete, stats.nBoxHave)
	stats.pBoxIncompleteAll = pcnt(stats.nBoxIncomplete, stats.nBoxScans)

	stats.pMediaHave = pcnt(stats.nMediaHave, stats.nMediaScans)
	stats.pMediaGood = pcnt(stats.nMediaGood, stats.nMediaHave)
	stats.pMediaGoodAll = pcnt(stats.nMediaGood, stats.nMediaScans)
	stats.pMediaBad = pcnt(stats.nMediaBad, stats.nMediaHave)
	stats.pMediaBadAll = pcnt(stats.nMediaBad, stats.nMediaScans)
	stats.pMediaIncomplete = pcnt(stats.nMediaIncomplete, stats.nBoxHave)
	stats.pMediaIncompleteAll = pcnt(stats.nMediaIncomplete, stats.nMediaScans)
}

const pbarMinWidth = 300
const pbarHeight = 20
const pbarBorderThickness = 2

const pbarCSS_base = `
		/* progress bar */
		div.pbar {
			min-width: %dpx;
			height: %dpx;
			border: %dpx solid black;
			background-color: {{missingcolor}};
		}
		span.pbar_good {
			display: inline-block;
			height: 100%%;
			background-color: {{goodcolor}};
		}
		span.pbar_inc {
			display: inline-block;
			height: 100%%;
			background-color: {{incompletecolor}};
		}
		span.pbar_bad {
			display: inline-block;
			height: 100%%;
			background-color: {{badcolor}};
		}
`

var pbarCSS = fmt.Sprintf(pbarCSS_base,
	pbarMinWidth, pbarHeight, pbarBorderThickness)

// this one is not
const pbarHTML = `<div class="pbar">
<span class="pbar_good"
    style="width: %g%%;"></span><span class="pbar_inc"
    style="width: %g%%;"></span><span class="pbar_bad"
    style="width: %g%%;"></span>
</div>`

func progressbar(pGoodAll float64, pBadAll float64, pIncAll float64) template.HTML {
	return template.HTML(fmt.Sprintf(pbarHTML,
		pGoodAll, pIncAll, pBadAll))
}

func (s Stats) BoxProgressBar() template.HTML {
	return progressbar(s.pBoxGoodAll, s.pBoxBadAll, s.pBoxIncompleteAll)
}

func (s Stats) MediaProgressBar() template.HTML {
	return progressbar(s.pMediaGoodAll, s.pMediaBadAll, s.pMediaIncompleteAll)
}

var gameStatsHTML = `<table>
		<tr>
			<th rowspan=4 valign=top align=right>Box</th>
			<td style="white-space: nowrap;">We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are good (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;">%s</td></tr>
		<tr>
			<th rowspan=4 valign=top align=right>Media</th>
			<td style="white-space: nowrap;">We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are good (%.2f%% overall)</td>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;">%s</td></tr>
	</table>`

func (stats Stats) HTML() template.HTML {
	boxbar := stats.BoxProgressBar()
	mediabar := stats.MediaProgressBar()
	return template.HTML(fmt.Sprintf(gameStatsHTML,
		stats.nBoxHave, stats.nBoxScans, stats.pBoxHave,
		stats.nBoxGood, stats.pBoxGood, stats.pBoxGoodAll,
		stats.nBoxBad, stats.pBoxBad, stats.pBoxBadAll,
		boxbar,
		stats.nMediaHave, stats.nMediaScans, stats.pMediaHave,
		stats.nMediaGood, stats.pMediaGood, stats.pMediaGoodAll,
		stats.nMediaBad, stats.pMediaBad, stats.pMediaBadAll,
		mediabar))
}
