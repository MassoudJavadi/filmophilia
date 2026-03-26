package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type RatingHandler struct {
	ratingSvc service.RatingService
}

func NewRatingHandler(ratingSvc service.RatingService) *RatingHandler {
	return &RatingHandler{ratingSvc: ratingSvc}
}

// RateMovie godoc
// @Summary Rate a movie
// @Description Add or update your rating for a movie (1-10)
// @Tags ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Param request body dto.RateMovieRequest true "Rating score"
// @Success 200 {object} map[string]dto.RatingResponse "Rating created/updated"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Movie not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/rating [put]
func (h *RatingHandler) RateMovie(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var req dto.RateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	rating, err := h.ratingSvc.RateMovie(c.Request.Context(), userID, int32(movieID), req.Score)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		default:
			log.Printf("rate movie error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToRatingResponse(rating)})
}

// GetMyRating godoc
// @Summary Get my rating for a movie
// @Description Get the authenticated user's rating for a specific movie
// @Tags ratings
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} map[string]dto.RatingResponse "User's rating"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 404 {object} map[string]string "Rating not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/rating [get]
func (h *RatingHandler) GetMyRating(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	rating, err := h.ratingSvc.GetUserRating(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrRatingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "you haven't rated this movie"})
			return
		}
		log.Printf("get my rating error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToRatingResponse(rating)})
}

// GetMovieRatings godoc
// @Summary Get movie ratings
// @Description Get all ratings for a specific movie with pagination
// @Tags ratings
// @Produce json
// @Param movieId path int true "Movie ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} map[string][]dto.RatingWithUserResponse "List of ratings"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/ratings [get]
func (h *RatingHandler) GetMovieRatings(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	ratings, err := h.ratingSvc.GetMovieRatings(c.Request.Context(), int32(movieID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movie ratings error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToRatingWithUserResponses(ratings)})
}

// GetMyRatings godoc
// @Summary Get my ratings
// @Description Get all ratings made by the authenticated user
// @Tags ratings
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} map[string][]dto.RatingWithMovieResponse "List of ratings"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/ratings [get]
func (h *RatingHandler) GetMyRatings(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	ratings, err := h.ratingSvc.GetUserRatings(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my ratings error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToRatingWithMovieResponses(ratings)})
}

// DeleteRating godoc
// @Summary Delete my rating
// @Description Remove the authenticated user's rating for a movie
// @Tags ratings
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} map[string]string "Rating deleted"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 404 {object} map[string]string "Rating not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/rating [delete]
func (h *RatingHandler) DeleteRating(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.ratingSvc.DeleteRating(c.Request.Context(), userID, int32(movieID)); err != nil {
		if errors.Is(err, service.ErrRatingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "rating not found"})
			return
		}
		log.Printf("delete rating error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rating deleted"})
}
