package event

import (
	"math"
	"time"

	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/order"
	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/ticket"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/util"
)

type CreateLocationRequest struct {
	Country          string  `json:"country" validate:"required"`
	City             string  `json:"city" validate:"required"`
	FormattedAddress string  `json:"formatted_address" validate:"formatted_address"`
	Latitude         float64 `json:"latitude" validate:"required"`
	Longitude        float64 `json:"longitude" validate:"required"`
}

type CreateTicketAllocation struct {
	Tier                   string  `json:"tier" validate:"oneof=WOOD BRONZE SILVER GOLD"`
	AllocationByPercentage float64 `json:"allocation_by_percentage" validate:"required"`
	Price                  float64 `json:"price" validate:"required"`
}

type CreateShowRequest struct {
	Venue                 string                   `json:"venue" validate:"required"`
	Type                  string                   `json:"type" validate:"oneof=LIVE HOLOGRAM_LIVE"`
	Online                bool                     `json:"online" validate:"-"`
	Location              *CreateLocationRequest   `json:"location" validate:"-"`
	TotalTicketAllocation int64                    `json:"total_ticket_allocation"`
	TicketAllocation      []CreateTicketAllocation `json:"ticket_allocation" validate:"required,dive"`
}

type CreateEventRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Artists     []string `json:"artists" validate:"required,dive,required"`
	Promotors   []struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"email"`
		Phone string `json:"phone" validate:"required"`
	} `json:"promotors" validate:"required,dive"`
	OnlineTicketPrice           float64             `json:"online_ticket_price" validate:"required"`
	TotalOnlineTicketAllocation int64               `json:"total_online_ticket_allocation" validate:"required"`
	Shows                       []CreateShowRequest `json:"shows" validate:"required,dive,required"`
	ShowTime                    string              `json:"show_time" validate:"datetime=2006-01-02 15:04:05"`
	OrderRuleDay                []int64             `json:"order_rule_day" validate:"omitempty,dive,min=1,max=7"`
	OrderRuleRangeDate          struct {
		StartDate string `json:"start_date" validate:"datetime=2006-01-02 15:04:05"`
		EndDate   string `json:"end_date" validate:"datetime=2006-01-02 15:04:05"`
	} `json:"order_rule_range_date" validate:"required"`
	OrderRuleMaximumTicket int64 `json:"order_rule_maximum_ticket" validate:"-"`
}

func (r CreateEventRequest) ToEntityEvent(location *time.Location, now time.Time) (Event, error) {
	event := Event{
		ID:          util.GenerateTimestampWithPrefix("EVENT"),
		Name:        r.Name,
		Promotors:   nil,
		Artists:     nil,
		Shows:       nil,
		Description: r.Description,
		Status:      "ACTIVE",
		OrderRules:  OrderRuleAggregation{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	promotors := make([]Promotor, len(r.Promotors))
	for k, v := range r.Promotors {
		promotors[k] = Promotor{
			EventID: event.ID,
			Name:    v.Name,
			Email:   v.Email,
			Phone:   v.Phone,
		}
	}
	event.Promotors = promotors

	artists := make([]Artist, len(r.Artists))
	for k, v := range r.Artists {
		artists[k] = Artist{
			EventID: event.ID,
			Name:    v,
		}
	}
	event.Artists = artists

	showTime, _ := time.ParseInLocation(time.DateTime, r.ShowTime, location)

	shows := make([]Show, 0)
	for _, v := range r.Shows {
		liveShow := Show{
			EventID: event.ID,
			ID:      util.GenerateTimestampWithPrefix("SHOW"),
			Venue:   v.Venue,
			Type:    v.Type,
			Time:    showTime,
			Status:  "ACTIVE",
		}
		liveShow.Location = &Location{
			EventID:          event.ID,
			ShowID:           liveShow.ID,
			Country:          v.Location.Country,
			City:             v.Location.City,
			FormattedAddress: v.Location.FormattedAddress,
			Latitude:         v.Location.Latitude,
			Longitude:        v.Location.Longitude,
		}

		liveShowTicketStock := make([]ticket.TicketStock, len(v.TicketAllocation))
		for tark, tarv := range v.TicketAllocation {

			allocation := int64(math.Round(tarv.AllocationByPercentage / 100 * float64(v.TotalTicketAllocation)))

			liveShowTicketStock[tark] = ticket.TicketStock{
				EventID:         event.ID,
				ShowID:          liveShow.ID,
				ID:              util.GenerateTimestampWithPrefix("TSTK"),
				OnlineFor:       nil,
				Tier:            tarv.Tier,
				Allocation:      allocation,
				Price:           tarv.Price,
				Acquired:        0,
				LastStockUpdate: now,
			}
		}
		liveShow.TicketStock = liveShowTicketStock

		shows = append(shows, liveShow)

		if v.Online {
			onlineShow := Show{
				EventID: event.ID,
				ID:      util.GenerateTimestampWithPrefix("SHOW"),
				Venue:   VenueOnline,
				Type:    ShowTypeOnline,
				Time:    showTime,
				Status:  "ACTIVE",
			}
			onlineShow.Location = &Location{
				EventID:          event.ID,
				ShowID:           onlineShow.ID,
				Country:          v.Location.Country,
				City:             v.Location.City,
				FormattedAddress: v.Location.FormattedAddress,
				Latitude:         v.Location.Latitude,
				Longitude:        v.Location.Longitude,
			}

			defaultOnlineAllocationPercentage := float64(100) / float64(len(r.Shows))
			allocation := int64(math.Round(defaultOnlineAllocationPercentage / 100 * float64(r.TotalOnlineTicketAllocation)))
			onlineShow.TicketStock = append(onlineShow.TicketStock, ticket.TicketStock{
				EventID:         event.ID,
				ShowID:          onlineShow.ID,
				ID:              util.GenerateTimestampWithPrefix("TSTK"),
				OnlineFor:       &liveShow.ID,
				Tier:            TicketTierBronze,
				Allocation:      allocation,
				Price:           r.OnlineTicketPrice,
				Acquired:        0,
				LastStockUpdate: now,
			})

			shows = append(shows, onlineShow)
		}
	}

	event.Shows = shows

	ruleStartDate, _ := time.ParseInLocation(time.DateTime, r.OrderRuleRangeDate.StartDate, location)
	ruleEndDate, _ := time.ParseInLocation(time.DateTime, r.OrderRuleRangeDate.EndDate, location)

	orderRuleDay := make([]order.OrderRuleDay, len(r.OrderRuleDay))
	for k, v := range r.OrderRuleDay {
		orderRuleDay[k] = order.OrderRuleDay{
			EventID: event.ID,
			Day:     v,
		}
	}
	event.OrderRules = OrderRuleAggregation{
		OrderRuleRangeDate: order.OrderRuleRangeDate{
			EventID:   event.ID,
			StartDate: ruleStartDate,
			EndDate:   ruleEndDate,
		},
		OrderRuleDay: orderRuleDay,
	}

	return event, nil
}
