package midtrans

const (
	BankTransferType = "bank_transfer"
	BCA              = "bca"
	BNI              = "bni"
	BRI              = "bri"
)

type BankTransfer struct {
	Bank string `json:"bank"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type ChargeRequest struct {
	PaymentType        string             `json:"payment_type"`
	BankTransfer       BankTransfer       `json:"bank_transfer"`
	TransactionDetails TransactionDetails `json:"transaction_details"`
}

type VANumber struct {
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`
}

type ChargeResponse struct {
	StatusCode        string     `json:"status_code"`
	StatusMessage     string     `json:"status_message"`
	TransactionID     string     `json:"transaction_id"`
	OrderID           string     `json:"order_id"`
	MerchantID        string     `json:"merchant_id"`
	GrossAmount       string     `json:"gross_amount"`
	Currency          string     `json:"currency"`
	PaymentType       string     `json:"payment_type"`
	SignatureKey      string     `json:"signature_key"`
	TransactionTime   string     `json:"transaction_time"`
	TransactionStatus string     `json:"transaction_status"`
	FraudStatus       string     `json:"fraud_status"`
	PermataVaNumber   string     `json:"permata_va_number"`
	VaNumbers         []VANumber `json:"va_numbers"`
	ExpiryTime        string     `json:"expiry_time"`
}
