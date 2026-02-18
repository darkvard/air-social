package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
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
//	@Description	Generates a list of presigned URLs and policy signatures for direct file upload to MinIO/S3.
//	@Description
//	@Description	**Client Integration:**
//	@Description	* **Target:** Send a `POST` request to the returned `upload_url`.
//	@Description	* **Payload:** Use `multipart/form-data`. Append all fields from `form_data` (must be FIRST) followed by the file binary (must be LAST).
//	@Description	* **Success:** Upon successful upload (HTTP 204), confirm the `object_key` with the Backend database.
//	@Tags			Media
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.BulkPresignedUploadRequest	true	"File Metadata List"
//	@Success		200		{array}		dto.PresignedFileResponse		"Upload Credentials List"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/media/presigned-urls [post]
func (h *MediaHandler) PresignedUpload(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req dto.BulkPresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := h.toPresignedParams(claims.UserID, req.Files)

	res, err := h.srv.GetPresignedURL(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPresignedResponse(res))
}

// Internal helper

func (h *MediaHandler) toPresignedParams(userID int64, files []dto.PresignedUploadRequest) []domain.PresignedFileParams {
	params := make([]domain.PresignedFileParams, len(files))
	for i, f := range files {
		params[i] = domain.PresignedFileParams{
			EntityID: userID,
			FileName: f.FileName,
			FileType: f.FileType,
			FileSize: f.FileSize,
			Domain:   f.Domain,
			Feature:  f.Feature,
		}
	}
	return params
}

func (h *MediaHandler) toPresignedResponse(files []domain.PresignedFile) []dto.PresignedFileResponse {
	resp := make([]dto.PresignedFileResponse, len(files))
	for i, r := range files {
		resp[i] = dto.PresignedFileResponse{
			FileName:  r.FileName,
			UploadURL: r.UploadURL,
			FormData:  r.FormData,
			ObjectKey: r.ObjectKey,
			PublicURL: r.PublicURL,
			ExpireAt:  r.ExpireAt,
		}
	}
	return resp
}
