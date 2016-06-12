package gcache

import (
	"sync"
	"fmt"
	"time"
)

var mu sync.Mutex
var sum = 0
var max = 1000000

func testCase1() {
	start := time.Now().Nanosecond()
	for i:=1; i<max; i++ {
		mu.Lock()
		sum += 1
		mu.Unlock()
	}

	t := time.Now().Nanosecond() - start

	fmt.Println(t)
}

func testCase2() {
	var ch =  make(chan bool,max)
	var wg sync.WaitGroup
	start := time.Now().Nanosecond()
	go func() {
		wg.Add(1)
		for{
			if <-ch {
				sum++
			}
			if sum == max {
				wg.Add(-1)
			}
		}
	}()

	for i:=1; i<max; i++ {
		ch <- true
	}
	wg.Wait()
	t := time.Now().Nanosecond() - start
	fmt.Println(t)
}
