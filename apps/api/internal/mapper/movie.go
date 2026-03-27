package mapper

import (
	"math/big"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToMovieResponse(m db.Movie) dto.MovieResponse {
	resp := dto.MovieResponse{
		ID:    m.ID,
		Title: m.Title,
		Slug:  m.Slug,
	}

	if m.Overview.Valid {
		resp.Overview = m.Overview.String
	}
	if m.PosterUrl.Valid {
		resp.PosterURL = m.PosterUrl.String
	}
	if m.BackdropUrl.Valid {
		resp.BackdropURL = m.BackdropUrl.String
	}
	if m.TrailerUrl.Valid {
		resp.TrailerURL = m.TrailerUrl.String
	}
	if m.ReleaseDate.Valid {
		resp.ReleaseDate = m.ReleaseDate.Time.Format("2006-01-02")
	}
	if m.Runtime.Valid {
		resp.Runtime = m.Runtime.Int32
	}
	if m.ContentRating.Valid {
		resp.ContentRating = m.ContentRating.String
	}
	if m.OriginalLanguage.Valid {
		resp.OriginalLanguage = m.OriginalLanguage.String
	}
	if m.Country.Valid {
		resp.Country = m.Country.String
	}
	if m.ImdbID.Valid {
		resp.ImdbID = m.ImdbID.String
	}
	if m.TmdbID.Valid {
		resp.TmdbID = m.TmdbID.Int32
	}
	if m.UserAvgRating.Valid {
		resp.UserAvgRating = m.UserAvgRating.Float32
	}
	if m.UserRatingCount.Valid {
		resp.UserRatingCount = m.UserRatingCount.Int32
	}
	if m.ImdbRating.Valid {
		resp.ImdbRating = numericToFloat64(m.ImdbRating)
	}
	if m.RottenTomatoes.Valid {
		resp.RottenTomatoes = m.RottenTomatoes.Int32
	}
	if m.MetacriticScore.Valid {
		resp.MetacriticScore = m.MetacriticScore.Int32
	}
	if m.LetterboxdRating.Valid {
		resp.LetterboxdRating = numericToFloat64(m.LetterboxdRating)
	}

	return resp
}

func ToMovieCardResponse(m db.ListMoviesRow) dto.MovieCardResponse {
	resp := dto.MovieCardResponse{
		ID:        m.ID,
		Title:     m.Title,
		Slug:      m.Slug,
		Directors: m.Directors,
	}

	if m.PosterUrl.Valid {
		resp.PosterURL = m.PosterUrl.String
	}
	if m.ReleaseDate.Valid {
		resp.ReleaseDate = m.ReleaseDate.Time.Format("2006-01-02")
	}
	if m.UserAvgRating.Valid {
		resp.UserAvgRating = m.UserAvgRating.Float32
	}
	if m.ImdbRating.Valid {
		resp.ImdbRating = numericToFloat64(m.ImdbRating)
	}
	if m.RottenTomatoes.Valid {
		resp.RottenTomatoes = m.RottenTomatoes.Int32
	}
	if m.MetacriticScore.Valid {
		resp.MetacriticScore = m.MetacriticScore.Int32
	}

	return resp
}

func ToMovieCardResponses(movies []db.ListMoviesRow) []dto.MovieCardResponse {
	result := make([]dto.MovieCardResponse, len(movies))
	for i, m := range movies {
		result[i] = ToMovieCardResponse(m)
	}
	return result
}

func ToMovieResponses(movies []db.Movie) []dto.MovieResponse {
	result := make([]dto.MovieResponse, len(movies))
	for i, m := range movies {
		result[i] = ToMovieResponse(m)
	}
	return result
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	f := new(big.Float).SetInt(n.Int)
	exp := new(big.Float).SetFloat64(1)
	for i := int32(0); i < -n.Exp; i++ {
		exp.Mul(exp, big.NewFloat(10))
	}
	f.Quo(f, exp)
	result, _ := f.Float64()
	return result
}
