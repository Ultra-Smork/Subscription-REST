package dto

import (
	"github.com/google/uuid"
)

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"` // в формате MM-YYYY
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
