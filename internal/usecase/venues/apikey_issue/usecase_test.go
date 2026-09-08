package apikey_issue_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	apikeyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/apikey"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	venueapikeyissue "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/apikey_issue"
)

func TestRun_VenueNotFound(t *testing.T) {
	uc := venueapikeyissue.New(
		&uctest.VenueRepo{GetByIDErr: pgerr.ErrNotFound},
		(*apikeyrepo.Repository)(nil),
	)

	_, err := uc.Run(context.Background(), venueapikeyissue.In{VenueID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}
