package event

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/order"
	"github.com/tsel-ticketmaster/tm-order/internal/module/adminapp/ticket"
)

type EventUseCase interface {
	CreateEvent(ctx context.Context, req CreateEventRequest) (interface{}, error)
}

type eventUseCase struct {
	logger                       *logrus.Logger
	location                     *time.Location
	timeout                      time.Duration
	eventRepository              EventRepository
	artistRepository             ArtistRepository
	promotorRepository           PromotorRepository
	showRepository               ShowRepository
	locationRepository           LocationRepository
	orderRuleDayRepository       order.OrderRuleDayRepository
	orderRuleRangeDateRepository order.OrderRuleRangeDateRepository
	ticketStockRepository        ticket.TicketStockRepository
}

type EventUseCaseProperty struct {
	Logger                       *logrus.Logger
	Location                     *time.Location
	Timeout                      time.Duration
	EventRepository              EventRepository
	ArtistRepository             ArtistRepository
	PromotorRepository           PromotorRepository
	ShowRepository               ShowRepository
	LocationRepository           LocationRepository
	OrderRuleDayRepository       order.OrderRuleDayRepository
	OrderRuleRangeDateRepository order.OrderRuleRangeDateRepository
	TicketStockRepository        ticket.TicketStockRepository
}

func NewEventUseCase(props EventUseCaseProperty) EventUseCase {
	return &eventUseCase{
		logger:                       props.Logger,
		location:                     props.Location,
		timeout:                      props.Timeout,
		eventRepository:              props.EventRepository,
		artistRepository:             props.ArtistRepository,
		promotorRepository:           props.PromotorRepository,
		showRepository:               props.ShowRepository,
		locationRepository:           props.LocationRepository,
		orderRuleDayRepository:       props.OrderRuleDayRepository,
		orderRuleRangeDateRepository: props.OrderRuleRangeDateRepository,
		ticketStockRepository:        props.TicketStockRepository,
	}
}

func (u *eventUseCase) createArtists(ctx context.Context, e Event, tx *sql.Tx) error {
	for _, a := range e.Artists {
		if err := u.artistRepository.Save(ctx, a, tx); err != nil {
			return err
		}
	}

	return nil
}

func (u *eventUseCase) createPromotors(ctx context.Context, e Event, tx *sql.Tx) error {
	for _, p := range e.Promotors {
		if err := u.promotorRepository.Save(ctx, p, tx); err != nil {
			return err
		}
	}

	return nil
}

func (u *eventUseCase) createShows(ctx context.Context, e Event, tx *sql.Tx) error {
	for _, s := range e.Shows {
		if err := u.showRepository.Save(ctx, s, tx); err != nil {
			return err
		}

		if err := u.locationRepository.Save(ctx, *s.Location, tx); err != nil {
			return err
		}

		for _, ts := range s.TicketStock {
			if err := u.ticketStockRepository.Save(ctx, ts, tx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *eventUseCase) createRules(ctx context.Context, e Event, tx *sql.Tx) error {
	if err := u.orderRuleRangeDateRepository.Save(ctx, e.OrderRules.OrderRuleRangeDate, tx); err != nil {
		return err
	}

	for _, v := range e.OrderRules.OrderRuleDay {
		if err := u.orderRuleDayRepository.Save(ctx, v, tx); err != nil {
			return err
		}
	}

	return nil
}

// CreateEvent implements EventUseCase.
func (u *eventUseCase) CreateEvent(ctx context.Context, req CreateEventRequest) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	now := time.Now()
	e, err := req.ToEntityEvent(u.location, now)
	if err != nil {
		return nil, err
	}

	tx, err := u.eventRepository.BeginTx(ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := u.eventRepository.Save(ctx, e, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := u.createArtists(ctx, e, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := u.createPromotors(ctx, e, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := u.createShows(ctx, e, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := u.createRules(ctx, e, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	resp := CreateEventResponse{}
	resp.PopulateFromEntity(e)

	return resp, nil
}
