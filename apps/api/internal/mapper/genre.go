package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToGenreResponse(g db.Genre) dto.GenreResponse {
	return dto.GenreResponse{
		ID:   g.ID,
		Name: g.Name,
		Slug: g.Slug,
	}
}

func ToGenreResponses(rows []db.Genre) []dto.GenreResponse {
	result := make([]dto.GenreResponse, len(rows))
	for i, r := range rows {
		result[i] = ToGenreResponse(r)
	}
	return result
}
