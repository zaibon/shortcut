package geoip

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"cloud.google.com/go/storage"

	_ "embed"

	geo "github.com/oschwald/geoip2-golang"
)

//go:embed GeoLite2-City.mmdb
var b []byte

var db *geo.Reader

func init() {
	var err error
	db, err = geo.FromBytes(b)
	if err != nil {
		log.Printf("failed to load geoip database: %v", err)
		return
	}
	log.Printf("geoip database loaded")
}

func isDBLoaded() bool {
	return db != nil
}

func DownloadGeoIPDB(bucket, dbFile string) error {
	if isDBLoaded() {
		log.Printf("geoip database already loaded")
		return nil
	}

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create storage client: %w", err)
	}

	r, err := client.Bucket(bucket).Object(dbFile).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to create storage reader: %w", err)
	}

	buf := bytes.Buffer{}

	if _, err := io.Copy(&buf, r); err != nil {
		_ = r.Close()
		return fmt.Errorf("unable to copy file: %w", err)
	}
	_ = r.Close()

	db, err = geo.FromBytes(buf.Bytes())
	if err != nil {
		log.Printf("failed to load geoip database: %v", err)
	}
	buf.Reset()

	log.Printf("geoip database loaded")

	return nil
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

func Locate(ip string) (IPLocation, error) {
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
