package dto

import (
	"github.com/Ultra-Smork/Subscription-service/internals/model"
)

func ToSubscriptionResponse(sub *model.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"),
	}
}

func ToSubscriptionResponseList(subs []*model.Subscription) []SubscriptionResponse {
	result := make([]SubscriptionResponse, len(subs))
	for i, sub := range subs {
		result[i] = ToSubscriptionResponse(sub)
	}
	return result
}
