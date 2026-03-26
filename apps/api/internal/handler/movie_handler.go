package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	movieSvc service.MovieService
}

func NewMovieHandler(movieSvc service.MovieService) *MovieHandler {
	return &MovieHandler{movieSvc: movieSvc}
}

// GetMovies godoc
// @Summary List movies
// @Description Get a paginated list of movies
// @Tags movies
// @Produce json
// @Param limit query int false "Number of movies per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.MovieResponse} "List of movies"
// @Failure 500 {object} dto.InternalServerErrorResponse "Failed to fetch movies"
// @Router /movies [get]
func (h *MovieHandler) GetMovies(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.movieSvc.GetMovies(c.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "MOVIES_FETCH_FAILED", "failed to fetch movies")
		return
	}

	response.OK(c, mapper.ToMovieResponses(movies))
}

// GetMovie godoc
// @Summary Get movie by slug
// @Description Get detailed information about a specific movie
// @Tags movies
// @Produce json
// @Param slug path string true "Movie slug"
// @Success 200 {object} dto.SuccessResponse{data=dto.MovieResponse} "Movie details"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not found"
// @Router /movies/slug/{slug} [get]
func (h *MovieHandler) GetMovie(c *gin.Context) {
	slug := c.Param("slug")
	movie, err := h.movieSvc.GetMovie(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, http.StatusNotFound, "MOVIE_NOT_FOUND", "movie not found")
		return
	}

	response.OK(c, mapper.ToMovieResponse(movie))
}

// AdvancedSearch godoc
// @Summary Search movies with filters
// @Description Search movies with various filters including genre, year, ratings, and more
// @Tags movies
// @Produce json
// @Param q query string false "Search query"
// @Param genre_ids query string false "Comma-separated genre IDs"
// @Param year_from query int false "Minimum release year"
// @Param year_to query int false "Maximum release year"
// @Param imdb_min query number false "Minimum IMDB rating"
// @Param imdb_max query number false "Maximum IMDB rating"
// @Param user_rating_min query number false "Minimum user rating"
// @Param user_rating_max query number false "Maximum user rating"
// @Param rt_min query int false "Minimum Rotten Tomatoes score"
// @Param rt_max query int false "Maximum Rotten Tomatoes score"
// @Param metacritic_min query int false "Minimum Metacritic score"
// @Param metacritic_max query int false "Maximum Metacritic score"
// @Param sort_by query string false "Sort field" Enums(title_asc, title_desc, release_date_asc, release_date_desc, imdb_rating_asc, imdb_rating_desc, user_rating_asc, user_rating_desc)
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=dto.AdvancedSearchResponse} "Search results"
// @Failure 500 {object} dto.InternalServerErrorResponse "Search failed"
// @Router /movies/search [get]
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
		response.Error(c, http.StatusInternalServerError, "SEARCH_FAILED", "search failed")
		return
	}

	// Calculate total pages
	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	response.OK(c, dto.AdvancedSearchResponse{
		Movies:     mapper.ToMovieResponses(movies),
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	})
}
