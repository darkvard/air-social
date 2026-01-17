package handler

import (
	"path/filepath"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type MediaHandler struct {
	srv service.MediaService
}

func NewMediaHandler(srv service.MediaService) *MediaHandler {
	return &MediaHandler{srv: srv}
}

// PresignedUpload godoc
//
//	@Summary		Get presigned upload URL
//	@Description	Generate a presigned URL for uploading a file to object storage.
//	@Tags			Media
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.PresignedFileUploadRequest	true	"Presigned Upload Request"
//	@Success		200		{object}	domain.PresignedFileResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Router			/media/presigned [post]
func (h *MediaHandler) PresignedUpload(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.PresignedFileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}
	if v := pkg.ValidateImageFile(req.FileName); v != nil {
		pkg.HandleValidationResult(c, v)
		return
	}

	folder := "misc"
	switch domain.FileType(req.FileType) {
	case domain.AvatarType, domain.CoverType:
		folder = "users"
	}

	res, err := h.srv.GetPresignedURL(
		c.Request.Context(),
		domain.PresignedFile{
			UserID: payload.UserID,
			Ext:    filepath.Ext(req.FileName),
			Folder: folder,
			Typ:    domain.FileType(req.FileType),
		},
	)

	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, res)
}
