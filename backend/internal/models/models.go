package models

import "time"

type Role string

const (
	RoleCustomer  Role = "customer"
	RoleOrganizer Role = "organizer"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:150;not null" json:"name"`
	Email        string    `gorm:"size:191;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         Role      `gorm:"type:varchar(20);not null;index" json:"role"`
	ReferralCode string    `gorm:"size:20;uniqueIndex;not null" json:"referral_code"`
	ReferredBy   *uint     `gorm:"index" json:"referred_by"`
	AvatarURL    string    `gorm:"size:255" json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Slug string `gorm:"size:100;uniqueIndex;not null" json:"slug"`
}

type EventStatus string

const (
	EventDraft     EventStatus = "draft"
	EventPublished EventStatus = "published"
	EventCompleted EventStatus = "completed"
	EventCancelled EventStatus = "cancelled"
)

type Event struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	OrganizerID    uint        `gorm:"index;not null" json:"organizer_id"`
	Organizer      *User       `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	CategoryID     uint        `gorm:"index;not null" json:"category_id"`
	Category       *Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Title          string      `gorm:"size:200;not null" json:"title"`
	Slug           string      `gorm:"size:220;uniqueIndex;not null" json:"slug"`
	Description    string      `gorm:"type:text" json:"description"`
	Location       string      `gorm:"size:255" json:"location"`
	City           string      `gorm:"size:120;index" json:"city"`
	IsPaid         bool        `gorm:"default:false" json:"is_paid"`
	Price          float64     `gorm:"type:decimal(12,2);default:0" json:"price"`
	StartDate      time.Time   `gorm:"index" json:"start_date"`
	EndDate        time.Time   `json:"end_date"`
	TotalSeats     int         `gorm:"not null" json:"total_seats"`
	AvailableSeats int         `gorm:"not null" json:"available_seats"`
	BannerURL      string      `gorm:"size:255" json:"banner_url"`
	Status         EventStatus `gorm:"type:varchar(20);default:'published';index" json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`

	TicketTypes []TicketType `gorm:"foreignKey:EventID" json:"ticket_types,omitempty"`
	Vouchers    []Voucher    `gorm:"foreignKey:EventID" json:"vouchers,omitempty"`
}

type TicketType struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	EventID   uint    `gorm:"index;not null" json:"event_id"`
	Name      string  `gorm:"size:100;not null" json:"name"`
	Price     float64 `gorm:"type:decimal(12,2);not null" json:"price"`
	Quota     int     `gorm:"not null" json:"quota"`
	Remaining int     `gorm:"not null" json:"remaining"`
}

type DiscountType string

const (
	DiscountAmount  DiscountType = "amount"
	DiscountPercent DiscountType = "percentage"
)

// Voucher is an organizer-issued promo code scoped to a single event.
// Can be quota-limited ("limited persons") and/or date-windowed
// (e.g. only valid in the weeks leading up to the event).
type Voucher struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	EventID       uint         `gorm:"uniqueIndex:idx_event_code;not null" json:"event_id"`
	Code          string       `gorm:"size:30;uniqueIndex:idx_event_code;not null" json:"code"`
	DiscountType  DiscountType `gorm:"type:varchar(20);not null" json:"discount_type"`
	DiscountValue float64      `gorm:"type:decimal(12,2);not null" json:"discount_value"`
	Quota         int          `gorm:"not null" json:"quota"`
	UsedCount     int          `gorm:"default:0" json:"used_count"`
	StartDate     time.Time    `json:"start_date"`
	EndDate       time.Time    `json:"end_date"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// Coupon is the 10% discount automatically granted to a user who registers
// using someone else's referral code. Valid for 3 months from creation.
type Coupon struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	UserID        uint         `gorm:"index;not null" json:"user_id"`
	Code          string       `gorm:"size:30;uniqueIndex;not null" json:"code"`
	DiscountType  DiscountType `gorm:"type:varchar(20);not null" json:"discount_type"`
	DiscountValue float64      `gorm:"type:decimal(12,2);not null" json:"discount_value"`
	Source        string       `gorm:"size:50;not null" json:"source"`
	IsUsed        bool         `gorm:"default:false" json:"is_used"`
	UsedAt        *time.Time   `json:"used_at"`
	ExpiresAt     time.Time    `gorm:"index" json:"expires_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

type PointType string

const (
	PointEarn   PointType = "earn"
	PointRedeem PointType = "redeem"
)

// PointTransaction is an append-only ledger. Balance is derived by summing
// earn rows that have not expired minus all redeemed points.
type PointTransaction struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Points      int        `gorm:"not null" json:"points"`
	Type        PointType  `gorm:"type:varchar(20);not null" json:"type"`
	Description string     `gorm:"size:255" json:"description"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TxStatus string

const (
	TxPendingPayment TxStatus = "pending_payment"
	TxSuccess        TxStatus = "success"
	TxExpired        TxStatus = "expired"
	TxCancelled      TxStatus = "cancelled"
	TxFailed         TxStatus = "failed"
)

type Transaction struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	InvoiceNo    string `gorm:"size:50;uniqueIndex;not null" json:"invoice_no"`
	UserID       uint   `gorm:"index;not null" json:"user_id"`
	User         *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EventID      uint   `gorm:"index;not null" json:"event_id"`
	Event        *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
	TicketTypeID *uint  `json:"ticket_type_id"`

	Quantity        int     `gorm:"not null" json:"quantity"`
	UnitPrice       float64 `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Subtotal        float64 `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	VoucherID       *uint   `json:"voucher_id"`
	VoucherDiscount float64 `gorm:"type:decimal(12,2);default:0" json:"voucher_discount"`
	CouponID        *uint   `json:"coupon_id"`
	CouponDiscount  float64 `gorm:"type:decimal(12,2);default:0" json:"coupon_discount"`
	PointsUsed      int     `gorm:"default:0" json:"points_used"`
	PointsDiscount  float64 `gorm:"type:decimal(12,2);default:0" json:"points_discount"`
	TotalPrice      float64 `gorm:"type:decimal(12,2);not null" json:"total_price"`

	Status TxStatus `gorm:"type:varchar(20);not null;index" json:"status"`

	MidtransOrderID     string    `gorm:"size:100;uniqueIndex" json:"midtrans_order_id"`
	MidtransToken       string    `gorm:"size:255" json:"-"`
	MidtransRedirectURL string    `gorm:"size:255" json:"midtrans_redirect_url"`
	PaymentDeadline     time.Time `json:"payment_deadline"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Review struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID uint      `gorm:"uniqueIndex;not null" json:"transaction_id"`
	EventID       uint      `gorm:"index;not null" json:"event_id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Rating        int       `gorm:"not null" json:"rating"`
	Comment       string    `gorm:"type:text" json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

// AllModels is used by AutoMigrate.
func AllModels() []any {
	return []any{
		&User{}, &Category{}, &Event{}, &TicketType{}, &Voucher{},
		&Coupon{}, &PointTransaction{}, &Transaction{}, &Review{},
	}
}
