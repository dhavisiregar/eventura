package handlers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/testutil"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newEventTestRouter(db *gorm.DB) *gin.Engine {
	h := NewEventHandler(db)
	r := gin.New()
	r.GET("/events", h.List)
	r.GET("/events/:slug", h.Detail)
	// Simulate an authenticated organizer without going through real JWT middleware.
	r.POST("/organizer/events", func(c *gin.Context) { c.Set(middleware.CtxUserID, uint(1)); h.Create(c) })
	r.PUT("/organizer/events/:id", func(c *gin.Context) { c.Set(middleware.CtxUserID, uint(1)); h.Update(c) })
	r.DELETE("/organizer/events/:id", func(c *gin.Context) { c.Set(middleware.CtxUserID, uint(1)); h.Delete(c) })
	return r
}

func seedEvents(t *testing.T, db *gorm.DB, n int, category models.Category) {
	t.Helper()
	for i := 0; i < n; i++ {
		title := fmt.Sprintf("Music Fest %d", i)
		e := models.Event{
			OrganizerID: 1, CategoryID: category.ID, Title: title,
			Slug:        utils.Slugify(fmt.Sprintf("%s-%d", title, time.Now().UnixNano())),
			Description: "desc", Location: "loc", City: "Jakarta",
			TotalSeats: 100, AvailableSeats: 100, Status: models.EventPublished,
			StartDate: time.Now().Add(time.Duration(i+1) * time.Hour), EndDate: time.Now().Add(time.Duration(i+2) * time.Hour),
		}
		if err := db.Create(&e).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func TestListEvents_Pagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	cat := models.Category{Name: "Music", Slug: "music"}
	db.Create(&cat)
	seedEvents(t, db, 15, cat)

	r := newEventTestRouter(db)
	w := doJSON(r, http.MethodGet, "/events?page=2&limit=10", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []models.Event   `json:"data"`
		Meta utils.Pagination `json:"meta"`
	}
	mustDecode(t, w, &resp)

	if len(resp.Data) != 5 {
		t.Errorf("expected 5 events on page 2 (15 total, limit 10), got %d", len(resp.Data))
	}
	if resp.Meta.Total != 15 {
		t.Errorf("expected total 15, got %d", resp.Meta.Total)
	}
	if resp.Meta.TotalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", resp.Meta.TotalPages)
	}
}

func TestListEvents_SearchByTitle(t *testing.T) {
	db := testutil.NewTestDB(t)
	cat := models.Category{Name: "Music", Slug: "music"}
	db.Create(&cat)
	db.Create(&models.Event{
		OrganizerID: 1, CategoryID: cat.ID, Title: "Jazz Night", Slug: "jazz-night",
		Description: "an evening of jazz", Location: "loc", City: "Bandung",
		TotalSeats: 50, AvailableSeats: 50, Status: models.EventPublished,
		StartDate: time.Now().Add(time.Hour), EndDate: time.Now().Add(2 * time.Hour),
	})
	db.Create(&models.Event{
		OrganizerID: 1, CategoryID: cat.ID, Title: "Tech Summit", Slug: "tech-summit",
		Description: "technology talks", Location: "loc", City: "Bandung",
		TotalSeats: 50, AvailableSeats: 50, Status: models.EventPublished,
		StartDate: time.Now().Add(time.Hour), EndDate: time.Now().Add(2 * time.Hour),
	})

	r := newEventTestRouter(db)
	w := doJSON(r, http.MethodGet, "/events?q=jazz", nil)

	var resp struct {
		Data []models.Event `json:"data"`
	}
	mustDecode(t, w, &resp)
	if len(resp.Data) != 1 || resp.Data[0].Title != "Jazz Night" {
		t.Errorf("expected search to return only 'Jazz Night', got %+v", resp.Data)
	}
}

func TestListEvents_NoResultsReturnsEmptyArray(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newEventTestRouter(db)

	w := doJSON(r, http.MethodGet, "/events?q=nonexistent", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with no matches, got %d", w.Code)
	}

	var resp struct {
		Data []models.Event `json:"data"`
	}
	mustDecode(t, w, &resp)
	if len(resp.Data) != 0 {
		t.Errorf("expected an empty array, got %d items", len(resp.Data))
	}
}

func TestListEvents_FilterByCity(t *testing.T) {
	db := testutil.NewTestDB(t)
	cat := models.Category{Name: "Music", Slug: "music"}
	db.Create(&cat)
	db.Create(&models.Event{
		OrganizerID: 1, CategoryID: cat.ID, Title: "Jakarta Show", Slug: "jakarta-show",
		Description: "x", Location: "loc", City: "Jakarta",
		TotalSeats: 10, AvailableSeats: 10, Status: models.EventPublished,
		StartDate: time.Now().Add(time.Hour), EndDate: time.Now().Add(2 * time.Hour),
	})
	db.Create(&models.Event{
		OrganizerID: 1, CategoryID: cat.ID, Title: "Bali Show", Slug: "bali-show",
		Description: "x", Location: "loc", City: "Bali",
		TotalSeats: 10, AvailableSeats: 10, Status: models.EventPublished,
		StartDate: time.Now().Add(time.Hour), EndDate: time.Now().Add(2 * time.Hour),
	})

	r := newEventTestRouter(db)
	w := doJSON(r, http.MethodGet, "/events?city=Bali", nil)

	var resp struct {
		Data []models.Event `json:"data"`
	}
	mustDecode(t, w, &resp)
	if len(resp.Data) != 1 || resp.Data[0].City != "Bali" {
		t.Errorf("expected only the Bali event, got %+v", resp.Data)
	}
}

func TestCreateEvent_GeneratesUniqueSlugAndTicketTypes(t *testing.T) {
	db := testutil.NewTestDB(t)
	cat := models.Category{Name: "Tech", Slug: "tech"}
	db.Create(&cat)

	r := newEventTestRouter(db)
	w := doJSON(r, http.MethodPost, "/organizer/events", map[string]any{
		"title": "Go Conf", "category_id": cat.ID, "description": "desc",
		"location": "loc", "city": "Jakarta", "is_paid": true, "price": 150000,
		"start_date": time.Now().Add(48 * time.Hour), "end_date": time.Now().Add(50 * time.Hour),
		"total_seats": 200,
		"ticket_types": []map[string]any{
			{"name": "Regular", "price": 150000, "quota": 150},
			{"name": "VIP", "price": 300000, "quota": 50},
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var event models.Event
	db.Preload("TicketTypes").Where("title = ?", "Go Conf").First(&event)
	if event.Slug == "" {
		t.Error("expected a generated slug")
	}
	if len(event.TicketTypes) != 2 {
		t.Errorf("expected 2 ticket types, got %d", len(event.TicketTypes))
	}
}

func TestUpdateEvent_ForbiddenForNonOwner(t *testing.T) {
	db := testutil.NewTestDB(t)
	cat := models.Category{Name: "Tech", Slug: "tech"}
	db.Create(&cat)
	event := models.Event{
		OrganizerID: 999, CategoryID: cat.ID, Title: "Someone Else's Event", Slug: "someone-elses-event",
		Description: "d", Location: "l", City: "c", TotalSeats: 10, AvailableSeats: 10,
		StartDate: time.Now().Add(time.Hour), EndDate: time.Now().Add(2 * time.Hour), Status: models.EventPublished,
	}
	db.Create(&event)

	r := newEventTestRouter(db)
	w := doJSON(r, http.MethodPut, fmt.Sprintf("/organizer/events/%d", event.ID), map[string]any{
		"title": "Hijacked", "category_id": cat.ID, "description": "d", "location": "l", "city": "c",
		"start_date": time.Now().Add(time.Hour), "end_date": time.Now().Add(2 * time.Hour),
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a non-owner update, got %d", w.Code)
	}
}
