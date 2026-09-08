package list

import (
	"context"
	"math"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/geo"
	nowercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/nower"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type UseCase struct {
	venues venuerepo.Repository
	nower  nowercontract.Nower
	defTZ  *time.Location
}

func New(venues venuerepo.Repository, nower nowercontract.Nower, defTZ *time.Location) *UseCase {
	return &UseCase{venues: venues, nower: nower, defTZ: defTZ}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	active := "active"
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	venues, err := uc.venues.List(ctx, venuerepo.ListFilter{
		CityID: in.CityID, Status: &active, Query: in.Search,
		Lat: in.Lat, Lon: in.Lon, Sort: in.Sort, Limit: limit + 1,
		CursorTS: in.CursorTS, CursorID: in.CursorID,
	})
	if err != nil {
		return nil, err
	}

	hasNext := len(venues) > limit
	if hasNext {
		venues = venues[:limit]
	}

	now := uc.nower.Now()
	out := &Out{Items: make([]VenueOut, 0, len(venues))}
	for _, v := range venues {
		hours, _ := uc.venues.GetWorkingHours(ctx, v.ID)
		isOpen, _ := shared.VenueIsOpen(&v, hours, now, uc.defTZ)
		if in.OnlyOpen && !isOpen {
			continue
		}
		eta := v.AvgPrepMinutes + 10
		var distance *int
		if in.Lat != nil && in.Lon != nil {
			d := int(math.Round(geo.Haversine(v.Lat, v.Lon, *in.Lat, *in.Lon)))
			distance = &d
		}
		out.Items = append(out.Items, VenueOut{
			Venue: v, IsOpen: isOpen, ETAMinutes: &eta, DistanceM: distance,
		})
	}

	if len(out.Items) > 0 {
		last := out.Items[len(out.Items)-1].Venue
		ts := last.CreatedAt
		id := last.ID
		out.LastTS = &ts
		out.LastID = &id
	}
	out.HasNext = hasNext
	return out, nil
}
