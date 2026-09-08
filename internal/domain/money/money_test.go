package money_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/money"
)

func TestLineTotal(t *testing.T) {
	assert.Equal(t, money.Minor(49000), money.LineTotal(49000, 1))
	assert.Equal(t, money.Minor(118000), money.LineTotal(59000, 2))
	assert.Equal(t, money.Minor(5841000), money.LineTotal(59000, 99))
}

func TestSum_WithModifiers(t *testing.T) {
	base := money.Minor(49000)
	opt := money.Minor(10000)
	total := money.Sum(base+opt, base+opt)
	assert.Equal(t, money.Minor(118000), total)
}

func TestLineTotal_ZeroPrice(t *testing.T) {
	assert.Equal(t, money.Minor(0), money.LineTotal(0, 5))
}

func TestSum_LargeValues(t *testing.T) {
	// price ≤ 10^7, qty ≤ 99 → safe int64
	line := money.LineTotal(10_000_000, 99)
	assert.Equal(t, money.Minor(990_000_000), line)
	total := money.Sum(line, line)
	assert.Equal(t, money.Minor(1_980_000_000), total)
}

func TestApplyPercent_HalfUp(t *testing.T) {
	// 10% of 105 → 10.5 → rounds to 11 (half-up)
	assert.Equal(t, money.Minor(11), money.ApplyPercent(105, 10))
	// 10% of 104 → 10.4 → rounds to 10
	assert.Equal(t, money.Minor(10), money.ApplyPercent(104, 10))
}

func TestApplyPercent_EdgeCases(t *testing.T) {
	assert.Equal(t, money.Minor(1000), money.ApplyPercent(1000, 0))
	assert.Equal(t, money.Minor(1000), money.ApplyPercent(1000, -5))
	assert.Equal(t, money.Minor(0), money.ApplyPercent(1000, 100))
	assert.Equal(t, money.Minor(0), money.ApplyPercent(1000, 150))
}

func TestApplyPercent_FiftyPercent(t *testing.T) {
	assert.Equal(t, money.Minor(250), money.ApplyPercent(500, 50))
	assert.Equal(t, money.Minor(251), money.ApplyPercent(501, 50)) // 250.5 → 251
}
