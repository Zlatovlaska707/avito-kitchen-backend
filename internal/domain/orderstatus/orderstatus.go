package orderstatus

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPlaced     Status = "placed"
	StatusAccepted   Status = "accepted"
	StatusReady      Status = "ready"
	StatusInDelivery Status = "in_delivery"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
	StatusRejected   Status = "rejected"
)

type ActorType string

const (
	ActorCustomer ActorType = "customer"
	ActorVenue    ActorType = "venue"
	ActorCourier  ActorType = "courier"
	ActorSystem   ActorType = "system"
	ActorAdmin    ActorType = "admin"
)

type FulfillmentType string

const (
	FulfillmentDelivery FulfillmentType = "delivery"
	FulfillmentPickup   FulfillmentType = "pickup"
)

var (
	ErrInvalidTransition = errors.New("invalid transition")
	ErrNotCancellable    = errors.New("not cancellable")
)

type TransitionContext struct {
	Now              time.Time
	AcceptedAt       *time.Time
	CancelGrace      time.Duration
	AcceptTimeout    time.Duration
	PlacedAt         time.Time
	FulfillmentType  FulfillmentType
	Reason           string
	SystemTimeoutHit bool
}

func CanTransition(from, to Status, actor ActorType, tc TransitionContext) error {
	if isTerminal(from) {
		return ErrInvalidTransition
	}

	switch from {
	case StatusPlaced:
		switch to {
		case StatusAccepted:
			if actor == ActorVenue {
				return nil
			}
		case StatusRejected:
			if actor == ActorVenue && tc.Reason != "" {
				return nil
			}
		case StatusCancelled:
			switch actor {
			case ActorCustomer, ActorAdmin:
				return nil
			case ActorSystem:
				if tc.SystemTimeoutHit || tc.Now.Sub(tc.PlacedAt) >= tc.AcceptTimeout {
					return nil
				}
			}
		}
	case StatusAccepted:
		switch to {
		case StatusReady:
			if actor == ActorVenue {
				return nil
			}
		case StatusCancelled:
			switch actor {
			case ActorAdmin:
				return nil
			case ActorCustomer:
				if tc.AcceptedAt == nil {
					return ErrNotCancellable
				}
				if tc.Now.Sub(*tc.AcceptedAt) <= tc.CancelGrace {
					return nil
				}
				return ErrNotCancellable
			}
		}
	case StatusReady:
		switch to {
		case StatusInDelivery:
			if (actor == ActorVenue || actor == ActorSystem) && tc.FulfillmentType == FulfillmentDelivery {
				return nil
			}
		case StatusCompleted:
			if actor == ActorVenue && tc.FulfillmentType == FulfillmentPickup {
				return nil
			}
		}
	case StatusInDelivery:
		switch to {
		case StatusCompleted:
			if actor == ActorSystem || actor == ActorCourier {
				return nil
			}
		case StatusCancelled:
			if actor == ActorAdmin {
				return nil
			}
		}
	}

	return ErrInvalidTransition
}

func isTerminal(s Status) bool {
	switch s {
	case StatusCompleted, StatusCancelled, StatusRejected:
		return true
	default:
		return false
	}
}
