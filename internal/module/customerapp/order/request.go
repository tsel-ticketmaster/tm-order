package order

type ItemRequest struct {
	EventID       string `json:"event_id" validate:"required"`
	ShowID        string `json:"show_id" validate:"required"`
	TicketStockID string `json:"ticket_stock_id" validate:"required"`
	Quantity      int64  `json:"quantity" validate:"required"`
}
type PlaceOrderRequest struct {
	PaymentMethod string `json:"payment_method" validate:"oneof=bca bri bni"`
	EventID       string `json:"event_id" validate:"required"`
	ShowID        string `json:"show_id" validate:"required"`
	TicketStockID string `json:"ticket_stock_id" validate:"required"`
	Quantity      int64  `json:"quantity" validate:"required"`
}
