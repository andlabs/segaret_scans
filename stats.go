// split out 12 september 2012
package main

import (
	// for drawing the progress bar
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"encoding/base64"
	"bytes"

	// stats report
	"fmt"
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

const pbarWidth = 300
const pbarHeight = 20
const pbarPercentFactor = 3
const pbarBorderThickness = 2

var (
	black = image.NewUniform(color.Black)
	white = image.NewUniform(color_missing)
	red = image.NewUniform(color_bad)
	green = image.NewUniform(color_good)
	yellow = image.NewUniform(color_incomplete)
)

func progressbar(pGoodAll float64, pBadAll float64, pIncAll float64) string {
	pbar := image.NewRGBA(image.Rect(0, 0,
		pbarWidth + (pbarBorderThickness * 2),
		pbarHeight + (pbarBorderThickness * 2)))
	// 1) fill black for border
	draw.Draw(pbar, pbar.Rect, black, image.ZP, draw.Src)
	// 2) draw white for what we have
	draw.Draw(pbar, image.Rect(
		pbarBorderThickness, pbarBorderThickness,
		pbarBorderThickness + pbarWidth,
		pbarBorderThickness + pbarHeight), white, image.ZP, draw.Src)
	// 3) figure out the rectanges for good, incomplete, and bad
	goodWid := int(pGoodAll + 0.5) * pbarPercentFactor
	badWid := int(pBadAll + 0.5) * pbarPercentFactor
	incompleteWid := int(pIncAll + 0.5) * pbarPercentFactor
	goodRect := image.Rect(
		pbarBorderThickness, pbarBorderThickness,
		pbarBorderThickness + goodWid,
		pbarBorderThickness + pbarHeight)
	incRect := image.Rect(
		pbarBorderThickness + goodWid, pbarBorderThickness,
		pbarBorderThickness + goodWid + incompleteWid,
		pbarBorderThickness + pbarHeight)
	badRect := image.Rect(
		pbarBorderThickness + goodWid + incompleteWid, pbarBorderThickness,
		pbarBorderThickness + goodWid + incompleteWid + badWid,
		pbarBorderThickness + pbarHeight)
	// 4) draw good, incomplete, and bad
	draw.Draw(pbar, goodRect, green, image.ZP, draw.Src)
	draw.Draw(pbar, incRect, yellow, image.ZP, draw.Src)
	draw.Draw(pbar, badRect, red, image.ZP, draw.Src)
	// 5) convert to base64 and return
	_pngDat := new(bytes.Buffer)
	pngDat := base64.NewEncoder(base64.StdEncoding, _pngDat)
	defer pngDat.Close()
	err := png.Encode(pngDat, pbar)
	if err != nil {
		panic(fmt.Errorf("error producing progress bar PNG: %v\n", err))
	}
	return _pngDat.String()
}

func (s Stats) BoxProgressBar() string {
	return progressbar(s.pBoxGoodAll, s.pBoxBadAll, s.pBoxIncompleteAll)
}

func (s Stats) MediaProgressBar() string {
	return progressbar(s.pMediaGoodAll, s.pMediaBadAll, s.pMediaIncompleteAll)
}

var gameStatsHTML = `<table>
		<tr>
			<th rowspan=4 valign=top align=right>Box</th>
			<td style="white-space: nowrap;">We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are good (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;"><img src="data:image/png;base64,%s"></td></tr>
		<tr>
			<th rowspan=4 valign=top align=right>Media</th>
			<td style="white-space: nowrap;">We have <b>%d</b> of %d known scans (%.2f%%)</td>
		</tr>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are good (%.2f%% overall)</td>
		<tr><td style="white-space: nowrap;">%d (%.2f%%) of them are bad (%.2f%% overall)</td></tr>
		<tr><td style="white-space: nowrap;"><img src="data:image/png;base64,%s"></td></tr>
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
