package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/service"
)

type FollowHandler struct {
	srv service.FollowService
}

func NewFollowHandler(srv service.FollowService) *FollowHandler {
	return &FollowHandler{
		srv: srv,
	}
}

func (h *FollowHandler) Follow(c *gin.Context) {

}

func (h *FollowHandler) Unfollow(c *gin.Context) {

}

func (h *FollowHandler) GetFollowers(c *gin.Context) {

}

func (h *FollowHandler) GetFollowings(c *gin.Context) {

}

