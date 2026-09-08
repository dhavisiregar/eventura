package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type seriesPoint struct {
	Label   string  `json:"label"`
	Revenue float64 `json:"revenue"`
	Tickets int     `json:"tickets_sold"`
}

// Stats returns summary numbers plus a time-bucketed series for charting,
// grouped per year (months), per month (days), or per day (hours).
func (h *DashboardHandler) Stats(c *gin.Context) {
	organizerID := middleware.UserID(c)
	rangeType := c.DefaultQuery("range", "year")
	now := time.Now()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	day, _ := strconv.Atoi(c.DefaultQuery("day", strconv.Itoa(now.Day())))

	var from, to time.Time
	var dateFormat string
	switch rangeType {
	case "month":
		from = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
		dateFormat = "%Y-%m-%d"
	case "day":
		from = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 0, 1)
		dateFormat = "%Y-%m-%d %H:00"
	default: // "year"
		rangeType = "year"
		from = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(1, 0, 0)
		dateFormat = "%Y-%m"
	}

	baseQuery := h.db.Table("transactions").
		Joins("JOIN events ON events.id = transactions.event_id").
		Where("events.organizer_id = ? AND transactions.status = ? AND transactions.created_at >= ? AND transactions.created_at < ?",
			organizerID, models.TxSuccess, from, to)

	var summary struct {
		TotalRevenue float64
		TotalTickets int
		TotalOrders  int
	}
	baseQuery.Session(&gorm.Session{}).
		Select("COALESCE(SUM(transactions.total_price),0) as total_revenue, COALESCE(SUM(transactions.quantity),0) as total_tickets, COUNT(*) as total_orders").
		Scan(&summary)

	var totalEvents int64
	h.db.Model(&models.Event{}).Where("organizer_id = ?", organizerID).Count(&totalEvents)

	series := []seriesPoint{}
	rows, err := baseQuery.Session(&gorm.Session{}).
		Select(fmt.Sprintf("DATE_FORMAT(transactions.created_at, '%s') as label, COALESCE(SUM(transactions.total_price),0) as revenue, COALESCE(SUM(transactions.quantity),0) as tickets", dateFormat)).
		Group("label").Order("label ASC").Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p seriesPoint
			if scanErr := rows.Scan(&p.Label, &p.Revenue, &p.Tickets); scanErr == nil {
				series = append(series, p)
			}
		}
	}

	type topEvent struct {
		Title   string  `json:"title"`
		Tickets int     `json:"tickets_sold"`
		Revenue float64 `json:"revenue"`
	}
	topEvents := []topEvent{}
	baseQuery.Session(&gorm.Session{}).
		Select("events.title as title, COALESCE(SUM(transactions.quantity),0) as tickets, COALESCE(SUM(transactions.total_price),0) as revenue").
		Group("events.id, events.title").Order("revenue DESC").Limit(5).Scan(&topEvents)

	utils.Success(c, http.StatusOK, gin.H{
		"range": rangeType,
		"from":  from,
		"to":    to,
		"summary": gin.H{
			"total_events":  totalEvents,
			"total_revenue": summary.TotalRevenue,
			"total_tickets": summary.TotalTickets,
			"total_orders":  summary.TotalOrders,
		},
		"series":     series,
		"top_events": topEvents,
	})
}
