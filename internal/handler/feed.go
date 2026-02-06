package handler

import (
	"strconv"

	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/response"

	"github.com/gin-gonic/gin"
)

type FeedHandler struct {
	feedRepo *repository.FeedRepository
}

func NewFeedHandler(feedRepo *repository.FeedRepository) *FeedHandler {
	return &FeedHandler{feedRepo: feedRepo}
}

type FeedResponse struct {
	Items      []repository.FeedItemDetail `json:"items"`
	TotalCount int                         `json:"total_count"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
}

// GetFeed godoc
// @Summary Get user feed
// @Description Get paginated feed items for the authenticated user
// @Tags feed
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} FeedResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/feed [get]
func (h *FeedHandler) GetFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "5"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 5
	}

	offset := (page - 1) * pageSize

	items, err := h.feedRepo.GetUserFeed(userID.(int64), pageSize, offset)
	if err != nil {
		response.InternalServerError(c, "failed to get feed")
		return
	}

	response.Success(c, FeedResponse{
		Items:      items,
		TotalCount: len(items),
		Page:       page,
		PageSize:   pageSize,
	})
}
