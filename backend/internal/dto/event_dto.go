package dto

import "time"

type TicketTypeInput struct {
	Name  string  `json:"name" binding:"required,max=100"`
	Price float64 `json:"price" binding:"min=0"`
	Quota int     `json:"quota" binding:"required,min=1"`
}

type CreateEventRequest struct {
	Title       string            `json:"title" binding:"required,min=3,max=200"`
	CategoryID  uint              `json:"category_id" binding:"required"`
	Description string            `json:"description" binding:"required"`
	Location    string            `json:"location" binding:"required"`
	City        string            `json:"city" binding:"required"`
	IsPaid      bool              `json:"is_paid"`
	Price       float64           `json:"price" binding:"min=0"`
	StartDate   time.Time         `json:"start_date" binding:"required"`
	EndDate     time.Time         `json:"end_date" binding:"required"`
	TotalSeats  int               `json:"total_seats" binding:"required,min=1"`
	BannerURL   string            `json:"banner_url"`
	TicketTypes []TicketTypeInput `json:"ticket_types"`
}

type UpdateEventRequest struct {
	Title       string    `json:"title" binding:"required,min=3,max=200"`
	CategoryID  uint      `json:"category_id" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Location    string    `json:"location" binding:"required"`
	City        string    `json:"city" binding:"required"`
	IsPaid      bool      `json:"is_paid"`
	Price       float64   `json:"price" binding:"min=0"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	BannerURL   string    `json:"banner_url"`
	Status      string    `json:"status" binding:"omitempty,oneof=draft published completed cancelled"`
}

type CreateVoucherRequest struct {
	Code          string    `json:"code" binding:"required,min=3,max=30"`
	DiscountType  string    `json:"discount_type" binding:"required,oneof=amount percentage"`
	DiscountValue float64   `json:"discount_value" binding:"required,min=1"`
	Quota         int       `json:"quota" binding:"required,min=1"`
	StartDate     time.Time `json:"start_date" binding:"required"`
	EndDate       time.Time `json:"end_date" binding:"required"`
}

type CreateReviewRequest struct {
	TransactionID uint   `json:"transaction_id" binding:"required"`
	Rating        int    `json:"rating" binding:"required,min=1,max=5"`
	Comment       string `json:"comment" binding:"max=2000"`
}
