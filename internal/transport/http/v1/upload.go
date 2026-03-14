package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/domain"
)

type UploadHandler struct {
	storage domain.StorageService
}

// NewUploadHandler creates a new handler for media uploads.
func NewUploadHandler(storage domain.StorageService) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// UploadPhoto handles multipart/form-data image uploads from clients.
//
//	@Summary      Upload a photo
//	@Description  Compresses the image to JPEG, uploads to MinIO, and returns the public URL.
//	@Tags         media
//	@Accept       multipart/form-data
//	@Produce      json
//	@Security     RoleAuth
//	@Param        file  formData  file  true  "Image file to upload"
//	@Success      200   {object}  map[string]string "Returns the public URL of the uploaded image"
//	@Router       /staff/lots/upload [post]
func (h *UploadHandler) UploadPhoto(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required in the form data"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	fileURL, err := h.storage.UploadPhoto(c.Request.Context(), file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process and upload photo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"url":     fileURL,
	})
}

// DeletePhotoRequest is the payload for deleting a photo.
type DeletePhotoRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// DeletePhoto removes a previously uploaded image from storage.
//
//	@Summary      Delete a photo
//	@Description  Removes the image from MinIO using its public URL.
//	@Tags         media
//	@Accept       json
//	@Produce      json
//	@Security     RoleAuth
//	@Param        request  body      DeletePhotoRequest  true  "Public URL of the photo to delete"
//	@Success      200      {object}  map[string]string
//	@Router       /staff/lots/photo [delete]
func (h *UploadHandler) DeletePhoto(c *gin.Context) {
	var req DeletePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	if err := h.storage.DeletePhoto(c.Request.Context(), req.URL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete photo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "photo deleted successfully"})
}
