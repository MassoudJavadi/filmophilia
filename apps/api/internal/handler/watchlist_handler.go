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

type WatchlistHandler struct {
	watchlistSvc service.WatchlistService
}

func NewWatchlistHandler(watchlistSvc service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistSvc: watchlistSvc}
}

func (h *WatchlistHandler) AddToWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var req dto.AddToWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Add(c.Request.Context(), userID, int32(movieID), req.Notes, req.Rank)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		case errors.Is(err, service.ErrAlreadyInWatchlist):
			c.JSON(http.StatusConflict, gin.H{"error": "movie already in watchlist"})
		default:
			log.Printf("add to watchlist error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mapper.ToWatchlistEntryResponse(entry)})
}

func (h *WatchlistHandler) GetWatchlist(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	filterStr := c.DefaultQuery("filter", "all")
	filter := service.WatchlistFilter(filterStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	items, err := h.watchlistSvc.List(c.Request.Context(), userID, filter, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get watchlist error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistItemResponses(items)})
}

func (h *WatchlistHandler) GetWatchlistCount(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	count, err := h.watchlistSvc.Count(c.Request.Context(), userID)
	if err != nil {
		log.Printf("get watchlist count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistCountResponse(count)})
}

func (h *WatchlistHandler) GetWatchlistEntry(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Get(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not in watchlist"})
			return
		}
		log.Printf("get watchlist entry error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistEntryResponse(entry)})
}

func (h *WatchlistHandler) UpdateWatchlistEntry(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var req dto.UpdateWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Update(c.Request.Context(), userID, int32(movieID), req.Notes, req.Rank)
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not in watchlist"})
			return
		}
		log.Printf("update watchlist entry error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistEntryResponse(entry)})
}

func (h *WatchlistHandler) MarkAsWatched(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.MarkWatched(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not in watchlist"})
			return
		}
		log.Printf("mark as watched error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistEntryResponse(entry)})
}

func (h *WatchlistHandler) MarkAsUnwatched(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.MarkUnwatched(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not in watchlist"})
			return
		}
		log.Printf("mark as unwatched error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToWatchlistEntryResponse(entry)})
}

func (h *WatchlistHandler) RemoveFromWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.watchlistSvc.Remove(c.Request.Context(), userID, int32(movieID)); err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not in watchlist"})
			return
		}
		log.Printf("remove from watchlist error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed from watchlist"})
}

func (h *WatchlistHandler) CheckInWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	inWatchlist, err := h.watchlistSvc.IsInWatchlist(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		log.Printf("check in watchlist error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"in_watchlist": inWatchlist})
}
