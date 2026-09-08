package services

import (
	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"gorm.io/gorm"
)

// UniqueEventSlug builds a URL-safe slug from title, appending a numeric
// suffix if it already exists.
func UniqueEventSlug(db *gorm.DB, title string) (string, error) {
	base := utils.Slugify(title)
	if base == "" {
		base = "event"
	}
	slug := base
	for i := 2; ; i++ {
		var count int64
		if err := db.Model(&models.Event{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = utils.UniqueSlugSuffix(base, i)
	}
}
