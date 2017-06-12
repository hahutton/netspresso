package main

import "fmt"

func RunWeightedRandom() {
	var choice = RandIntn(100)
	fmt.Println(choice)

	switch {
	case choice < 30:
		PutDocument()
	default:
		ReadRemote()
	}
}
