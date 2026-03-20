package dto

type GenreResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type GenreWithCountResponse struct {
	ID         int32  `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	MovieCount int32  `json:"movie_count,omitempty"`
}
