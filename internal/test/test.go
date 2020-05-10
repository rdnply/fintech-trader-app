package main

import (
	"fmt"
	"time"
)

func f() {
	for i := 0; i < 1000; i++ {
		fmt.Println(i)
	}
}

func main() {
	fmt.Println("START")
	go f()
	fmt.Println("END")
	time.Sleep(100000)
}
