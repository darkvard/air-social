package user

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
)

type Handler struct {
	provider common.LinkProvider
	usecase  user.UseCase
}

func NewHandler(provider common.LinkProvider, usecase user.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}

func (h Handler) Profile(c *gin.Context) {

}

func (h Handler) UpdateProfile(c *gin.Context) {

}

func (h Handler) ChangePassword(c *gin.Context) {

}

func (h Handler) UpdateAvatar(c *gin.Context) {

}

func (h Handler) UpdateImageCover(c *gin.Context) {

}


// todo: merge user
//  - delete old layer
//	- merge usecase unit tests
//	- test endpoints