package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/zaibon/shortcut/db/datastore"
	"github.com/zaibon/shortcut/services/geoip"
)

//TODO zaibon: This is noe needed anymore, might need to clean it up

func init() {
	// goose.AddMigrationNoTxContext(
	// 	pgxMigrateFunc(upPopulateGeoip),
	// 	downPopulateGeoip,
	// )
}

func upPopulateGeoip(ctx context.Context, store datastore.Querier) error {
	rows, err := store.ListVisits(ctx)
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
		loc, err := geoip.Locate(host)
		if err != nil {
			log.Printf("failed to get country for ip %s: %v", host, err)
			continue
		}
		if loc.CountryCode == "" {
			log.Printf("no country found for ip %s skipping", host)
			continue
		}

		_, err = store.InsertVisitLocation(ctx, datastore.InsertVisitLocationParams{
			VisitID: row.ID,
			Address: row.IpAddress,
			CountryCode: pgtype.Text{
				String: loc.CountryCode,
				Valid:  loc.CountryCode != "",
			},
			CountryName: pgtype.Text{
				String: loc.CountryName,
				Valid:  loc.CountryName != "",
			},
			Subdivision: pgtype.Text{
				String: loc.Subdivision,
				Valid:  loc.Subdivision != "",
			},
			Continent: pgtype.Text{
				String: loc.Continent,
				Valid:  loc.Continent != "",
			},
			CityName: pgtype.Text{
				String: loc.CityName,
				Valid:  loc.CityName != "",
			},
			Latitude: pgtype.Float8{
				Float64: loc.Latitude,
				Valid:   loc.Latitude != 0,
			},
			Longitude: pgtype.Float8{
				Float64: loc.Longitude,
				Valid:   loc.Longitude != 0,
			},
			Source: pgtype.Text{
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

func downPopulateGeoip(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM visit_locations"); err != nil {
		if err := tx.Rollback(); err != nil {
			slog.Error("failed to rollback transaction", "err", err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
