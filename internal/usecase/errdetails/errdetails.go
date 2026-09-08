package errdetails

import (
	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
)

type Validation struct {
	Fields []FieldError
}

type FieldError struct {
	Field   string
	Message string
}

type UnavailableItems struct {
	ItemIDs []uuid.UUID
}

type CartVenueMismatch struct {
	CurrentVenueID uuid.UUID
}

type BelowMinOrder struct {
	MinOrderMinor   int64
	ItemsTotalMinor int64
}

type OutsideDeliveryZone struct {
	DistanceM int
	RadiusM   int
}

type PriceChanged struct {
	ExpectedTotalMinor int64
	ActualTotalMinor   int64
}

type InvalidTransition struct {
	CurrentStatus   orderstatus.Status
	RequestedStatus orderstatus.Status
}

type DuplicateExternalID struct {
	ExternalIDs []string
}

type VenueClosed struct {
	Reason string
}

func ValidationFields(fields ...FieldError) Validation {
	return Validation{Fields: fields}
}

func UnavailableItemIDs(ids ...uuid.UUID) UnavailableItems {
	return UnavailableItems{ItemIDs: ids}
}

func CartVenueMismatchDetail(venueID uuid.UUID) CartVenueMismatch {
	return CartVenueMismatch{CurrentVenueID: venueID}
}

func BelowMinOrderDetail(min, items int64) BelowMinOrder {
	return BelowMinOrder{MinOrderMinor: min, ItemsTotalMinor: items}
}

func OutsideDeliveryZoneDetail(distance, radius int) OutsideDeliveryZone {
	return OutsideDeliveryZone{DistanceM: distance, RadiusM: radius}
}

func PriceChangedDetail(expected, actual int64) PriceChanged {
	return PriceChanged{ExpectedTotalMinor: expected, ActualTotalMinor: actual}
}

func InvalidTransitionDetail(current, requested orderstatus.Status) InvalidTransition {
	return InvalidTransition{CurrentStatus: current, RequestedStatus: requested}
}

func DuplicateExternalIDs(ids ...string) DuplicateExternalID {
	return DuplicateExternalID{ExternalIDs: ids}
}

func VenueClosedDetail(reason string) VenueClosed {
	return VenueClosed{Reason: reason}
}
