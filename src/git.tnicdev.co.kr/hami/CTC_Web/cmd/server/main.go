package main

import (
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go main_mobile(wg)
	go main_web(wg)
	wg.Wait()
}
