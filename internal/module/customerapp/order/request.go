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
	Quantity      int64  `json:"quantity" validate:"eq=1"`
}

type GetManyOrderRequest struct {
	Page int64 `validate:"required"`
	Size int64 `validate:"required"`
}
