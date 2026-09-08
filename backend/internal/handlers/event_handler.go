package handlers

import (
	"net/http"
	"strconv"
	"time"

	"eventman/backend/internal/dto"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/services"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventHandler struct {
	db *gorm.DB
}

func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{db: db}
}

// List handles GET /events with search (?q=), category & city filters, and pagination.
func (h *EventHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	query := h.db.Model(&models.Event{}).Where("status = ?", models.EventPublished)

	if q := c.Query("q"); q != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Joins("JOIN categories ON categories.id = events.category_id").
			Where("categories.slug = ?", category)
	}
	if city := c.Query("city"); city != "" {
		query = query.Where("city = ?", city)
	}
	if isPaid := c.Query("is_paid"); isPaid != "" {
		query = query.Where("is_paid = ?", isPaid == "true")
	}
	// Only show events that haven't finished, by default.
	if c.DefaultQuery("include_past", "false") != "true" {
		query = query.Where("end_date >= ?", time.Now())
	}

	sort := "start_date ASC"
	switch c.Query("sort") {
	case "price_asc":
		sort = "price ASC"
	case "price_desc":
		sort = "price DESC"
	case "newest":
		sort = "created_at DESC"
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to count events")
		return
	}

	events := []models.Event{}
	if err := query.Preload("Category").Preload("Organizer").
		Order(sort).Offset((page - 1) * limit).Limit(limit).Find(&events).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, events, utils.NewPagination(page, limit, total))
}

func (h *EventHandler) Detail(c *gin.Context) {
	slug := c.Param("slug")

	var event models.Event
	if err := h.db.Preload("Category").Preload("Organizer").Preload("TicketTypes").
		Where("slug = ?", slug).First(&event).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "event not found")
		return
	}

	var avgRating float64
	var reviewCount int64
	h.db.Model(&models.Review{}).Where("event_id = ?", event.ID).
		Select("COALESCE(AVG(rating),0)").Scan(&avgRating)
	h.db.Model(&models.Review{}).Where("event_id = ?", event.ID).Count(&reviewCount)

	utils.Success(c, http.StatusOK, gin.H{
		"event":          event,
		"average_rating": avgRating,
		"review_count":   reviewCount,
	})
}

func (h *EventHandler) Reviews(c *gin.Context) {
	slug := c.Param("slug")

	var event models.Event
	if err := h.db.Where("slug = ?", slug).First(&event).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "event not found")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var total int64
	h.db.Model(&models.Review{}).Where("event_id = ?", event.ID).Count(&total)

	reviews := []models.Review{}
	h.db.Preload("User").Where("event_id = ?", event.ID).
		Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&reviews)

	utils.SuccessWithMeta(c, http.StatusOK, reviews, utils.NewPagination(page, limit, total))
}

func (h *EventHandler) ListCategories(c *gin.Context) {
	categories := []models.Category{}
	h.db.Order("name ASC").Find(&categories)
	utils.Success(c, http.StatusOK, categories)
}

// --- Organizer-only endpoints ---

func (h *EventHandler) Create(c *gin.Context) {
	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.EndDate.Before(req.StartDate) {
		utils.Error(c, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	slug, err := services.UniqueEventSlug(h.db, req.Title)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate slug")
		return
	}

	event := models.Event{
		OrganizerID:    middleware.UserID(c),
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Slug:           slug,
		Description:    req.Description,
		Location:       req.Location,
		City:           req.City,
		IsPaid:         req.IsPaid,
		Price:          req.Price,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.TotalSeats,
		BannerURL:      req.BannerURL,
		Status:         models.EventPublished,
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		for _, t := range req.TicketTypes {
			tt := models.TicketType{
				EventID: event.ID, Name: t.Name, Price: t.Price,
				Quota: t.Quota, Remaining: t.Quota,
			}
			if err := tx.Create(&tt).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to create event: "+err.Error())
		return
	}

	h.db.Preload("Category").Preload("TicketTypes").First(&event, event.ID)
	utils.Success(c, http.StatusCreated, event)
}

func (h *EventHandler) Get(c *gin.Context) {
	event, ok := h.loadOwnedEvent(c)
	if !ok {
		return
	}
	h.db.Preload("Category").Preload("TicketTypes").First(event, event.ID)
	utils.Success(c, http.StatusOK, event)
}

func (h *EventHandler) loadOwnedEvent(c *gin.Context) (*models.Event, bool) {
	id := c.Param("id")
	var event models.Event
	if err := h.db.First(&event, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "event not found")
		return nil, false
	}
	if event.OrganizerID != middleware.UserID(c) {
		utils.Error(c, http.StatusForbidden, "you do not own this event")
		return nil, false
	}
	return &event, true
}

func (h *EventHandler) Update(c *gin.Context) {
	event, ok := h.loadOwnedEvent(c)
	if !ok {
		return
	}

	var req dto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.EndDate.Before(req.StartDate) {
		utils.Error(c, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	updates := map[string]any{
		"title": req.Title, "category_id": req.CategoryID, "description": req.Description,
		"location": req.Location, "city": req.City, "is_paid": req.IsPaid, "price": req.Price,
		"start_date": req.StartDate, "end_date": req.EndDate, "banner_url": req.BannerURL,
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.db.Model(event).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to update event")
		return
	}

	h.db.Preload("Category").Preload("TicketTypes").First(event, event.ID)
	utils.Success(c, http.StatusOK, event)
}

func (h *EventHandler) Delete(c *gin.Context) {
	event, ok := h.loadOwnedEvent(c)
	if !ok {
		return
	}

	var soldCount int64
	h.db.Model(&models.Transaction{}).
		Where("event_id = ? AND status IN ?", event.ID, []models.TxStatus{models.TxSuccess, models.TxPendingPayment}).
		Count(&soldCount)
	if soldCount > 0 {
		utils.Error(c, http.StatusConflict, "cannot delete an event with existing transactions; cancel it instead")
		return
	}

	if err := h.db.Delete(event).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to delete event")
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *EventHandler) MyEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&models.Event{}).Where("organizer_id = ?", middleware.UserID(c))

	var total int64
	query.Count(&total)

	events := []models.Event{}
	query.Preload("Category").Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&events)

	utils.SuccessWithMeta(c, http.StatusOK, events, utils.NewPagination(page, limit, total))
}

func (h *EventHandler) Attendees(c *gin.Context) {
	event, ok := h.loadOwnedEvent(c)
	if !ok {
		return
	}

	transactions := []models.Transaction{}
	h.db.Preload("User").Where("event_id = ? AND status = ?", event.ID, models.TxSuccess).
		Order("created_at DESC").Find(&transactions)

	utils.Success(c, http.StatusOK, transactions)
}
