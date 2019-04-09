package utils

import (
	"github.com/ipipdotnet/ipdb-go"
	"log"
)

var db *ipdb.BaseStation

func initIpIp() {
	var err error
	db, err = ipdb.NewBaseStation("ipipfree.ipdb")
	if err != nil {
		log.Fatal(err)
	}
}

func GetAddress(ip string) string {
	ips, err := db.FindMap(ip, "CN")
	if err != nil {
		return "unknown"
	}

	return ips["country_name"] + " " + ips["region_name"] + " " + ips["city_name"]
}
