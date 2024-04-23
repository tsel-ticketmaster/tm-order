package order

import "time"

type GetManyOrderResponse []PlaceOrderResponse

type GetByOrderIDResponse PlaceOrderResponse

type PlaceOrderResponse struct {
	ID                      string         `json:"id"`
	PaymentMethod           string         `json:"payment_method"`
	TransactionID           *string        `json:"transaction_id"`
	VirtualAccount          *string        `json:"virtual_account"`
	Status                  string         `json:"status"`
	CustomerID              int64          `json:"customer_id"`
	CustomerName            string         `json:"customer_name"`
	CustomerEmail           string         `json:"customer_email"`
	TaxPercentage           float64        `json:"tax_percentage"`
	ServiceChargePercentage float64        `json:"service_charge_percentage"`
	DiscountPercentage      float64        `json:"discount_percentage"`
	ServiceCharge           float64        `json:"service_charge"`
	Tax                     float64        `json:"tax"`
	Discount                float64        `json:"discount"`
	Items                   []ItemResponse `json:"items"`
	Subtotal                float64        `json:"subtotal"`
	TotalAmount             float64        `json:"total_amount"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

func (r *PlaceOrderResponse) PopulateFromEntity(o Order) {
	r.ID = o.ID
	r.PaymentMethod = o.PaymentMethod
	r.VirtualAccount = o.VirtualAccount
	r.TransactionID = o.TransactionID
	r.Status = o.Status
	r.CustomerID = o.CustomerID
	r.CustomerName = o.CustomerName
	r.CustomerEmail = o.CustomerEmail
	r.TaxPercentage = o.TaxPercentage
	r.ServiceChargePercentage = o.ServiceChargePercentage
	r.Tax = o.Tax
	r.ServiceCharge = o.ServiceCharge
	r.Discount = o.Discount
	r.Subtotal = o.Subtotal
	r.TotalAmount = o.TotalAmount
	r.CreatedAt = o.CreatedAt
	r.UpdatedAt = o.UpdatedAt

	itemsResponse := make([]ItemResponse, len(o.Items))
	for k, v := range o.Items {
		itemsResponse[k] = ItemResponse{
			OrderID:       v.OrderID,
			TicketStockID: v.TicketStockID,
			ShowID:        v.ShowID,
			EventID:       v.EventID,
			EventName:     v.EventName,
			ShowVenue:     v.ShowVenue,
			Tier:          v.Tier,
			Price:         v.Price,
			Quantity:      v.Quantity,
		}
	}
	r.Items = itemsResponse
}

type ItemResponse struct {
	OrderID       string  `json:"order_id"`
	TicketStockID string  `json:"ticket_stock_id"`
	ShowID        string  `json:"show_id"`
	EventID       string  `json:"event_id"`
	EventName     string  `json:"event_name"`
	ShowVenue     string  `json:"show_venue"`
	Tier          string  `json:"tier"`
	Price         float64 `json:"price"`
	Quantity      int64   `json:"quantity"`
}
