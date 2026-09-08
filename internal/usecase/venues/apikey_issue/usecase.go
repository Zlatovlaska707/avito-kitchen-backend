package apikey_issue

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	apikeyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/apikey"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type UseCase struct {
	venues  venuerepo.Repository
	apikeys *apikeyrepo.Repository
}

func New(venues venuerepo.Repository, apikeys *apikeyrepo.Repository) *UseCase {
	return &UseCase{venues: venues, apikeys: apikeys}
}

func (uc *UseCase) Run(ctx context.Context, in In) (*Out, error) {
	if _, err := uc.venues.GetByID(ctx, in.VenueID); err != nil {
		if pgerr.IsNotFound(err) {
			return nil, usecase.ErrVenueNotFound
		}
		return nil, err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	prefixBytes := make([]byte, 4)
	if _, err := rand.Read(prefixBytes); err != nil {
		return nil, fmt.Errorf("generate prefix: %w", err)
	}
	prefix := hex.EncodeToString(prefixBytes)[:8]
	secretB64 := base64.RawURLEncoding.EncodeToString(secret)
	raw := fmt.Sprintf("ak_%s_%s", prefix, secretB64)
	hash := sha256.Sum256([]byte(raw))
	scopes := in.Scopes
	if len(scopes) == 0 {
		scopes = []string{"menu:write", "orders:write"}
	}
	rec, err := uc.apikeys.Issue(ctx, in.VenueID, prefix, hex.EncodeToString(hash[:]), scopes)
	if err != nil {
		return nil, err
	}
	return &Out{KeyID: rec.ID, APIKey: raw, Prefix: prefix}, nil
}
