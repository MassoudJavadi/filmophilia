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
