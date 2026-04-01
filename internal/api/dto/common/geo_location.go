// internal/api/dto/common/geo_location.go
package common

// GeoLocation representa coordenadas geográficas
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
