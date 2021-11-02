package main

import (
	"fmt"
)

func main() {
	arr := []string{"I", "am", "stupid", "and", "weak"}

	for k, _ := range arr {
		switch k {
		case 2:
			arr[k] = "smart"
		case 4:
			arr[k] = "strong"
		}
	}

	fmt.Printf("%+v\n", arr)
}
