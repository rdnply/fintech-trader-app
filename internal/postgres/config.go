package postgres

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type Configuration struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

func ParseConfig(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "unable to read input json file: "+filename)
	}

	defer f.Close()

	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.Wrap(err, "unable to read input json file as a byte array: "+filename)
	}

	var c Configuration

	err = json.Unmarshal(byteData, &c)
	if err != nil {
		return "", errors.Wrap(err, "can't unmarshal json with configuration")
	}

	URL := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)

	return URL, nil
}
