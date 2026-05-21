package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"github.com/zaibon/shortcut/db"
	"github.com/zaibon/shortcut/domain"
	"github.com/zaibon/shortcut/log"
	"github.com/zaibon/shortcut/services"
)

var (
	envFile     string
	dryRun      bool
	onlyActive  bool
	scanLimit   int
	scanDelayMs int
)

var scanURLFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "env-file",
		Usage:       "path to env file (e.g. .env-prod) to overload environment variables",
		Value:       "",
		Destination: &envFile,
	},
	&cli.BoolFlag{
		Name:        "dry-run",
		Usage:       "if true, only scans and logs threats without making database changes",
		Value:       true,
		Destination: &dryRun,
	},
	&cli.BoolFlag{
		Name:        "only-active",
		Usage:       "if true, only scans URLs that are active and not archived",
		Value:       true,
		Destination: &onlyActive,
	},
	&cli.IntFlag{
		Name:        "limit",
		Usage:       "maximum number of URLs to scan (0 for unlimited)",
		Value:       0,
		Destination: &scanLimit,
	},
	&cli.IntFlag{
		Name:        "delay",
		Usage:       "delay between API calls in milliseconds",
		Value:       100,
		Destination: &scanDelayMs,
	},
	&cli.StringFlag{
		Name:        "db",
		Usage:       "database connection string",
		Value:       "",
		Destination: &c.DBConnString,
	},
	&cli.StringFlag{
		Name:        "webrisk-key",
		Usage:       "Google Web Risk API Key",
		Value:       "",
		Destination: &c.GoogleWebRiskAPIKey,
	},
}

func runScanURLs(cCtx *cli.Context, cfg config) error {
	// 1. Env Overloading
	if envFile != "" {
		log.Info("Loading environment file", "path", envFile)
		if err := godotenv.Overload(envFile); err != nil {
			return fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	}

	// 2. Fetch config values from env if they were not explicitly passed via CLI flags
	dbString := cfg.DBConnString
	if dbString == "" {
		dbString = os.Getenv("SHORTCUT_DB")
	}

	webRiskAPIKey := cfg.GoogleWebRiskAPIKey
	if webRiskAPIKey == "" {
		webRiskAPIKey = os.Getenv("SHORTCUT_GOOGLE_WEBRISK_API_KEY")
	}

	if dbString == "" {
		return fmt.Errorf("database connection string is empty (specify via --db or SHORTCUT_DB env var)")
	}

	ctx := context.Background()

	// Create a temporary masked config to output redacted connection strings safely
	tempConfig := config{DBConnString: dbString}
	log.Info("Connecting to database...", "dsn", tempConfig.SafeDBString())

	dbPool, err := pgxpool.New(ctx, dbString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	// 3. Fetch active URLs
	query := "SELECT id, title, short_url, long_url, author_id, is_active FROM urls"
	var args []any

	if onlyActive {
		query += " WHERE is_active = true AND is_archived = false"
	}

	query += " ORDER BY id ASC"

	if scanLimit > 0 {
		query += " LIMIT $1"
		args = append(args, scanLimit)
	}

	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	type URLItem struct {
		ID       int32
		Title    string
		ShortURL string
		LongURL  string
		AuthorID int32
		IsActive bool
	}

	var items []URLItem
	for rows.Next() {
		var item URLItem
		if err := rows.Scan(&item.ID, &item.Title, &item.ShortURL, &item.LongURL, &item.AuthorID, &item.IsActive); err != nil {
			return fmt.Errorf("failed to scan URL row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	log.Info("Query completed", "count", len(items))

	// 4. Initialize Safety Scanner & URL Store
	if webRiskAPIKey == "" {
		log.Warn("Google Web Risk API Key is empty. Standard Web Risk API scans will be skipped, running only local heuristics.")
	}

	scanner := services.NewWebRiskScanner(webRiskAPIKey)
	urlStore := db.NewURLStore(dbPool)

	// 5. Run loop and scan URLs
	var scannedCount, flaggedCount, errorCount int

	for i, item := range items {
		log.Info(fmt.Sprintf("[%d/%d] Scanning short_url=%s long_url=%s", i+1, len(items), item.ShortURL, item.LongURL))

		riskScore, threatType, err := scanner.Scan(ctx, item.LongURL)
		scannedCount++
		if err != nil {
			log.Error("Error scanning URL", "url", item.LongURL, "err", err)
			errorCount++
			continue
		}

		if threatType != "" || riskScore > 0 {
			flaggedCount++
			log.Warn("THREAT DETECTED!", "url", item.LongURL, "threat", threatType, "score", riskScore, "author_id", item.AuthorID)

			if dryRun {
				log.Info(fmt.Sprintf("[DRY RUN] Would deactivate URL %s, insert moderation flag, and suspend user ID %d", item.ShortURL, item.AuthorID))
			} else {
				log.Info(fmt.Sprintf("Applying moderation action for URL %s...", item.ShortURL))

				// 1. Deactivate URL
				if err := urlStore.UpdateURLStatus(ctx, item.ShortURL, false); err != nil {
					log.Error("Failed to deactivate URL", "short_url", item.ShortURL, "err", err)
					errorCount++
					continue
				}
				log.Info("Successfully deactivated URL", "short_url", item.ShortURL)

				// 2. Insert Moderation Flag
				if err := urlStore.InsertModerationFlag(ctx, domain.ID(item.ID), domain.ID(item.AuthorID), riskScore, threatType); err != nil {
					log.Error("Failed to insert moderation flag", "url_id", item.ID, "err", err)
					errorCount++
					// Continue anyway since URL is deactivated
				} else {
					log.Info("Successfully logged moderation flag in DB", "url_id", item.ID)
				}

				// 3. Suspend Creator
				if err := urlStore.SuspendUserByID(ctx, domain.ID(item.AuthorID), true); err != nil {
					log.Error("Failed to suspend user", "user_id", item.AuthorID, "err", err)
					errorCount++
				} else {
					log.Info("Successfully suspended offending user account", "user_id", item.AuthorID)
				}
			}
		}

		if scanDelayMs > 0 && i < len(items)-1 {
			time.Sleep(time.Duration(scanDelayMs) * time.Millisecond)
		}
	}

	log.Info("Scan complete!", "total_scanned", scannedCount, "flagged", flaggedCount, "errors", errorCount)
	return nil
}
