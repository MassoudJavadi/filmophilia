package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	movieSvc service.MovieService
}

func NewMovieHandler(movieSvc service.MovieService) *MovieHandler {
	return &MovieHandler{movieSvc: movieSvc}
}

func (h *MovieHandler) GetMovies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.movieSvc.GetMovies(c.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}

func (h *MovieHandler) GetMovie(c *gin.Context) {
	slug := c.Param("slug")
	movie, err := h.movieSvc.GetMovie(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

func (h *MovieHandler) AdvancedSearch(c *gin.Context) {
	var req dto.AdvancedSearchRequest

	// Parse query params
	req.Query = c.Query("q")
	req.SortBy = c.Query("sort_by")

	// Parse genre_ids (comma-separated or repeated params)
	if genreIDsStr := c.Query("genre_ids"); genreIDsStr != "" {
		for _, idStr := range strings.Split(genreIDsStr, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				req.GenreIDs = append(req.GenreIDs, int32(id))
			}
		}
	}

	// Parse year range
	if yearFrom := c.Query("year_from"); yearFrom != "" {
		if v, err := strconv.Atoi(yearFrom); err == nil {
			val := int32(v)
			req.YearFrom = &val
		}
	}
	if yearTo := c.Query("year_to"); yearTo != "" {
		if v, err := strconv.Atoi(yearTo); err == nil {
			val := int32(v)
			req.YearTo = &val
		}
	}

	// Parse IMDB rating range
	if imdbMin := c.Query("imdb_min"); imdbMin != "" {
		if v, err := strconv.ParseFloat(imdbMin, 64); err == nil {
			req.ImdbMin = &v
		}
	}
	if imdbMax := c.Query("imdb_max"); imdbMax != "" {
		if v, err := strconv.ParseFloat(imdbMax, 64); err == nil {
			req.ImdbMax = &v
		}
	}

	// Parse user rating range
	if userMin := c.Query("user_rating_min"); userMin != "" {
		if v, err := strconv.ParseFloat(userMin, 32); err == nil {
			val := float32(v)
			req.UserRatingMin = &val
		}
	}
	if userMax := c.Query("user_rating_max"); userMax != "" {
		if v, err := strconv.ParseFloat(userMax, 32); err == nil {
			val := float32(v)
			req.UserRatingMax = &val
		}
	}

	// Parse Rotten Tomatoes range
	if rtMin := c.Query("rt_min"); rtMin != "" {
		if v, err := strconv.Atoi(rtMin); err == nil {
			val := int32(v)
			req.RTMin = &val
		}
	}
	if rtMax := c.Query("rt_max"); rtMax != "" {
		if v, err := strconv.Atoi(rtMax); err == nil {
			val := int32(v)
			req.RTMax = &val
		}
	}

	// Parse Metacritic range
	if mcMin := c.Query("metacritic_min"); mcMin != "" {
		if v, err := strconv.Atoi(mcMin); err == nil {
			val := int32(v)
			req.MetacriticMin = &val
		}
	}
	if mcMax := c.Query("metacritic_max"); mcMax != "" {
		if v, err := strconv.Atoi(mcMax); err == nil {
			val := int32(v)
			req.MetacriticMax = &val
		}
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			req.Limit = int32(v)
		}
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	if page := c.Query("page"); page != "" {
		if v, err := strconv.Atoi(page); err == nil {
			req.Page = int32(v)
		}
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// Execute search
	movies, total, err := h.movieSvc.AdvancedSearch(c.Request.Context(), req)
	if err != nil {
		log.Printf("advanced search error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	// Calculate total pages
	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dto.AdvancedSearchResponse{
			Movies:     mapper.ToMovieResponses(movies),
			Total:      total,
			Page:       req.Page,
			Limit:      req.Limit,
			TotalPages: totalPages,
		},
	})
}
