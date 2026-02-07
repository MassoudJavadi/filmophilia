package importer

// TMDBMovie represents movie data from TMDB API
type TMDBMovie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	ReleaseDate string `json:"release_date"`
	PosterPath  string `json:"poster_path"`
	IMDBID      string `json:"imdb_id"`
	Runtime     int    `json:"runtime"`
}

// TMDBPerson represents a person in credits
type TMDBPerson struct {
	Name       string `json:"name"`
	Job        string `json:"job,omitempty"`
	Department string `json:"department,omitempty"`
	Character  string `json:"character,omitempty"`
}

// TMDBCredits represents movie credits from TMDB
type TMDBCredits struct {
	Cast []TMDBPerson `json:"cast"`
	Crew []TMDBPerson `json:"crew"`
}

// OMDBRating represents a rating from a specific source
type OMDBRating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

// OMDBResponse represents movie data from OMDB API
type OMDBResponse struct {
	Title      string       `json:"Title"`
	Year       string       `json:"Year"`
	ImdbRating string       `json:"imdbRating"`
	Metascore  string       `json:"Metascore"`
	Ratings    []OMDBRating `json:"Ratings"`
}
