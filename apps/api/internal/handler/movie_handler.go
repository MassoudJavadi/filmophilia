package handler

import (
	"net/http"
	"strconv"

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
