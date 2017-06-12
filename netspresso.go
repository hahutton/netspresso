package main

import (
	"flag"
	"fmt"
	"sync"
)

var cosmosAccountName string
var generatorCount int
var wg sync.WaitGroup

func init() {
	flag.StringVar(&cosmosAccountName, "account", "na", "the cosmosdb account name")
	flag.StringVar(&cosmosAccountName, "a", "na", "the cosmosdb account name")
	flag.IntVar(&generatorCount, "c", 0, "the count of generators")
}

func main() {
	flag.Parse()
	SetRegions()
	CurrentWriteRegion()
	ListDbs()
	fmt.Println(generatorCount)
	for i := 0; i < generatorCount; i++ {
		fmt.Println(i)
		wg.Add(1)
		go Generator(&wg)
	}
	wg.Wait()
}
