package geo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/geo"
)

func TestHaversine_MoscowToSPB(t *testing.T) {
	// ~635 km between city centers
	d := geo.Haversine(55.7558, 37.6173, 59.9343, 30.3351)
	assert.InDelta(t, 635_000, d, 635_000*0.005)
}

func TestHaversine_SamePoint(t *testing.T) {
	assert.InDelta(t, 0, geo.Haversine(55.76, 37.62, 55.76, 37.62), 1)
}

func TestHaversine_InsideDeliveryRadius(t *testing.T) {
	venueLat, venueLon := 55.7558, 37.6173
	customerLat, customerLon := 55.7600, 37.6200
	d := geo.Haversine(venueLat, venueLon, customerLat, customerLon)
	assert.Less(t, d, 1000.0)
}

func TestHaversine_OutsideDeliveryRadius(t *testing.T) {
	d := geo.Haversine(55.7558, 37.6173, 55.90, 37.90)
	assert.Greater(t, d, 10_000.0)
}

func TestHaversine_OnBoundary(t *testing.T) {
	// ~1000 m north at equator-ish latitude approximation
	radius := 1000.0
	dLat := radius / 111320
	d := geo.Haversine(55.0, 37.0, 55.0+dLat, 37.0)
	assert.InDelta(t, radius, d, radius*0.01)
}

func TestComputeBoundingBox_Valid(t *testing.T) {
	box, ok := geo.ComputeBoundingBox(55.7558, 37.6173, 2000)
	require.True(t, ok)
	assert.Less(t, box.MinLat, 55.7558)
	assert.Greater(t, box.MaxLat, 55.7558)
	assert.Less(t, box.MinLon, 37.6173)
	assert.Greater(t, box.MaxLon, 37.6173)
}

func TestComputeBoundingBox_NearPole(t *testing.T) {
	_, ok := geo.ComputeBoundingBox(88.0, 0, 1000)
	assert.True(t, ok)

	_, ok = geo.ComputeBoundingBox(89.1, 0, 1000)
	assert.False(t, ok)
}

func TestComputeBoundingBox_ContainsPoint(t *testing.T) {
	lat, lon := 55.7558, 37.6173
	radius := 5000.0
	box, ok := geo.ComputeBoundingBox(lat, lon, radius)
	require.True(t, ok)

	nearLat := lat + 0.01
	nearLon := lon + 0.01
	assert.GreaterOrEqual(t, nearLat, box.MinLat)
	assert.LessOrEqual(t, nearLat, box.MaxLat)
	assert.GreaterOrEqual(t, nearLon, box.MinLon)
	assert.LessOrEqual(t, nearLon, box.MaxLon)

	d := geo.Haversine(lat, lon, nearLat, nearLon)
	if d <= radius {
		assert.True(t, nearLat >= box.MinLat && nearLat <= box.MaxLat)
	}
}
