// internal/api/dto/common/map_bounds.go
package common

// MapBounds representa límites geográficos para mapas
type MapBounds struct {
	NorthEast GeoLocation `json:"north_east"`
	SouthWest GeoLocation `json:"south_west"`
}
