package event

import (
	"time"

	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/order"
	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/ticket"
)

const (
	VenueOnline            string = "ONLINE"
	ShowTypeLive           string = "LIVE"
	ShowTypeHologramLive   string = "HOLOGRAM_LIVE"
	ShowTypeOnline         string = "ONLINE"
	TicketTierOnline       string = "ONLINE"
	TicketTierWood         string = "WOOD"
	TicketTierBronze       string = "BRONZE"
	TicketTierSilver       string = "SILVER"
	TicketTierGold         string = "GOLD"
	TypeOrderRuleRangeDate string = "ORDER_RULE_RANGE_DATE"
)

type Location struct {
	EventID          string
	ShowID           string
	Country          string
	City             string
	FormattedAddress string
	Latitude         float64
	Longitude        float64
}

type Show struct {
	EventID     string
	ID          string
	Venue       string
	Type        string
	TicketStock []ticket.TicketStock
	Location    *Location
	Time        time.Time
	Status      string
}

type Promotor struct {
	EventID string
	Name    string
	Email   string
	Phone   string
}

type Artist struct {
	EventID string
	Name    string
}

type Event struct {
	ID          string
	Name        string
	Promotors   []Promotor
	Artists     []Artist
	Shows       []Show
	Description string
	Status      string
	OrderRules  OrderRuleAggregation
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderRuleAggregation struct {
	OrderRuleRangeDate order.OrderRuleRangeDate
	OrderRuleDay       []order.OrderRuleDay
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

type OrderRuleMaximumTicket struct {
	EventID string
	Maximum int64
}
