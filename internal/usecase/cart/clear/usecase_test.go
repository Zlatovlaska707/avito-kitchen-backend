package clear_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/clear"
)

type cartMock struct {
	cart         *cartrepo.Cart
	getActiveErr error
	clearErr     error
	abandonErr   error
	clearedID    uuid.UUID
	abandonedID  uuid.UUID
}

func (m *cartMock) GetActiveForUpdate(context.Context, uuid.UUID) (*cartrepo.Cart, error) {
	return nil, nil
}

func (m *cartMock) GetActive(_ context.Context, _ uuid.UUID) (*cartrepo.Cart, error) {
	if m.getActiveErr != nil {
		return nil, m.getActiveErr
	}
	return m.cart, nil
}

func (m *cartMock) Create(context.Context, cartrepo.CreateIn) (*cartrepo.Cart, error) {
	return nil, nil
}

func (m *cartMock) GetItems(context.Context, uuid.UUID) ([]cartrepo.Item, error) {
	return nil, nil
}

func (m *cartMock) AddItem(context.Context, cartrepo.AddItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartMock) UpdateItem(context.Context, cartrepo.UpdateItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartMock) DeleteItem(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (m *cartMock) ClearItems(_ context.Context, cartID uuid.UUID) error {
	m.clearedID = cartID
	return m.clearErr
}

func (m *cartMock) Abandon(_ context.Context, cartID uuid.UUID) error {
	m.abandonedID = cartID
	return m.abandonErr
}

func (m *cartMock) MarkConverted(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (m *cartMock) CountItems(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (m *cartMock) AbandonExpired(context.Context) (int, error) {
	return 0, nil
}

func TestRun_HappyPath(t *testing.T) {
	cartID := uuid.New()
	customerID := uuid.New()
	mock := &cartMock{cart: &cartrepo.Cart{ID: cartID, CustomerID: customerID}}
	uc := clear.New(mock)

	out, err := uc.Run(context.Background(), clear.In{CustomerID: customerID})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, cartID, mock.clearedID)
	assert.Equal(t, cartID, mock.abandonedID)
}

func TestRun_NoActiveCart(t *testing.T) {
	mock := &cartMock{getActiveErr: pgerr.ErrNotFound}
	uc := clear.New(mock)

	out, err := uc.Run(context.Background(), clear.In{CustomerID: uuid.New()})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, uuid.Nil, mock.clearedID)
}

func TestRun_GetActiveError(t *testing.T) {
	mock := &cartMock{getActiveErr: assert.AnError}
	uc := clear.New(mock)

	_, err := uc.Run(context.Background(), clear.In{CustomerID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_ClearItemsError(t *testing.T) {
	mock := &cartMock{
		cart:     &cartrepo.Cart{ID: uuid.New()},
		clearErr: assert.AnError,
	}
	uc := clear.New(mock)

	_, err := uc.Run(context.Background(), clear.In{CustomerID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_AbandonError(t *testing.T) {
	mock := &cartMock{
		cart:       &cartrepo.Cart{ID: uuid.New()},
		abandonErr: assert.AnError,
	}
	uc := clear.New(mock)

	_, err := uc.Run(context.Background(), clear.In{CustomerID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}
