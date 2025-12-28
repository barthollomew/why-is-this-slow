package main

import "time"

func main() {
	target := time.Now().Add(200 * time.Millisecond)
	x := 0
	for time.Now().Before(target) {
		x++
	}
	if x == 0 {
		panic("unreachable")
	}
}
