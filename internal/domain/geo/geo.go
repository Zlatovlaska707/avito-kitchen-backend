package geo

import "math"

const earthRadiusM = 6371000

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	φ1 := lat1 * rad
	φ2 := lat2 * rad
	dφ := (lat2 - lat1) * rad
	dλ := (lon2 - lon1) * rad

	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// ComputeBoundingBox returns approximate lat/lon bounds for radiusM meters around center.
// When |lat| > 89 the box is not applied (returns zero-value with Valid=false).
func ComputeBoundingBox(lat, lon float64, radiusM float64) (BoundingBox, bool) {
	if math.Abs(lat) > 89 {
		return BoundingBox{}, false
	}
	dLat := radiusM / 111320
	cosLat := math.Cos(lat * math.Pi / 180)
	if cosLat < 1e-6 {
		return BoundingBox{}, false
	}
	dLon := radiusM / (111320 * cosLat)
	return BoundingBox{
		MinLat: lat - dLat,
		MaxLat: lat + dLat,
		MinLon: lon - dLon,
		MaxLon: lon + dLon,
	}, true
}
