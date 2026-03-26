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

type WatchlistHandler struct {
	watchlistSvc service.WatchlistService
}

func NewWatchlistHandler(watchlistSvc service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistSvc: watchlistSvc}
}

// AddToWatchlist godoc
// @Summary Add movie to watchlist
// @Description Add a movie to the authenticated user's watchlist
// @Tags watchlist
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Param request body dto.AddToWatchlistRequest false "Optional notes and rank"
// @Success 201 {object} dto.SuccessResponse{data=dto.WatchlistEntryResponse} "Movie added to watchlist"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not found"
// @Failure 409 {object} dto.ConflictErrorResponse "Movie already in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist [post]
func (h *WatchlistHandler) AddToWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	var req dto.AddToWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Add(c.Request.Context(), userID, int32(movieID), req.Notes, req.Rank)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			response.Error(c, http.StatusNotFound, "MOVIE_NOT_FOUND", "movie not found")
		case errors.Is(err, service.ErrAlreadyInWatchlist):
			response.Error(c, http.StatusConflict, "ALREADY_IN_WATCHLIST", "movie already in watchlist")
		default:
			log.Printf("add to watchlist error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Created(c, mapper.ToWatchlistEntryResponse(entry))
}

// GetWatchlist godoc
// @Summary Get my watchlist
// @Description Get the authenticated user's watchlist with optional filtering
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param filter query string false "Filter: all, watched, unwatched" default(all)
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.WatchlistItemResponse} "Watchlist items"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/watchlist [get]
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
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistItemResponses(items))
}

// GetWatchlistCount godoc
// @Summary Get watchlist count
// @Description Get count of movies in the authenticated user's watchlist
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistCountResponse} "Watchlist count"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/watchlist/count [get]
func (h *WatchlistHandler) GetWatchlistCount(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	count, err := h.watchlistSvc.Count(c.Request.Context(), userID)
	if err != nil {
		log.Printf("get watchlist count error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistCountResponse(count))
}

// GetWatchlistEntry godoc
// @Summary Get watchlist entry
// @Description Get a specific movie's watchlist entry for the authenticated user
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistEntryResponse} "Watchlist entry"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist [get]
func (h *WatchlistHandler) GetWatchlistEntry(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Get(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			response.Error(c, http.StatusNotFound, "NOT_IN_WATCHLIST", "movie not in watchlist")
			return
		}
		log.Printf("get watchlist entry error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistEntryResponse(entry))
}

// UpdateWatchlistEntry godoc
// @Summary Update watchlist entry
// @Description Update notes or rank for a movie in the watchlist
// @Tags watchlist
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Param request body dto.UpdateWatchlistRequest true "Update data"
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistEntryResponse} "Updated entry"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist [patch]
func (h *WatchlistHandler) UpdateWatchlistEntry(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	var req dto.UpdateWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.Update(c.Request.Context(), userID, int32(movieID), req.Notes, req.Rank)
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			response.Error(c, http.StatusNotFound, "NOT_IN_WATCHLIST", "movie not in watchlist")
			return
		}
		log.Printf("update watchlist entry error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistEntryResponse(entry))
}

// MarkAsWatched godoc
// @Summary Mark movie as watched
// @Description Mark a movie in the watchlist as watched
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistEntryResponse} "Updated entry"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist/watched [post]
func (h *WatchlistHandler) MarkAsWatched(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.MarkWatched(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			response.Error(c, http.StatusNotFound, "NOT_IN_WATCHLIST", "movie not in watchlist")
			return
		}
		log.Printf("mark as watched error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistEntryResponse(entry))
}

// MarkAsUnwatched godoc
// @Summary Mark movie as unwatched
// @Description Mark a movie in the watchlist as not watched
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistEntryResponse} "Updated entry"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist/watched [delete]
func (h *WatchlistHandler) MarkAsUnwatched(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	entry, err := h.watchlistSvc.MarkUnwatched(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			response.Error(c, http.StatusNotFound, "NOT_IN_WATCHLIST", "movie not in watchlist")
			return
		}
		log.Printf("mark as unwatched error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToWatchlistEntryResponse(entry))
}

// RemoveFromWatchlist godoc
// @Summary Remove from watchlist
// @Description Remove a movie from the authenticated user's watchlist
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Removed successfully"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not in watchlist"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist [delete]
func (h *WatchlistHandler) RemoveFromWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.watchlistSvc.Remove(c.Request.Context(), userID, int32(movieID)); err != nil {
		if errors.Is(err, service.ErrNotInWatchlist) {
			response.Error(c, http.StatusNotFound, "NOT_IN_WATCHLIST", "movie not in watchlist")
			return
		}
		log.Printf("remove from watchlist error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.Message(c, http.StatusOK, "removed from watchlist")
}

// CheckInWatchlist godoc
// @Summary Check if movie is in watchlist
// @Description Check if a movie is in the authenticated user's watchlist
// @Tags watchlist
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WatchlistStatusData} "In watchlist status"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/watchlist/check [get]
func (h *WatchlistHandler) CheckInWatchlist(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	inWatchlist, err := h.watchlistSvc.IsInWatchlist(c.Request.Context(), userID, int32(movieID))
	if err != nil {
		log.Printf("check in watchlist error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.WatchlistStatusData{InWatchlist: inWatchlist})
}
