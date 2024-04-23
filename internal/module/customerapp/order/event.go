package order

type ExpireOrderEvent struct {
	ID            string
	TransactionID string
}

type PaymentNotificationEvent struct {
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	OrderID           string `json:"order_id"`
}
