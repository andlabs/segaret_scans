// 22 august 2012
package main

func GetGameList(console string) ([]string, error) {
	return sql_getgames(console)
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
