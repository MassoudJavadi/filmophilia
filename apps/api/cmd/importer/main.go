package main

import (
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
	ctx := context.Background()

	// Connect to Database
	pool, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/filmophilia?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

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

	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}

		title := line[0]
		year, _ := strconv.Atoi(line[1])

		fmt.Printf("Processing %s (%d)...\n", title, year)

		tmdb, credits, omdb, err := importer.FetchFullMovieData(title, year)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		// Slug generation
		slug := strings.ToLower(strings.ReplaceAll(title, " ", "-")) + "-" + strconv.Itoa(year)

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

		// Create Movie
		movieID, err := queries.CreateMovie(ctx, db.CreateMovieParams{
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
			fmt.Printf("Failed to create movie: %v\n", err)
			continue
		}

		// Handle Crew (Director)
		for _, person := range credits.Crew {
			if person.Job == "Director" {
				pID, err := queries.UpsertPerson(ctx, db.UpsertPersonParams{
					Name: person.Name,
					Slug: strings.ToLower(strings.ReplaceAll(person.Name, " ", "-")),
				})
				if err != nil {
					fmt.Printf("Failed to upsert person: %v\n", err)
					continue
				}

				err = queries.CreateCredit(ctx, db.CreateCreditParams{
					MovieID:    movieID,
					PersonID:   pID,
					Department: db.DepartmentDIRECTING,
					Role:       "Director",
					Character:  pgtype.Text{},
				})
				if err != nil {
					fmt.Printf("Failed to create credit: %v\n", err)
				}
			}
		}

		fmt.Printf("Successfully imported: %s\n", title)
	}
}
