package models

type Order struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Total   float64 `json:"total"`
}

type OrderStatus struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
