package main

import (
	"cw1/internal/db"
	"log"
	"net/http"
)



func main() {
	db.Init("configuration.json")
	db := db.GetDBConn()
	defer db.Close()

	if err := http.ListenAndServe(":5000", r); err != nil {
		log.Fatal(err)
	}
}
