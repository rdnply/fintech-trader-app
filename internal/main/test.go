package main

import (
	"cw1/internal/format"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // init postgres driver
	"time"
)



func main() {
	//config := logger.Configuration{
	//	EnableConsole:     true,
	//	ConsoleLevel:      logger.Debug,
	//	ConsoleJSONFormat: true,
	//	EnableFile:        true,
	//	FileLevel:         logger.Info,
	//	FileJSONFormat:    true,
	//	FileLocation:      "log.log",
	//}
	//
	//logger, err := logger.New(config, logger.InstanceZapLogger)
	//if err != nil {
	//	log.Fatal("could not instantiate logger: ", err)
	//}
	//
	//db, err := postgres.New(logger, "C:\\Users\\rodion\\go\\src\\cw1\\configuration.json")
	//if err != nil {
	//	logger.Fatalf("Can't create database instance %v", err)
	//}
	//
	//err = db.CheckConnection()
	//if err != nil {
	//	logger.Fatalf("Can't connect to database %v", err)
	//}
	//
	//userStorage, err := postgres.NewUserStorage(db)
	//if err != nil {
	//	logger.Fatalf("Can't create user storage: %s", err)
	//}

	//u, err := userStorage.FindByID(38)
	//fmt.Println(*u)
	//j, err := json.Marshal(u)
	//if err != nil {
	//	fmt.Errorf("can't marshal object")
	//}
	//
	//fmt.Println(string(j))

	//info := user.NewInfo(u)
	//i, err := json.Marshal(info)
	//fmt.Println(string(i))

	t := format.NullTime{V:&sql.NullTime{time.Now(), true}}
	j, err := t.MarshalJSON()
	if err != nil {
		fmt.Errorf("cant marshal")
	}
	fmt.Println(string(j))
}
