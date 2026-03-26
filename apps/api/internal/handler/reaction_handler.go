package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/MassoudJavadi/filmophilia/api/internal/service"
	"github.com/gin-gonic/gin"
)

type ReactionHandler struct {
	reactionSvc service.ReactionService
}

func NewReactionHandler(reactionSvc service.ReactionService) *ReactionHandler {
	return &ReactionHandler{reactionSvc: reactionSvc}
}

// ReactToComment godoc
// @Summary React to a comment
// @Description Add or update a reaction (like, love, etc.) to a comment
// @Tags reactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Param request body dto.AddReactionRequest true "Reaction type"
// @Success 200 {object} dto.SuccessResponse{data=dto.ReactionMutationData} "Reaction created or updated"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 404 {object} dto.NotFoundErrorResponse "Comment not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId}/reactions [post]
func (h *ReactionHandler) ReactToComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	var req dto.AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)

	reaction, err := h.reactionSvc.ReactToComment(c.Request.Context(), userID, commentID, db.ReactionType(req.Type))
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "COMMENT_NOT_FOUND", "comment not found")
			return
		}
		log.Printf("react to comment error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.ReactionMutationData{
		ID:        reaction.ID,
		Type:      string(reaction.Type),
		CreatedAt: reaction.CreatedAt.Time,
	})
}

// RemoveCommentReaction godoc
// @Summary Remove reaction from comment
// @Description Remove your reaction from a comment
// @Tags reactions
// @Produce json
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Reaction removed"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Reaction not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId}/reactions [delete]
func (h *ReactionHandler) RemoveCommentReaction(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.reactionSvc.RemoveCommentReaction(c.Request.Context(), userID, commentID); err != nil {
		if errors.Is(err, service.ErrReactionNotFound) {
			response.Error(c, http.StatusNotFound, "REACTION_NOT_FOUND", "reaction not found")
			return
		}
		log.Printf("remove comment reaction error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.Message(c, http.StatusOK, "reaction removed")
}

// GetCommentReactions godoc
// @Summary Get comment reactions
// @Description Get all reactions on a comment
// @Tags reactions
// @Produce json
// @Param commentId path int true "Comment ID"
// @Param limit query int false "Results per page" default(50)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.ReactionResponse} "List of reactions"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId}/reactions [get]
func (h *ReactionHandler) GetCommentReactions(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	reactions, err := h.reactionSvc.GetCommentReactions(c.Request.Context(), commentID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get comment reactions error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToReactionResponses(reactions))
}

// GetCommentReactionSummary godoc
// @Summary Get reaction summary
// @Description Get reaction counts by type for a comment, plus the current user's reaction if authenticated
// @Tags reactions
// @Produce json
// @Param commentId path int true "Comment ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ReactionSummaryResponse} "Reaction summary"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId}/reactions/summary [get]
func (h *ReactionHandler) GetCommentReactionSummary(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	counts, err := h.reactionSvc.GetCommentReactionCounts(c.Request.Context(), commentID)
	if err != nil {
		log.Printf("get comment reaction counts error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	summary := dto.ReactionSummaryResponse{
		Counts: mapper.ToReactionCounts(counts),
	}

	// If authenticated, include user's reaction
	if userID, exists := c.Get("user_id"); exists {
		userReaction, err := h.reactionSvc.GetUserCommentReaction(c.Request.Context(), userID.(int32), commentID)
		if err != nil {
			log.Printf("get user comment reaction error: %v", err)
		} else if userReaction != nil {
			reactionType := string(userReaction.Type)
			summary.UserReaction = &reactionType
		}
	}

	response.OK(c, summary)
}
