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
