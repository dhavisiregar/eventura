package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	cfg *config.Config
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

var allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

// Upload saves a multipart "file" into uploads/<kind>/ and returns its public URL.
// kind is restricted to a small allowlist to avoid arbitrary path writes.
func (h *UploadHandler) Upload(kind string) gin.HandlerFunc {
	allowedKinds := map[string]bool{"banners": true, "avatars": true}
	return func(c *gin.Context) {
		if !allowedKinds[kind] {
			utils.Error(c, http.StatusBadRequest, "invalid upload kind")
			return
		}

		fileHeader, err := c.FormFile("file")
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "file is required")
			return
		}
		if fileHeader.Size > 5*1024*1024 {
			utils.Error(c, http.StatusBadRequest, "file must be 5MB or smaller")
			return
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !allowedImageExt[ext] {
			utils.Error(c, http.StatusBadRequest, "only jpg, jpeg, png, or webp images are allowed")
			return
		}

		filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), utils.RandomCode(6), ext)
		destDir := filepath.Join(h.cfg.UploadDir, kind)
		destPath := filepath.Join(destDir, filename)

		if err := c.SaveUploadedFile(fileHeader, destPath); err != nil {
			utils.Error(c, http.StatusInternalServerError, "failed to save file")
			return
		}

		utils.Success(c, http.StatusCreated, gin.H{"url": fmt.Sprintf("/uploads/%s/%s", kind, filename)})
	}
}
