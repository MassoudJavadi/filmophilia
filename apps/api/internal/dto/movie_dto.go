package dto

type AdvancedSearchRequest struct {
	Query         string   `form:"q"`
	GenreIDs      []int32  `form:"genre_ids"`
	YearFrom      *int32   `form:"year_from"`
	YearTo        *int32   `form:"year_to"`
	ImdbMin       *float64 `form:"imdb_min"`
	ImdbMax       *float64 `form:"imdb_max"`
	UserRatingMin *float32 `form:"user_rating_min"`
	UserRatingMax *float32 `form:"user_rating_max"`
	RTMin         *int32   `form:"rt_min"`
	RTMax         *int32   `form:"rt_max"`
	MetacriticMin *int32   `form:"metacritic_min"`
	MetacriticMax *int32   `form:"metacritic_max"`
	SortBy        string   `form:"sort_by"`
	Limit         int32    `form:"limit"`
	Page          int32    `form:"page"`
}

type AdvancedSearchResponse struct {
	Movies     []MovieResponse `json:"movies"`
	Total      int32           `json:"total"`
	Page       int32           `json:"page"`
	Limit      int32           `json:"limit"`
	TotalPages int32           `json:"total_pages"`
}

type MovieCardResponse struct {
	ID              int32    `json:"id"`
	Title           string   `json:"title"`
	Slug            string   `json:"slug"`
	PosterURL       string   `json:"poster_url,omitempty"`
	Directors       []string `json:"directors"`
	UserAvgRating   *float32 `json:"user_avg_rating"`
	ImdbRating      *float64 `json:"imdb_rating"`
	RottenTomatoes  *int32   `json:"rotten_tomatoes"`
	MetacriticScore *int32   `json:"metacritic_score"`
	ReleaseDate     string   `json:"release_date,omitempty"`
}

type MovieResponse struct {
	ID               int32    `json:"id"`
	Title            string   `json:"title"`
	Slug             string   `json:"slug"`
	Overview         string   `json:"overview,omitempty"`
	PosterURL        string   `json:"poster_url,omitempty"`
	BackdropURL      string   `json:"backdrop_url,omitempty"`
	TrailerURL       string   `json:"trailer_url,omitempty"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Runtime          int32    `json:"runtime,omitempty"`
	ContentRating    string   `json:"content_rating,omitempty"`
	OriginalLanguage string   `json:"original_language,omitempty"`
	Country          string   `json:"country,omitempty"`
	ImdbID           string   `json:"imdb_id,omitempty"`
	TmdbID           int32    `json:"tmdb_id,omitempty"`
	UserAvgRating    *float32 `json:"user_avg_rating"`
	UserRatingCount  int32    `json:"user_rating_count,omitempty"`
	ImdbRating       *float64 `json:"imdb_rating"`
	RottenTomatoes   *int32   `json:"rotten_tomatoes"`
	MetacriticScore  *int32   `json:"metacritic_score"`
	LetterboxdRating *float64 `json:"letterboxd_rating"`
	Genres           []string `json:"genres,omitempty"`
}

// Valid sort options
var ValidSortOptions = map[string]bool{
	"title_asc":            true,
	"title_desc":           true,
	"release_date_asc":     true,
	"release_date_desc":    true,
	"imdb_rating_asc":      true,
	"imdb_rating_desc":     true,
	"user_rating_asc":      true,
	"user_rating_desc":     true,
	"rotten_tomatoes_asc":  true,
	"rotten_tomatoes_desc": true,
	"metacritic_asc":       true,
	"metacritic_desc":      true,
}
