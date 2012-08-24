// 22 august 2012
package main

import (
	"fmt"
	"strings"
)

func GetGameList(console string) ([]string, error) {
	s, err := sql_getgames(console)
	if err != nil {
		return nil, fmt.Errorf("error getting %s games list: %v", console, err)
	}
	// turn _ to space for human readability
	for i := 0; i < len(s); i++ {
		if s[i] != "C_So!" {		// except for the one game that actually does have an underscore in its name (or does it? TODO...)
			s[i] = strings.Replace(s[i], "_", " ", -1)
		}
	}
	return s, nil
}

/*
// test
func main() {
	l, err := GetGameList("Mega Drive")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	for _, v := range l {
		fmt.Println(v)
	}
}
*/
