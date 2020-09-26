package context

import "fmt"

func checkErr(e error) bool {
	if e != nil {
		fmt.Println("Error:", e.Error())
		return true
	}
	return false
}
