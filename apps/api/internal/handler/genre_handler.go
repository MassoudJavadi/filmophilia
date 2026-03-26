package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type GenreHandler struct {
	genreSvc service.GenreService
}

func NewGenreHandler(genreSvc service.GenreService) *GenreHandler {
	return &GenreHandler{genreSvc: genreSvc}
}

// GetAll godoc
// @Summary Get all genres
// @Description Get a list of all movie genres
// @Tags genres
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=[]dto.GenreResponse} "List of genres"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres [get]
func (h *GenreHandler) GetAll(c *gin.Context) {
	genres, err := h.genreSvc.GetAll(c.Request.Context())
	if err != nil {
		log.Printf("get all genres error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToGenreResponses(genres))
}

// GetByID godoc
// @Summary Get genre by ID
// @Description Get a specific genre by its ID
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.GenreResponse} "Genre details"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid genre ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Genre not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres/{genreId} [get]
func (h *GenreHandler) GetByID(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		return
	}

	genre, err := h.genreSvc.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre not found")
			return
		}
		log.Printf("get genre by id error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToGenreResponse(genre))
}

// GetBySlug godoc
// @Summary Get genre by slug
// @Description Get a specific genre by its slug
// @Tags genres
// @Produce json
// @Param slug path string true "Genre slug"
// @Success 200 {object} dto.SuccessResponse{data=dto.GenreResponse} "Genre details"
// @Failure 400 {object} dto.BadRequestErrorResponse "Slug required"
// @Failure 404 {object} dto.NotFoundErrorResponse "Genre not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres/slug/{slug} [get]
func (h *GenreHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "GENRE_SLUG_REQUIRED", "slug is required")
		return
	}

	genre, err := h.genreSvc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre not found")
			return
		}
		log.Printf("get genre by slug error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToGenreResponse(genre))
}

// GetMoviesByGenre godoc
// @Summary Get movies by genre
// @Description Get all movies in a specific genre
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.MovieResponse} "List of movies"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid genre ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres/{genreId}/movies [get]
func (h *GenreHandler) GetMoviesByGenre(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.genreSvc.GetMoviesByGenre(c.Request.Context(), int32(id), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movies by genre error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToMovieResponses(movies))
}

// GetMoviesByGenreSlug godoc
// @Summary Get movies by genre slug
// @Description Get all movies in a specific genre using slug
// @Tags genres
// @Produce json
// @Param slug path string true "Genre slug"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.MovieResponse} "List of movies"
// @Failure 400 {object} dto.BadRequestErrorResponse "Slug required"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres/slug/{slug}/movies [get]
func (h *GenreHandler) GetMoviesByGenreSlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "GENRE_SLUG_REQUIRED", "slug is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.genreSvc.GetMoviesByGenreSlug(c.Request.Context(), slug, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movies by genre slug error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToMovieResponses(movies))
}

// GetGenreWithCount godoc
// @Summary Get genre with movie count
// @Description Get genre details including the number of movies in it
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.GenreWithCountResponse} "Genre with movie count"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid genre ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Genre not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /genres/{genreId}/stats [get]
func (h *GenreHandler) GetGenreWithCount(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		return
	}

	genre, err := h.genreSvc.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre not found")
			return
		}
		log.Printf("get genre error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	count, err := h.genreSvc.CountMoviesByGenre(c.Request.Context(), int32(id))
	if err != nil {
		log.Printf("count movies error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.GenreWithCountResponse{
		ID:         genre.ID,
		Name:       genre.Name,
		Slug:       genre.Slug,
		MovieCount: count,
	})
}
