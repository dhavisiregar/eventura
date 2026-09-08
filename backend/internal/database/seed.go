package database

import (
	"log"

	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"gorm.io/gorm"
)

var defaultCategories = []string{
	"Music", "Technology", "Business", "Sports", "Arts & Culture",
	"Food & Drink", "Education", "Health & Wellness", "Community", "Film & Media",
}

// Seed inserts baseline reference data (categories) if the table is empty.
// Safe to call on every boot.
func Seed(db *gorm.DB) {
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count > 0 {
		return
	}

	for _, name := range defaultCategories {
		cat := models.Category{Name: name, Slug: utils.Slugify(name)}
		if err := db.Create(&cat).Error; err != nil {
			log.Printf("seed: failed to create category %s: %v", name, err)
		}
	}
	log.Println("seed: default categories created")
}
