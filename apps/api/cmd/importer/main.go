package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/importer"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load .env file if it exists (try multiple paths)
	loadEnvFile(".env")
	loadEnvFile("../../.env")
	loadEnvFile("apps/api/.env")

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

	// Load user average ratings from Beta - Film.csv
	userRatings := loadUserRatings(resolveInputPath("Beta - Film.csv"))
	fmt.Printf("Loaded %d user avg ratings from 'Beta - Film.csv'\n", len(userRatings))

	// Open movies.csv
	f, err := os.Open(resolveInputPath("movies.csv"))
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
	skipped := 0
	skipUnmatched := true

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

		// Parse user avg rating from Beta - Film.csv
		var userAvgRating pgtype.Float4
		if avg, ok := userRatings[normalizeTitle(title)]; ok {
			userAvgRating = pgtype.Float4{
				Float32: avg,
				Valid:   true,
			}
		}

		if movieID, found, err := findMovieIDByTitleAndYear(ctx, pool, title, year); err != nil {
			fmt.Printf("FAILED to match existing movie: %v\n", err)
			failed++
			continue
		} else if found {
			if err := updateMovieUserAvgRating(ctx, pool, movieID, userAvgRating); err != nil {
				fmt.Printf("FAILED to update user_avg_rating: %v\n", err)
				failed++
				continue
			}
			fmt.Printf("OK (updated existing movie)\n")
			success++
			continue
		} else if skipUnmatched {
			fmt.Printf("SKIPPED (unmatched in DB)\n")
			skipped++
			continue
		}

		// Fetch extra movie data with retry for TMDB/OMDB rate limit
		tmdb, credits, omdb, err := fetchWithRetry(title, year, 3)
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

		if userAvgRating.Valid {
			if err := updateMovieUserAvgRating(ctx, pool, movieID, userAvgRating); err != nil {
				fmt.Printf("FAILED to update user_avg_rating: %v\n", err)
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

		time.Sleep(200 * time.Millisecond) // slight delay to avoid hammering APIs
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Import completed!\n")
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total: %d\n", total)
}

// Advanced TMDB/OMDB fetcher with retry logic
func fetchWithRetry(title string, year int, maxTry int) (tmdb *importer.TMDBMovie, credits *importer.TMDBCredits, omdb *importer.OMDBResponse, err error) {
	for attempt := 0; attempt < maxTry; attempt++ {
		tmdb, credits, omdb, err = importer.FetchFullMovieData(title, year)
		if err == nil {
			return
		}
		dur := time.Duration(1000+attempt*500) * time.Millisecond
		time.Sleep(dur)
	}
	return
}

func updateMovieUserAvgRating(ctx context.Context, pool *pgxpool.Pool, movieID int32, userAvgRating pgtype.Float4) error {
	if !userAvgRating.Valid {
		return nil
	}

	_, err := pool.Exec(ctx, "UPDATE movies SET user_avg_rating = $1 WHERE id = $2", userAvgRating, movieID)
	return err
}

func findMovieIDByTitleAndYear(ctx context.Context, pool *pgxpool.Pool, title string, year int) (int32, bool, error) {
	var movieID int32
	err := pool.QueryRow(ctx, `
		SELECT id
		FROM movies
		WHERE (
			lower(title) = lower($3)
			OR regexp_replace(
				translate(lower(title), 'åáàâäãāéèêëēíìîïīóòôöõōúùûüūñçø', 'aaaaaaaeeeeeiiiiioooooouuuuunco'),
				'[^a-z0-9]+',
				'',
				'g'
			) = $1
		)
		  AND ($2 = 0 OR EXTRACT(YEAR FROM release_date) = $2)
		LIMIT 1
	`, normalizeLookupKey(title), year, title).Scan(&movieID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	return movieID, true, nil
}

// Helper: load AVG column from Beta - Film.csv into map[normalizedTitle]float32
func loadUserRatings(path string) map[string]float32 {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[WARN] Could not open ratings file: %v", err)
		return map[string]float32{}
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Printf("[WARN] Failed reading ratings csv: %v", err)
		return map[string]float32{}
	}

	ratings := make(map[string]float32)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			continue
		}
		title := normalizeTitle(row[0])
		avgStr := strings.TrimSpace(row[3])
		if avgStr == "" {
			continue
		}
		val, err := strconv.ParseFloat(avgStr, 32)
		if err != nil {
			continue
		}
		ratings[title] = float32(val)
	}
	return ratings
}

func resolveInputPath(name string) string {
	candidates := []string{
		name,
		filepath.Join("cmd", "importer", name),
		filepath.Join("apps", "api", "cmd", "importer", name),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return name
}

func normalizeTitle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "!", "")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "  ", " ")
	return s
}

func normalizeLookupKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	replacer := strings.NewReplacer(
		"&", "and",
		"æ", "ae",
		"œ", "oe",
		"ß", "ss",
		"ø", "o",
		"å", "a",
		"á", "a",
		"à", "a",
		"â", "a",
		"ä", "a",
		"ã", "a",
		"ā", "a",
		"é", "e",
		"è", "e",
		"ê", "e",
		"ë", "e",
		"ē", "e",
		"í", "i",
		"ì", "i",
		"î", "i",
		"ï", "i",
		"ī", "i",
		"ó", "o",
		"ò", "o",
		"ô", "o",
		"ö", "o",
		"õ", "o",
		"ō", "o",
		"ú", "u",
		"ù", "u",
		"û", "u",
		"ü", "u",
		"ū", "u",
		"ñ", "n",
		"ç", "c",
	)
	s = replacer.Replace(s)

	var builder strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// generateSlug creates a URL-friendly slug from a title
func generateSlug(title string, year int) string {
	slug := strings.ToLower(title)
	replacer := strings.NewReplacer(
		" ", "-",
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
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if year > 0 {
		slug = fmt.Sprintf("%s-%d", slug, year)
	}
	return slug
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
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
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}
