package media

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/media"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  media.UseCase
}

func NewHandler(provider common.LinkProvider, usecase media.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
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
//	@Param			request	body		BulkPresignRequest	true	"File Metadata List"
//	@Success		200		{array}		PresignResponse		"Upload Credentials List"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/media/presigned-urls [post]
func (h Handler) PresignedUpload(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req BulkPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := h.toPresignParams(claims.UserID, req.Files)

	res, err := h.usecase.GetPresignedURL(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPresignResponse(res))
}

func (h Handler) toPresignParams(userID int64, files []PresignRequest) []media.PresignParams {
	params := make([]media.PresignParams, len(files))
	for i, f := range files {
		params[i] = media.PresignParams{
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

func (h Handler) toPresignResponse(files []media.PresignedResult) []PresignResponse {
	resp := make([]PresignResponse, len(files))
	for i, r := range files {
		resp[i] = PresignResponse{
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
