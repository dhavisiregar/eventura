package services

import (
	"testing"

	"eventman/backend/internal/models"
	"eventman/backend/internal/testutil"
)

func TestUniqueEventSlug_NoCollision(t *testing.T) {
	db := testutil.NewTestDB(t)

	slug, err := UniqueEventSlug(db, "Go Conference 2024")
	if err != nil {
		t.Fatalf("UniqueEventSlug returned error: %v", err)
	}
	if slug != "go-conference-2024" {
		t.Errorf("expected 'go-conference-2024', got %q", slug)
	}
}

func TestUniqueEventSlug_AppendsSuffixOnCollision(t *testing.T) {
	db := testutil.NewTestDB(t)

	db.Create(&models.Event{
		OrganizerID: 1, CategoryID: 1, Title: "Go Conference 2024", Slug: "go-conference-2024",
		TotalSeats: 10, AvailableSeats: 10,
	})

	slug, err := UniqueEventSlug(db, "Go Conference 2024")
	if err != nil {
		t.Fatalf("UniqueEventSlug returned error: %v", err)
	}
	if slug != "go-conference-2024-2" {
		t.Errorf("expected a de-duplicated slug 'go-conference-2024-2', got %q", slug)
	}
}
