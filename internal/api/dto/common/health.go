// internal/api/dto/common/health.go
package common

import "time"

// HealthCheck representa el estado del servicio
type HealthCheck struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
