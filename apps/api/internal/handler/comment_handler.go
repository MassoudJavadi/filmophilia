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

type CommentHandler struct {
	commentSvc service.CommentService
}

func NewCommentHandler(commentSvc service.CommentService) *CommentHandler {
	return &CommentHandler{commentSvc: commentSvc}
}

// CreateComment godoc
// @Summary Create a comment
// @Description Add a comment to a movie
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Param request body dto.CreateCommentRequest true "Comment content"
// @Success 201 {object} object "Created comment"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Movie not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	comment, err := h.commentSvc.Create(c.Request.Context(), userID, int32(movieID), nil, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		default:
			log.Printf("create comment error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mapper.ToCommentWithUserResponse(comment)})
}

// CreateReply godoc
// @Summary Reply to a comment
// @Description Create a reply to an existing comment
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movieId path int true "Movie ID"
// @Param commentId path int true "Parent Comment ID"
// @Param request body dto.CreateCommentRequest true "Reply content"
// @Success 201 {object} object "Created reply"
// @Failure 400 {object} map[string]string "Invalid request or cannot reply to reply"
// @Failure 404 {object} map[string]string "Movie or parent comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/comments/{commentId}/replies [post]
func (h *CommentHandler) CreateReply(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)
	parentID := commentID

	comment, err := h.commentSvc.Create(c.Request.Context(), userID, int32(movieID), &parentID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		case errors.Is(err, service.ErrParentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "parent comment not found"})
		case errors.Is(err, service.ErrCannotReplyReply):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reply to a reply"})
		default:
			log.Printf("create reply error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": mapper.ToCommentWithUserResponse(comment)})
}

// GetComment godoc
// @Summary Get a comment
// @Description Get a specific comment by ID
// @Tags comments
// @Produce json
// @Param commentId path int true "Comment ID"
// @Success 200 {object} object "Comment details"
// @Failure 400 {object} map[string]string "Invalid comment ID"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /comments/{commentId} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	comment, err := h.commentSvc.GetByID(c.Request.Context(), commentID)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		log.Printf("get comment error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToCommentWithUserResponse(comment)})
}

// GetMovieComments godoc
// @Summary Get movie comments
// @Description Get all top-level comments for a movie
// @Tags comments
// @Produce json
// @Param movieId path int true "Movie ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of comments"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/comments [get]
func (h *CommentHandler) GetMovieComments(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	comments, err := h.commentSvc.GetMovieComments(c.Request.Context(), int32(movieID), int32(limit), int32(offset))
	if err != nil {
		log.Printf("get movie comments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToCommentResponses(comments)})
}

// GetReplies godoc
// @Summary Get comment replies
// @Description Get all replies to a comment
// @Tags comments
// @Produce json
// @Param commentId path int true "Comment ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of replies"
// @Failure 400 {object} map[string]string "Invalid comment ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /comments/{commentId}/replies [get]
func (h *CommentHandler) GetReplies(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	replies, err := h.commentSvc.GetReplies(c.Request.Context(), commentID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get replies error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToReplyResponses(replies)})
}

// GetMyComments godoc
// @Summary Get my comments
// @Description Get all comments made by the authenticated user
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object "List of comments"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /me/comments [get]
func (h *CommentHandler) GetMyComments(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	comments, err := h.commentSvc.GetUserComments(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		log.Printf("get my comments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToCommentWithMovieResponses(comments)})
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update the content of your own comment
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Param request body dto.UpdateCommentRequest true "New content"
// @Success 200 {object} object "Updated comment"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "Not authorized"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /comments/{commentId} [patch]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(int32)

	comment, err := h.commentSvc.Update(c.Request.Context(), userID, commentID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		case errors.Is(err, service.ErrNotCommentOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to edit this comment"})
		default:
			log.Printf("update comment error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mapper.ToCommentWithUserResponse(comment)})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete your own comment
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Success 200 {object} map[string]string "Comment deleted"
// @Failure 400 {object} map[string]string "Invalid comment ID"
// @Failure 403 {object} map[string]string "Not authorized"
// @Failure 404 {object} map[string]string "Comment not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /comments/{commentId} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.commentSvc.Delete(c.Request.Context(), userID, commentID); err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		case errors.Is(err, service.ErrNotCommentOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this comment"})
		default:
			log.Printf("delete comment error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

// GetMovieCommentCount godoc
// @Summary Get movie comment count
// @Description Get the total number of comments for a movie
// @Tags comments
// @Produce json
// @Param movieId path int true "Movie ID"
// @Success 200 {object} map[string]int "Comment count"
// @Failure 400 {object} map[string]string "Invalid movie ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /movies/{movieId}/comments/count [get]
func (h *CommentHandler) GetMovieCommentCount(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	count, err := h.commentSvc.CountMovieComments(c.Request.Context(), int32(movieID))
	if err != nil {
		log.Printf("get movie comment count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
