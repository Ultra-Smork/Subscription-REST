package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" validate:"required"`
	Price       int       `json:"price"       validate:"required,min=0"`
	UserID      uuid.UUID `json:"user_id"     validate:"required"`
	StartDate   string    `json:"start_date"  validate:"required,monthyear"` // формат MM-YYYY
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" validate:"omitempty"`
	Price       *int    `json:"price,omitempty"        validate:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty"   validate:"omitempty,monthyear"`
}

func (r *CreateSubscriptionRequest) Validate() error {
	if r.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if r.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user_id must be a valid UUID")
	}
	if _, err := time.Parse("01-2006", r.StartDate); err != nil {
		return fmt.Errorf("start_date must be in MM-YYYY format")
	}
	return nil
}

func (r *UpdateSubscriptionRequest) Validate() error {
	if r.Price != nil && *r.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if r.StartDate != nil && *r.StartDate != "" {
		if _, err := time.Parse("01-2006", *r.StartDate); err != nil {
			return fmt.Errorf("start_date must be in MM-YYYY format")
		}
	}
	return nil
}
