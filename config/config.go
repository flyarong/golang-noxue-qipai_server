package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var Config Conf

type Conf struct {
	AppKey string
	Db Db
}

type Db struct {
	Url string
}

func init() {

	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Panicln(err.Error())
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Panicln(err.Error())
	}
}
