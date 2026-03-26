package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
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
// @Success 200 {object} object "List of genres"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres [get]
func (h *GenreHandler) GetAll(c *gin.Context) {
	genres, err := h.genreSvc.GetAll(c.Request.Context())
	if err != nil {
		log.Printf("get all genres error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToGenreResponses(genres)})
}

// GetByID godoc
// @Summary Get genre by ID
// @Description Get a specific genre by its ID
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Success 200 {object} object "Genre details"
// @Failure 400 {object} map[string]string "Invalid genre ID"
// @Failure 404 {object} map[string]string "Genre not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres/{genreId} [get]
func (h *GenreHandler) GetByID(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre id"})
		return
	}

	genre, err := h.genreSvc.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		log.Printf("get genre by id error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToGenreResponse(genre)})
}

// GetBySlug godoc
// @Summary Get genre by slug
// @Description Get a specific genre by its slug
// @Tags genres
// @Produce json
// @Param slug path string true "Genre slug"
// @Success 200 {object} object "Genre details"
// @Failure 400 {object} map[string]string "Slug required"
// @Failure 404 {object} map[string]string "Genre not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres/slug/{slug} [get]
func (h *GenreHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	genre, err := h.genreSvc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		log.Printf("get genre by slug error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToGenreResponse(genre)})
}

// GetMoviesByGenre godoc
// @Summary Get movies by genre
// @Description Get all movies in a specific genre
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of movies"
// @Failure 400 {object} map[string]string "Invalid genre ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres/{genreId}/movies [get]
func (h *GenreHandler) GetMoviesByGenre(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.genreSvc.GetMoviesByGenre(c.Request.Context(), int32(id), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movies by genre error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}

// GetMoviesByGenreSlug godoc
// @Summary Get movies by genre slug
// @Description Get all movies in a specific genre using slug
// @Tags genres
// @Produce json
// @Param slug path string true "Genre slug"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of movies"
// @Failure 400 {object} map[string]string "Slug required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres/slug/{slug}/movies [get]
func (h *GenreHandler) GetMoviesByGenreSlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	movies, err := h.genreSvc.GetMoviesByGenreSlug(c.Request.Context(), slug, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movies by genre slug error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movies})
}

// GetGenreWithCount godoc
// @Summary Get genre with movie count
// @Description Get genre details including the number of movies in it
// @Tags genres
// @Produce json
// @Param genreId path int true "Genre ID"
// @Success 200 {object} map[string]interface{} "Genre with movie count"
// @Failure 400 {object} map[string]string "Invalid genre ID"
// @Failure 404 {object} map[string]string "Genre not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /genres/{genreId}/stats [get]
func (h *GenreHandler) GetGenreWithCount(c *gin.Context) {
	idStr := c.Param("genreId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre id"})
		return
	}

	genre, err := h.genreSvc.GetByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrGenreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
			return
		}
		log.Printf("get genre error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	count, err := h.genreSvc.CountMoviesByGenre(c.Request.Context(), int32(id))
	if err != nil {
		log.Printf("count movies error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":          genre.ID,
			"name":        genre.Name,
			"slug":        genre.Slug,
			"movie_count": count,
		},
	})
}
