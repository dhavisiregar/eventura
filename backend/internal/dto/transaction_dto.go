package dto

type CreateTransactionRequest struct {
	EventID      uint   `json:"event_id" binding:"required"`
	TicketTypeID *uint  `json:"ticket_type_id"`
	Quantity     int    `json:"quantity" binding:"required,min=1,max=10"`
	VoucherCode  string `json:"voucher_code"`
	UseCoupon    bool   `json:"use_coupon"`
	PointsToUse  int    `json:"points_to_use" binding:"min=0"`
}
