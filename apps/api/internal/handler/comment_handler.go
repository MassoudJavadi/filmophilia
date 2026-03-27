package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/pagination"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
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
// @Success 201 {object} dto.SuccessResponse{data=dto.CommentResponse} "Created comment"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)

	comment, err := h.commentSvc.Create(c.Request.Context(), userID, int32(movieID), nil, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			response.Error(c, http.StatusNotFound, "MOVIE_NOT_FOUND", "movie not found")
		default:
			log.Printf("create comment error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Created(c, mapper.ToCommentWithUserResponse(comment))
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
// @Success 201 {object} dto.SuccessResponse{data=dto.CommentResponse} "Created reply"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request or cannot reply to reply"
// @Failure 404 {object} dto.NotFoundErrorResponse "Movie or parent comment not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/comments/{commentId}/replies [post]
func (h *CommentHandler) CreateReply(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)
	parentID := commentID

	comment, err := h.commentSvc.Create(c.Request.Context(), userID, int32(movieID), &parentID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMovieNotFound):
			response.Error(c, http.StatusNotFound, "MOVIE_NOT_FOUND", "movie not found")
		case errors.Is(err, service.ErrParentNotFound):
			response.Error(c, http.StatusNotFound, "PARENT_COMMENT_NOT_FOUND", "parent comment not found")
		case errors.Is(err, service.ErrCannotReplyReply):
			response.Error(c, http.StatusBadRequest, "CANNOT_REPLY_TO_REPLY", "cannot reply to a reply")
		default:
			log.Printf("create reply error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Created(c, mapper.ToCommentWithUserResponse(comment))
}

// GetComment godoc
// @Summary Get a comment
// @Description Get a specific comment by ID
// @Tags comments
// @Produce json
// @Param commentId path int true "Comment ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.CommentResponse} "Comment details"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 404 {object} dto.NotFoundErrorResponse "Comment not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId} [get]
func (h *CommentHandler) GetComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	comment, err := h.commentSvc.GetByID(c.Request.Context(), commentID)
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			response.Error(c, http.StatusNotFound, "COMMENT_NOT_FOUND", "comment not found")
			return
		}
		log.Printf("get comment error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToCommentWithUserResponse(comment))
}

// GetMovieComments godoc
// @Summary Get movie comments
// @Description Get all top-level comments for a movie
// @Tags comments
// @Produce json
// @Param movieId path int true "Movie ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.CommentResponse} "List of comments"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/comments [get]
func (h *CommentHandler) GetMovieComments(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	p := pagination.Parse(c)

	comments, err := h.commentSvc.GetMovieComments(c.Request.Context(), int32(movieID), p.Limit, p.Offset)
	if err != nil {
		log.Printf("get movie comments error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToCommentResponses(comments))
}

// GetReplies godoc
// @Summary Get comment replies
// @Description Get all replies to a comment
// @Tags comments
// @Produce json
// @Param commentId path int true "Comment ID"
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.CommentResponse} "List of replies"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId}/replies [get]
func (h *CommentHandler) GetReplies(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	p := pagination.Parse(c)

	replies, err := h.commentSvc.GetReplies(c.Request.Context(), commentID, p.Limit, p.Offset)
	if err != nil {
		log.Printf("get replies error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToReplyResponses(replies))
}

// GetMyComments godoc
// @Summary Get my comments
// @Description Get all comments made by the authenticated user
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.SuccessResponse{data=[]dto.CommentWithMovieResponse} "List of comments"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /me/comments [get]
func (h *CommentHandler) GetMyComments(c *gin.Context) {
	userID := c.MustGet("user_id").(int32)

	p := pagination.Parse(c)

	comments, err := h.commentSvc.GetUserComments(c.Request.Context(), userID, p.Limit, p.Offset)
	if err != nil {
		log.Printf("get my comments error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, mapper.ToCommentWithMovieResponses(comments))
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
// @Success 200 {object} dto.SuccessResponse{data=dto.CommentResponse} "Updated comment"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid request"
// @Failure 403 {object} dto.ForbiddenErrorResponse "Not authorized"
// @Failure 404 {object} dto.NotFoundErrorResponse "Comment not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId} [patch]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	userID := c.MustGet("user_id").(int32)

	comment, err := h.commentSvc.Update(c.Request.Context(), userID, commentID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			response.Error(c, http.StatusNotFound, "COMMENT_NOT_FOUND", "comment not found")
		case errors.Is(err, service.ErrNotCommentOwner):
			response.Error(c, http.StatusForbidden, "NOT_COMMENT_OWNER", "not authorized to edit this comment")
		default:
			log.Printf("update comment error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.OK(c, mapper.ToCommentWithUserResponse(comment))
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete your own comment
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageData} "Comment deleted"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid comment ID"
// @Failure 403 {object} dto.ForbiddenErrorResponse "Not authorized"
// @Failure 404 {object} dto.NotFoundErrorResponse "Comment not found"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /comments/{commentId} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_COMMENT_ID", "invalid comment id")
		return
	}

	userID := c.MustGet("user_id").(int32)

	if err := h.commentSvc.Delete(c.Request.Context(), userID, commentID); err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			response.Error(c, http.StatusNotFound, "COMMENT_NOT_FOUND", "comment not found")
		case errors.Is(err, service.ErrNotCommentOwner):
			response.Error(c, http.StatusForbidden, "NOT_COMMENT_OWNER", "not authorized to delete this comment")
		default:
			log.Printf("delete comment error: %v", err)
			response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		}
		return
	}

	response.Message(c, http.StatusOK, "comment deleted")
}

// GetMovieCommentCount godoc
// @Summary Get movie comment count
// @Description Get the total number of comments for a movie
// @Tags comments
// @Produce json
// @Param movieId path int true "Movie ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.CountData} "Comment count"
// @Failure 400 {object} dto.BadRequestErrorResponse "Invalid movie ID"
// @Failure 500 {object} dto.InternalServerErrorResponse "Internal server error"
// @Router /movies/{movieId}/comments/count [get]
func (h *CommentHandler) GetMovieCommentCount(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MOVIE_ID", "invalid movie id")
		return
	}

	count, err := h.commentSvc.CountMovieComments(c.Request.Context(), int32(movieID))
	if err != nil {
		log.Printf("get movie comment count error: %v", err)
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
		return
	}

	response.OK(c, dto.CountData{Count: count})
}
