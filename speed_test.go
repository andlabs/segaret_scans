// 1 september 2012
package main

import (
	"testing"
)

const console = "Mega_Drive"
const filename = "ThunderForce4_MD_JP_Box.jpg"

func BenchmarkFileCheck(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := checkScanGood(filename)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAllFiles(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, n := range filenames {
			_, err := checkScanGood(n)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkPageList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetGameList(console)
		if err != nil {
			b.Fatal(err)
		}
	}
}
