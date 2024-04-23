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
