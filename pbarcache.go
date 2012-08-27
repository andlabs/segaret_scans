// 26 august 2012
package main

import (
	"time"
)

type pbarRequest struct {
	console	string
	mode	string		// "box" or "media"
	response	chan<- string
}

var requests = make(chan pbarRequest)

const pbarCacheUpdateInterval = 10 * time.Minute

func pbarCache() {
	var consoles = make(map[string]int)
	var boxes, medias = []string{}, []string{}

	get := func(console string) (string, string, error) {
		ss, err := GetConsoleScans(console)
		if err != nil {
			return "", "", err
		}
		stats := ss.GetStats("")
		return stats.BoxProgressBar(), stats.MediaProgressBar(), nil
	}
	add := func(console string) int {
		box, media, err := get(console)
		if err != nil {
			panic(err)
		}
		consoles[console] = len(boxes)
		boxes = append(boxes, box)
		medias = append(medias, media)
		return consoles[console]
	}

	updateAlarm := time.Tick(pbarCacheUpdateInterval)
	for {
		select {
		case <-updateAlarm:
			for c, i := range consoles {
				if box, media, err := get(c); err == nil {
					boxes[i], medias[i] = box, media
				}
			}
		case req := <-requests:
			i, ok := consoles[req.console]
			if !ok {
				i = add(req.console)
			}
			switch req.mode {
			case "box":
				req.response <- boxes[i]
			case "media":
				req.response <- medias[i]
			default:
				panic("unknown request type " + req.mode)
			}
		}
	}
}

func requestBar(console string, mode string) string {
	response := make(chan string)
	defer close(response)
	requests <- pbarRequest{
		console:		console,
		mode:		mode,
		response:		response,
	}
	return <-response
}

func RequestBoxBar(console string) string {
	return requestBar(console, "box")
}

func RequestMediaBar(console string) string {
	return requestBar(console, "media")
}
