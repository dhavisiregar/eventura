package server

import (
	"os"

	"eventman/backend/internal/config"
	"eventman/backend/internal/handlers"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter wires every dependency and registers all routes.
func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(middleware.CORS(cfg.FrontendURL))

	_ = os.MkdirAll(cfg.UploadDir, 0o755)
	r.Static("/uploads", cfg.UploadDir)

	midtransSvc := services.NewMidtransService(cfg)
	txSvc := services.NewTransactionService(db, midtransSvc, cfg)

	authH := handlers.NewAuthHandler(db, cfg)
	eventH := handlers.NewEventHandler(db)
	voucherH := handlers.NewVoucherHandler(db)
	txH := handlers.NewTransactionHandler(db, txSvc)
	reviewH := handlers.NewReviewHandler(db)
	dashH := handlers.NewDashboardHandler(db)
	uploadH := handlers.NewUploadHandler(cfg)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api/v1")
	{
		// Public
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)
		api.GET("/categories", eventH.ListCategories)
		api.GET("/events", eventH.List)
		api.GET("/events/:slug", eventH.Detail)
		api.GET("/events/:slug/reviews", eventH.Reviews)
		api.POST("/midtrans/notification", func(c *gin.Context) {
			txH.MidtransNotification(c, midtransSvc.VerifySignature)
		})

		// Authenticated (any role)
		authed := api.Group("")
		authed.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			authed.GET("/auth/me", authH.Me)
			authed.GET("/referral/me", authH.ReferralInfo)
			authed.POST("/uploads/avatars", uploadH.Upload("avatars"))
		}

		// Customer only
		customer := api.Group("")
		customer.Use(middleware.RequireAuth(cfg.JWTSecret), middleware.RequireRole(models.RoleCustomer))
		{
			customer.POST("/transactions", txH.Checkout)
			customer.GET("/transactions/me", txH.MyTransactions)
			customer.GET("/transactions/:id", txH.Detail)
			customer.POST("/transactions/:id/cancel", txH.Cancel)
			customer.POST("/transactions/sync/:orderID", txH.SyncStatus)
			customer.POST("/reviews", reviewH.Create)
			customer.GET("/reviews/reviewable", reviewH.Reviewable)
		}

		// Organizer only
		organizer := api.Group("/organizer")
		organizer.Use(middleware.RequireAuth(cfg.JWTSecret), middleware.RequireRole(models.RoleOrganizer))
		{
			organizer.POST("/events", eventH.Create)
			organizer.PUT("/events/:id", eventH.Update)
			organizer.DELETE("/events/:id", eventH.Delete)
			organizer.GET("/events", eventH.MyEvents)
			organizer.GET("/events/:id", eventH.Get)
			organizer.GET("/events/:id/attendees", eventH.Attendees)
			organizer.POST("/events/:id/vouchers", voucherH.Create)
			organizer.GET("/events/:id/vouchers", voucherH.List)
			organizer.DELETE("/vouchers/:voucherId", voucherH.Delete)
			organizer.GET("/transactions", txH.OrganizerTransactions)
			organizer.GET("/dashboard/stats", dashH.Stats)
			organizer.POST("/uploads/banners", uploadH.Upload("banners"))
		}
	}

	return r
}
