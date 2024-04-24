package order

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/event"
	"github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/midtrans"
	"github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/ticket"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/session"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/util"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/gctasks"
	"github.com/tsel-ticketmaster/tm-order/pkg/pubsub"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type OrderUseCase interface {
	PlaceOrder(ctx context.Context, req PlaceOrderRequest) (PlaceOrderResponse, error)
	OnPaymentNotification(ctx context.Context, e PaymentNotificationEvent) error
	OnExpireOrder(ctx context.Context, e ExpireOrderEvent) error
	GetManyOrder(ctx context.Context) (GetManyOrderResponse, error)
}

type orderUseCase struct {
	logger                       *logrus.Logger
	timeout                      time.Duration
	baseURL                      string
	orderExpireDuration          time.Duration
	serviceChargePercentage      float64
	taxPercentage                float64
	eventRepository              event.EventRepository
	showRepository               event.ShowRepository
	ticketStockRepository        ticket.TicketStockRepository
	orderRuleRangeDateRepository OrderRuleRangeDateRepository
	orderRuleDayRepository       OrderRuleDayRepository
	orderRepository              OrderRepository
	itemRepository               ItemRepository
	publisher                    pubsub.Publisher
	midtransRepository           midtrans.MidtransRepository
	cloudTask                    gctasks.Client
}

type OrderUseCaseProperty struct {
	Logger                       *logrus.Logger
	Timeout                      time.Duration
	BaseURL                      string
	OrderExpireDuration          time.Duration
	ServiceChargePercentage      float64
	TaxPercentage                float64
	EventRepository              event.EventRepository
	ShowRepository               event.ShowRepository
	TicketStockRepository        ticket.TicketStockRepository
	OrderRuleRangeDateRepository OrderRuleRangeDateRepository
	OrderRuleDay                 OrderRuleDayRepository
	OrderRepository              OrderRepository
	ItemRepository               ItemRepository
	Publisher                    pubsub.Publisher
	MidtransRepository           midtrans.MidtransRepository
	CloudTask                    gctasks.Client
}

func NewOrderUseCase(props OrderUseCaseProperty) OrderUseCase {
	return &orderUseCase{
		logger:                       props.Logger,
		timeout:                      props.Timeout,
		baseURL:                      props.BaseURL,
		orderExpireDuration:          props.OrderExpireDuration,
		serviceChargePercentage:      props.ServiceChargePercentage,
		taxPercentage:                props.TaxPercentage,
		eventRepository:              props.EventRepository,
		showRepository:               props.ShowRepository,
		ticketStockRepository:        props.TicketStockRepository,
		orderRuleRangeDateRepository: props.OrderRuleRangeDateRepository,
		orderRuleDayRepository:       props.OrderRuleDay,
		orderRepository:              props.OrderRepository,
		itemRepository:               props.ItemRepository,
		publisher:                    props.Publisher,
		midtransRepository:           props.MidtransRepository,
		cloudTask:                    props.CloudTask,
	}
}

// GetManyOrder implements OrderUseCase.
func (u *orderUseCase) GetManyOrder(ctx context.Context) (GetManyOrderResponse, error) {
	panic("unimplemented")
}

// OnPaymentNotification implements OrderUseCase.
func (u *orderUseCase) OnPaymentNotification(ctx context.Context, e PaymentNotificationEvent) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	if e.TransactionStatus != "settlement" {
		return nil
	}

	tx, err := u.orderRepository.BeginTx(ctx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	order, err := u.orderRepository.FindByID(ctx, e.OrderID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	if order.Status != "WAITING_FOR_PAYMENT" {
		u.orderRepository.Rollback(ctx, tx)
		return nil
	}

	items, err := u.itemRepository.FindManyByOrderID(ctx, e.OrderID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}
	order.Items = items
	order.Status = "PAID"

	if err := u.orderRepository.Update(ctx, order.ID, order, tx); err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	if err := u.orderRepository.CommitTx(ctx, tx); err != nil {
		return err
	}

	orderBuff, _ := json.Marshal(order)
	u.publisher.Publish(ctx, "order-paid", *order.TransactionID, nil, orderBuff)

	return nil
}

// OnOrderExpired implements OrderUseCase.
func (u *orderUseCase) OnExpireOrder(ctx context.Context, e ExpireOrderEvent) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	tx, err := u.orderRepository.BeginTx(ctx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	order, err := u.orderRepository.FindByID(ctx, e.ID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	if order.Status == "PAID" {
		u.orderRepository.Rollback(ctx, tx)
		return nil
	}

	now := time.Now()
	order.Status = "EXPIRED"
	order.UpdatedAt = now

	if err := u.orderRepository.Update(ctx, order.ID, order, tx); err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return err
	}

	if err := u.orderRepository.CommitTx(ctx, tx); err != nil {
		return err
	}

	return nil
}

// PlaceOrder implements OrderUseCase.
func (u *orderUseCase) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (PlaceOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	acc, err := session.GetAccountFromCtx(ctx)
	if err != nil {
		return PlaceOrderResponse{}, err
	}

	tx, err := u.orderRepository.BeginTx(ctx)
	if err != nil {
		return PlaceOrderResponse{}, err
	}

	now := time.Now()
	order := Order{
		ID:                      util.GenerateTimestampWithPrefix("TO"),
		PaymentMethod:           req.PaymentMethod,
		VirtualAccount:          nil,
		Status:                  "WAITING_FOR_PAYMENT",
		CustomerID:              acc.ID,
		CustomerName:            acc.Name,
		CustomerEmail:           acc.Email,
		TaxPercentage:           u.taxPercentage,
		ServiceChargePercentage: u.serviceChargePercentage,
		DiscountPercentage:      0,
		ServiceCharge:           0,
		Tax:                     0,
		Discount:                0,
		Items:                   nil,
		Subtotal:                0,
		TotalAmount:             0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	var subtotal float64

	e, err := u.eventRepository.FindByID(ctx, req.EventID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	s, err := u.showRepository.FindByID(ctx, req.ShowID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	if e.ID != s.EventID {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, errors.New(http.StatusBadRequest, status.BAD_REQUEST, "invalid show id")
	}

	ts, err := u.ticketStockRepository.FindByIDForUpdate(ctx, req.TicketStockID, tx)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	if s.ID != ts.ShowID {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, errors.New(http.StatusBadRequest, status.BAD_REQUEST, "invalid ticket stock id")
	}

	stock := ts.Acquired + req.Quantity
	if stock > ts.Allocation {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, errors.New(http.StatusBadRequest, status.BAD_REQUEST, "out of stock")
	}

	subtotal = subtotal + (ts.Price * float64(req.Quantity))

	item := Item{
		OrderID:       order.ID,
		TicketStockID: ts.ID,
		ShowID:        s.ID,
		EventID:       e.ID,
		EventName:     e.Name,
		ShowVenue:     s.Venue,
		Tier:          ts.Tier,
		Price:         ts.Price,
		Quantity:      req.Quantity,
	}

	order.Items = []Item{item}

	serviceCharge := subtotal * u.serviceChargePercentage / 100
	fmt.Println(serviceCharge)
	tax := subtotal * u.taxPercentage / 100

	totalAmount := subtotal + serviceCharge + tax

	order.Tax = tax
	order.ServiceCharge = serviceCharge
	order.Subtotal = subtotal
	order.TotalAmount = math.Round(totalAmount)

	chargePayment := midtrans.ChargeRequest{
		PaymentType: midtrans.BankTransferType,
		BankTransfer: midtrans.BankTransfer{
			Bank: req.PaymentMethod,
		},
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:     order.ID,
			GrossAmount: int64(order.TotalAmount),
		},
	}

	chargeResponse, err := u.midtransRepository.Charge(ctx, chargePayment)
	if err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	order.TransactionID = &chargeResponse.TransactionID
	order.VirtualAccount = &chargeResponse.VaNumbers[0].VaNumber

	if err := u.orderRepository.Save(ctx, order, tx); err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	if err := u.itemRepository.Save(ctx, order.Items[0], tx); err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	if err := u.orderRepository.CommitTx(ctx, tx); err != nil {
		u.orderRepository.Rollback(ctx, tx)
		return PlaceOrderResponse{}, err
	}

	orderBuff, _ := json.Marshal(order)

	orderExpiredAt := now.Add(u.orderExpireDuration)
	tasksRequest := gctasks.Request{
		URL:    fmt.Sprintf("%s/v1/customerapp/orders/on-expire", u.baseURL),
		Method: cloudtaskspb.HttpMethod_POST,
		Body:   orderBuff,
	}
	u.cloudTask.DeferCreateTaskInTime("expire-order", tasksRequest, orderExpiredAt)

	resp := PlaceOrderResponse{}
	resp.PopulateFromEntity(order)

	return resp, nil
}
