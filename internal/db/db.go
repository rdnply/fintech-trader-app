package db

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"log"
	"os"
)

type Configuration struct {
	HOST     string `json:"host"`
	PORT     string `json:"port"`
	USER     string `json:"user"`
	PASSWORD string `json:"password"`
	DBNAME   string `json:"db_name"`
}

func getConfig(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("unable to read input json file %s, %v", filename, err)
	}
	defer f.Close()

	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("unable to read input json file as a byte array %s, %v", filename, err)
	}

	var c Configuration

	err = json.Unmarshal(byteData, &c)
	if err != nil {
		fmt.Errorf("can't unmarshal json with configuration: %v", err)
	}

	str := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.HOST, c.PORT, c.USER, c.PASSWORD, c.DBNAME)

	return str, nil
}

var (
	db  *gorm.DB
	err error
)

func Init(filename string) {
	conf, err := getConfig(filename)
	if err != nil {
		fmt.Printf("can't get configuration for database: %v", err)
	}

	db, err = gorm.Open("postgres", conf)
	if err != nil {
		log.Fatalf("can't connect to db: %s", err)
	}
}

func GetDBConn() *gorm.DB {
	return db
}
