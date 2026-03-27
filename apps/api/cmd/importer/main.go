package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/importer"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load .env file if it exists (try multiple paths)
	loadEnvFile(".env")
	loadEnvFile("../../.env")

	ctx := context.Background()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/filmophilia?sslmode=disable"
	}

	// Connect to Database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to database successfully!")

	queries := db.New(pool)

	// Open CSV
	f, err := os.Open("movies.csv")
	if err != nil {
		log.Fatalf("Failed to open movies.csv: %v", err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	total := len(lines) - 1 // Exclude header
	success := 0
	failed := 0

	fmt.Printf("Found %d movies to import\n\n", total)

	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}

		title := strings.TrimSpace(line[0])
		if title == "" {
			fmt.Printf("[%d/%d] Skipping empty title\n", i, total)
			failed++
			continue
		}

		year := 0
		if len(line) > 1 && line[1] != "" {
			year, _ = strconv.Atoi(strings.TrimSpace(line[1]))
		}

		fmt.Printf("[%d/%d] Processing: %s (%d)... ", i, total, title, year)

		tmdb, credits, omdb, err := importer.FetchFullMovieData(title, year)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failed++
			continue
		}

		// Slug generation - clean up special characters
		slug := generateSlug(tmdb.Title, year)

		// Parse release date
		var releaseDate pgtype.Date
		if tmdb.ReleaseDate != "" {
			if t, err := time.Parse("2006-01-02", tmdb.ReleaseDate); err == nil {
				releaseDate = pgtype.Date{Time: t, Valid: true}
			}
		}

		// Parse ratings from OMDB
		var imdbRating pgtype.Numeric
		var rottenTomatoes pgtype.Int4
		var metacriticScore pgtype.Int4

		if omdb != nil {
			if omdb.ImdbRating != "" && omdb.ImdbRating != "N/A" {
				if val, err := strconv.ParseFloat(omdb.ImdbRating, 64); err == nil {
					imdbRating.Scan(fmt.Sprintf("%.1f", val))
				}
			}
			if omdb.Metascore != "" && omdb.Metascore != "N/A" {
				if val, err := strconv.Atoi(omdb.Metascore); err == nil {
					metacriticScore = pgtype.Int4{Int32: int32(val), Valid: true}
				}
			}
			// Parse Rotten Tomatoes from Ratings array
			for _, rating := range omdb.Ratings {
				if rating.Source == "Rotten Tomatoes" {
					rtStr := strings.TrimSuffix(rating.Value, "%")
					if val, err := strconv.Atoi(rtStr); err == nil {
						rottenTomatoes = pgtype.Int4{Int32: int32(val), Valid: true}
					}
				}
			}
		}

		// Check if movie already exists by TMDB ID
		var movieID int32
		existingMovie, err := queries.GetMovieByTmdbID(ctx, pgtype.Int4{Int32: int32(tmdb.ID), Valid: true})
		if err == nil {
			// Movie exists, use its ID
			movieID = existingMovie.ID
			fmt.Printf("(exists) ")
		} else {
			// Create new movie
			movieID, err = queries.CreateMovie(ctx, db.CreateMovieParams{
				Title:           tmdb.Title,
				Slug:            slug,
				Overview:        pgtype.Text{String: tmdb.Overview, Valid: tmdb.Overview != ""},
				PosterUrl:       pgtype.Text{String: "https://image.tmdb.org/t/p/w500" + tmdb.PosterPath, Valid: tmdb.PosterPath != ""},
				ReleaseDate:     releaseDate,
				Runtime:         pgtype.Int4{Int32: int32(tmdb.Runtime), Valid: tmdb.Runtime > 0},
				ImdbID:          pgtype.Text{String: tmdb.IMDBID, Valid: tmdb.IMDBID != ""},
				TmdbID:          pgtype.Int4{Int32: int32(tmdb.ID), Valid: true},
				ImdbRating:      imdbRating,
				RottenTomatoes:  rottenTomatoes,
				MetacriticScore: metacriticScore,
			})
			if err != nil {
				fmt.Printf("FAILED to insert: %v\n", err)
				failed++
				continue
			}
		}

		// Handle Crew (Directors)
		directorOrder := 0
		for _, person := range credits.Crew {
			if person.Job == "Director" {
				pID, err := queries.UpsertPerson(ctx, db.UpsertPersonParams{
					Name: person.Name,
					Slug: generateSlug(person.Name, 0),
				})
				if err != nil {
					continue
				}

				queries.UpsertCredit(ctx, db.UpsertCreditParams{
					MovieID:    movieID,
					PersonID:   pID,
					Department: db.DepartmentDIRECTING,
					Role:       "Director",
					Character:  pgtype.Text{},
					Order:      pgtype.Int4{Int32: int32(directorOrder), Valid: true},
				})
				directorOrder++
			}
		}

		// Handle Cast (top 10 actors)
		maxActors := 10
		if len(credits.Cast) < maxActors {
			maxActors = len(credits.Cast)
		}
		for i := 0; i < maxActors; i++ {
			actor := credits.Cast[i]
			pID, err := queries.UpsertPerson(ctx, db.UpsertPersonParams{
				Name: actor.Name,
				Slug: generateSlug(actor.Name, 0),
			})
			if err != nil {
				continue
			}

			queries.UpsertCredit(ctx, db.UpsertCreditParams{
				MovieID:    movieID,
				PersonID:   pID,
				Department: db.DepartmentACTING,
				Role:       "Actor",
				Character:  pgtype.Text{String: actor.Character, Valid: actor.Character != ""},
				Order:      pgtype.Int4{Int32: int32(i), Valid: true},
			})
		}

		fmt.Printf("OK\n")
		success++

		// Small delay to avoid rate limiting
		time.Sleep(250 * time.Millisecond)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Import completed!\n")
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total: %d\n", total)
}

// generateSlug creates a URL-friendly slug from a title
func generateSlug(title string, year int) string {
	slug := strings.ToLower(title)
	// Replace common special characters
	replacer := strings.NewReplacer(
		" ", "-",
		"'", "",
		"'", "",
		":", "",
		",", "",
		".", "",
		"!", "",
		"?", "",
		"&", "and",
		"/", "-",
		"(", "",
		")", "",
	)
	slug = replacer.Replace(slug)
	// Remove multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if year > 0 {
		slug = fmt.Sprintf("%s-%d", slug, year)
	}
	return slug
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // File doesn't exist, skip
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" { // Don't override existing env vars
				os.Setenv(key, value)
			}
		}
	}
}
