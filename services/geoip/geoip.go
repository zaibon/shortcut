package geoip

import (
	"fmt"
	"log"
	"net"
	"os"

	_ "embed"

	geo "github.com/oschwald/geoip2-golang"
)

var db *geo.Reader

func init() {
	path := "GeoLite2-City.mmdb"
	if _, err := os.Stat(path); err == nil {
		db, err = geo.Open(path)
		if err != nil {
			log.Printf("failed to load geoip database: %v", err)
		}
	}
}

type IPLocation struct {
	Address     string  `json:"address"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	Subdivision string  `json:"subdivision"`
	Continent   string  `json:"continent"`
	CityName    string  `json:"city_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Source      string  `json:"source"`
}

var ErrInvalidIP = fmt.Errorf("invalid ip address")

func Country(ip string) (IPLocation, error) {
	if db == nil {
		return IPLocation{}, fmt.Errorf("geoip database not loaded")
	}

	ipobj := net.ParseIP(ip)
	if ipobj == nil {
		return IPLocation{}, ErrInvalidIP
	}

	record, err := db.City(ipobj)
	if err != nil {
		return IPLocation{}, fmt.Errorf("failed to get country for ip %s: %w", ip, err)
	}

	loc := IPLocation{
		Address:     ip,
		CountryCode: record.Country.IsoCode,
		CountryName: "Unknown",
		Subdivision: "Unknown",
		Continent:   "Unknown",
		CityName:    "Unknown",
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		Source:      "maxmind geolite2", // for credits
	}

	if val, ok := record.Country.Names["en"]; ok {
		loc.CountryName = val
	}

	if val, ok := record.Continent.Names["en"]; ok {
		loc.Continent = val
	}

	if val, ok := record.City.Names["en"]; ok {
		loc.CityName = val
	}

	if len(record.Subdivisions) > 0 {
		loc.Subdivision = record.Subdivisions[0].Names["en"]
	}

	return loc, nil
}
