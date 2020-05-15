package main

import (
	"fmt"
	"time"
)

const layout = "2006-01-02T15:04:05Z"

func main() {
	s := "2022-10-02T22:00:00Z"
	t, err := time.Parse( layout, s)
	if err != nil {
		fmt.Errorf("%v", err)
	}

	fmt.Println("Normal: ", t)
	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Errorf("%v", err)
	}


	fmt.Println("Loc: ", t.In(loc))
}
