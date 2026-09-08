package optionhash_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/optionhash"
)

func TestHash_OrderInvariant(t *testing.T) {
	a := uuid.MustParse("66666666-6666-4666-8666-666666666301")
	b := uuid.MustParse("66666666-6666-4666-8666-666666666302")
	h1 := optionhash.Hash([]uuid.UUID{a, b})
	h2 := optionhash.Hash([]uuid.UUID{b, a})
	assert.Equal(t, h1, h2)
}

func TestHash_DifferentSets(t *testing.T) {
	a := uuid.MustParse("66666666-6666-4666-8666-666666666301")
	b := uuid.MustParse("66666666-6666-4666-8666-666666666302")
	c := uuid.MustParse("66666666-6666-4666-8666-666666666303")
	assert.NotEqual(t, optionhash.Hash([]uuid.UUID{a}), optionhash.Hash([]uuid.UUID{a, b}))
	assert.NotEqual(t, optionhash.Hash([]uuid.UUID{a, b}), optionhash.Hash([]uuid.UUID{a, c}))
}

func TestHash_DuplicatesIgnored(t *testing.T) {
	a := uuid.MustParse("66666666-6666-4666-8666-666666666301")
	b := uuid.MustParse("66666666-6666-4666-8666-666666666302")
	single := optionhash.Hash([]uuid.UUID{a, b})
	dup := optionhash.Hash([]uuid.UUID{a, b, a, b})
	assert.Equal(t, single, dup)
}

func TestHash_EmptySet(t *testing.T) {
	// MD5 of empty payload — deterministic empty-set hash
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", optionhash.Hash(nil))
	assert.Equal(t, optionhash.Hash(nil), optionhash.Hash([]uuid.UUID{}))
}
