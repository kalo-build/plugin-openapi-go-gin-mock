package mock

import "time"

type Customer struct {
	CreatedAt time.Time              `json:"created_at"`
	Email     string                 `json:"email"`
	Id        string                 `json:"id"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Name      string                 `json:"name"`
}

type Invoice struct {
	ChargeAmount int       `json:"charge_amount"`
	CreatedAt    time.Time `json:"created_at"`
	Currency     string    `json:"currency"`
	CustomerId   string    `json:"customer_id"`
	Description  *string   `json:"description,omitempty"`
	Id           string    `json:"id"`
	Status       string    `json:"status"`
}

type Payment struct {
	ChargeAmount  int       `json:"charge_amount"`
	CreatedAt     time.Time `json:"created_at"`
	Currency      string    `json:"currency"`
	Id            string    `json:"id"`
	InvoiceId     string    `json:"invoice_id"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
	Status        string    `json:"status"`
}
