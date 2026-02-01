package handler

import (
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
//	@Description	Generates a presigned URL and policy signature for direct file upload to MinIO/S3.
//	@Description
//	@Description	**Client Integration:**
//	@Description	* **Target:** Send a `POST` request to the returned `upload_url`.
//	@Description	* **Payload:** Use `multipart/form-data`. Append all fields from `form_data` (must be FIRST) followed by the file binary (must be LAST).
//	@Description	* **Success:** Upon successful upload (HTTP 204), confirm the `object_key` with the Backend database.
//	@Tags			Media
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.PresignedFileUploadRequest	true	"File Metadata"
//	@Success		200		{object}	domain.PresignedFileResponse		"Upload Credentials"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/media/presigned [post]
func (h *MediaHandler) PresignedUpload(c *gin.Context) {
	payload, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.PresignedFileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	res, err := h.srv.GetPresignedURL(
		c.Request.Context(),
		domain.PresignedFileParams{
			EntityID: payload.UserID,
			FileName: req.FileName,
			FileType: req.FileType,
			FileSize: req.FileSize,
			Domain:   req.Domain,
			Feature:  req.Feature,
		},
	)

	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, res)
}
