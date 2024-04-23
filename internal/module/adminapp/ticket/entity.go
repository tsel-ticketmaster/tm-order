package ticket

import "time"

type TicketStock struct {
	EventID         string
	ShowID          string
	ID              string
	OnlineFor       *string
	Tier            string
	Allocation      int64
	Price           float64
	Acquired        int64
	LastStockUpdate time.Time
}

type TicketStockJournal struct {
	TicketStockID string
	ID            int
	Action        string
	Stock         int64
	Description   string
	CreatedAt     time.Time
}

type AcquiredTicket struct {
	EventID       string
	ShowID        string
	TicketStockID string
	Number        string
	CustomerID    int64
	CustomerEmail string
	CustomerName  string
	ShowTime      time.Time
	OrderID       string
}
