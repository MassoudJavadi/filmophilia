package importer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

var (
	tmdbKey = os.Getenv("TMDB_API_KEY")
	omdbKey = os.Getenv("OMDB_API_KEY")
)

// FetchFullMovieData coordinates API calls to get complete movie data
func FetchFullMovieData(title string, year int) (*TMDBMovie, *TMDBCredits, *OMDBResponse, error) {
	// 1. Search TMDB to get movie ID
	searchURL := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s&year=%d",
		tmdbKey, url.QueryEscape(title), year)

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("TMDB search failed: %w", err)
	}
	defer resp.Body.Close()

	var searchResult struct {
		Results []TMDBMovie `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode TMDB search: %w", err)
	}

	if len(searchResult.Results) == 0 {
		return nil, nil, nil, fmt.Errorf("movie not found: %s (%d)", title, year)
	}
	movieID := searchResult.Results[0].ID

	// 2. Get Movie Details from TMDB
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movieID, tmdbKey)
	resp, err = http.Get(detailsURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("TMDB details failed: %w", err)
	}
	defer resp.Body.Close()

	var movieDetail TMDBMovie
	if err := json.NewDecoder(resp.Body).Decode(&movieDetail); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode TMDB details: %w", err)
	}

	// 3. Get Credits from TMDB
	creditsURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?api_key=%s", movieID, tmdbKey)
	resp, err = http.Get(creditsURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("TMDB credits failed: %w", err)
	}
	defer resp.Body.Close()

	var credits TMDBCredits
	if err := json.NewDecoder(resp.Body).Decode(&credits); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode TMDB credits: %w", err)
	}

	// 4. Get Ratings from OMDB using IMDb ID
	var omdbData OMDBResponse
	if movieDetail.IMDBID != "" {
		omdbURL := fmt.Sprintf("https://www.omdbapi.com/?i=%s&apikey=%s", movieDetail.IMDBID, omdbKey)
		resp, err = http.Get(omdbURL)
		if err != nil {
			fmt.Printf("Warning: OMDB fetch failed: %v\n", err)
		} else {
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&omdbData)
		}
	}

	return &movieDetail, &credits, &omdbData, nil
}
