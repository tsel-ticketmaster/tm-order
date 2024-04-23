package order

import "time"

type Order struct {
	ID                      string
	PaymentMethod           string
	VirtualAccount          *string
	TransactionID           *string
	Status                  string
	CustomerID              int64
	CustomerName            string
	CustomerEmail           string
	TaxPercentage           float64
	ServiceChargePercentage float64
	DiscountPercentage      float64
	ServiceCharge           float64
	Tax                     float64
	Discount                float64
	Items                   []Item
	Subtotal                float64
	TotalAmount             float64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type Item struct {
	ID            int64
	OrderID       string
	TicketStockID string
	ShowID        string
	EventID       string
	EventName     string
	ShowVenue     string
	Tier          string
	Price         float64
	Quantity      int64
}

type OrderRuleRangeDate struct {
	EventID   string
	StartDate time.Time
	EndDate   time.Time
}

type OrderRuleDay struct {
	EventID string
	Day     int64
}
