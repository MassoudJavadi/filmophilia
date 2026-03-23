package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type ReactionHandler struct {
	reactionSvc service.ReactionService
}

func NewReactionHandler(reactionSvc service.ReactionService) *ReactionHandler {
	return &ReactionHandler{reactionSvc: reactionSvc}
}

func (h *ReactionHandler) ReactToComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req dto.AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	reaction, err := h.reactionSvc.ReactToComment(c.Request.Context(), userID, commentID, db.ReactionType(req.Type))
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		log.Printf("react to comment error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id":         reaction.ID,
		"type":       reaction.Type,
		"created_at": reaction.CreatedAt.Time,
	}})
}

func (h *ReactionHandler) RemoveCommentReaction(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.reactionSvc.RemoveCommentReaction(c.Request.Context(), userID, commentID); err != nil {
		if errors.Is(err, service.ErrReactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reaction not found"})
			return
		}
		log.Printf("remove comment reaction error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

func (h *ReactionHandler) GetCommentReactions(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	reactions, err := h.reactionSvc.GetCommentReactions(c.Request.Context(), commentID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get comment reactions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToReactionResponses(reactions)})
}

func (h *ReactionHandler) GetCommentReactionSummary(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	counts, err := h.reactionSvc.GetCommentReactionCounts(c.Request.Context(), commentID)
	if err != nil {
		log.Printf("get comment reaction counts error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := dto.ReactionSummaryResponse{
		Counts: mapper.ToReactionCounts(counts),
	}

	// If authenticated, include user's reaction
	if userID, exists := c.Get("user_id"); exists {
		userReaction, err := h.reactionSvc.GetUserCommentReaction(c.Request.Context(), userID.(int32), commentID)
		if err != nil {
			log.Printf("get user comment reaction error: %v", err)
		} else if userReaction != nil {
			reactionType := string(userReaction.Type)
			response.UserReaction = &reactionType
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}
