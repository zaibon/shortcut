package migrations

import (
	"context"
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/services/geoip"
)

func init() {
	goose.AddMigrationContext(upPopulateGeoip, downPopulateGeoip)
}

func upPopulateGeoip(ctx context.Context, tx *sql.Tx) error {
	db := datastore.New(tx)

	rows, err := db.ListVisits(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if !row.IpAddress.Valid {
			continue
		}

		// host, _, err := net.SplitHostPort(row.IpAddress.String)
		// if err != nil {
		// 	log.Printf("failed to split host and port for ip %s: %v", row.IpAddress.String, err)
		// 	continue
		// }
		host := row.IpAddress.String

		log.Printf("search location for ip %s", host)
		loc, err := geoip.Country(host)
		if err != nil {
			log.Printf("failed to get country for ip %s: %v", host, err)
			continue
		}
		if loc.CountryCode == "" {
			log.Printf("no country found for ip %s skipping", host)
			continue
		}

		_, err = db.InsertVisitLocation(ctx, datastore.InsertVisitLocationParams{
			VisitID:     row.ID,
			Address:     row.IpAddress,
			CountryCode: loc.CountryCode,
			CountryName: sql.NullString{
				String: loc.CountryName,
				Valid:  loc.CountryName != "",
			},
			Subdivision: sql.NullString{
				String: loc.Subdivision,
				Valid:  loc.Subdivision != "",
			},
			Continent: sql.NullString{
				String: loc.Continent,
				Valid:  loc.Continent != "",
			},
			CityName: sql.NullString{
				String: loc.CityName,
				Valid:  loc.CityName != "",
			},
			Latitude: sql.NullFloat64{
				Float64: loc.Latitude,
				Valid:   loc.Latitude != 0,
			},
			Longitude: sql.NullFloat64{
				Float64: loc.Longitude,
				Valid:   loc.Longitude != 0,
			},
			Source: sql.NullString{
				String: loc.Source,
				Valid:  loc.Source != "",
			},
		})
		if err != nil {
			return err
		}

	}

	return nil
}

func downPopulateGeoip(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM visit_locations")
	return err
}
