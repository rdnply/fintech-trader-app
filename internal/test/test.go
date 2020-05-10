package main

import (
	"cw1/internal/format"
	"encoding/json"
	"fmt"
)

func main() {
	//json := "{\"isfavourite\" : false,\"is_active\": false}"
	//var null format.NullInt64
	var null format.NullInt64
	err := json.Unmarshal([]byte("400"), &null)
	if err != nil {
		fmt.Errorf("error")
	}
	fmt.Println(null)
}
