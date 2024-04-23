package order

import "time"

type OrderRuleRangeDate struct {
	EventID   string
	StartDate time.Time
	EndDate   time.Time
}

type OrderRuleDay struct {
	EventID string
	Day     int64
}
