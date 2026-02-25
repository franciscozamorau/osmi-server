package entities

import (
	"errors"
	"strings"
	"time"
)

// Session representa una sesión de usuario en el sistema
// Mapea exactamente la tabla auth.sessions
type Session struct {
	ID int64 `json:"id" db:"id"`
	// CORREGIDO: Mantenemos SessionID para JSON pero db usa session_uuid
	SessionID string `json:"session_id" db:"session_uuid"`
	UserID    int64  `json:"user_id" db:"user_id"`

	RefreshTokenHash string  `json:"-" db:"refresh_token_hash"` // Nunca se expone en JSON
	UserAgent        *string `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress        *string `json:"ip_address,omitempty" db:"ip_address"`

	// CORREGIDO: device_info es JSONB en la BD
	DeviceInfo *map[string]interface{} `json:"device_info,omitempty" db:"device_info,type:jsonb"`

	IsValid       bool       `json:"is_valid" db:"is_valid"`
	InvalidatedAt *time.Time `json:"invalidated_at,omitempty" db:"invalidated_at"`
	ExpiresAt     time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Session

// IsExpired verifica si la sesión ha expirado
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive verifica si la sesión está activa y válida
func (s *Session) IsActive() bool {
	return s.IsValid && !s.IsExpired() && s.InvalidatedAt == nil
}

// IsInvalidated verifica si la sesión ha sido invalidada manualmente
func (s *Session) IsInvalidated() bool {
	return s.InvalidatedAt != nil
}

// CanBeRefreshed verifica si la sesión puede ser refrescada
func (s *Session) CanBeRefreshed() bool {
	// Una sesión puede ser refrescada si está activa o si expiró hace menos de 7 días
	if s.IsActive() {
		return true
	}
	if s.IsExpired() {
		expiryThreshold := s.ExpiresAt.Add(7 * 24 * time.Hour)
		return time.Now().Before(expiryThreshold) && s.InvalidatedAt == nil
	}
	return false
}

// Invalidate invalida la sesión
func (s *Session) Invalidate() {
	now := time.Now()
	s.IsValid = false
	s.InvalidatedAt = &now
	s.UpdatedAt = now
}

// Refresh renueva la sesión con una nueva fecha de expiración
func (s *Session) Refresh(newExpiresAt time.Time, newRefreshTokenHash string) {
	s.ExpiresAt = newExpiresAt
	s.RefreshTokenHash = newRefreshTokenHash
	s.IsValid = true
	s.InvalidatedAt = nil
	s.UpdatedAt = time.Now()
}

// GetDeviceInfoValue obtiene un valor específico de device_info
func (s *Session) GetDeviceInfoValue(key string) interface{} {
	if s.DeviceInfo == nil {
		return nil
	}
	return (*s.DeviceInfo)[key]
}

// SetDeviceInfoValue establece un valor en device_info
func (s *Session) SetDeviceInfoValue(key string, value interface{}) {
	if s.DeviceInfo == nil {
		s.DeviceInfo = &map[string]interface{}{}
	}
	(*s.DeviceInfo)[key] = value
	s.UpdatedAt = time.Now()
}

// GetDeviceType obtiene el tipo de dispositivo (mobile, desktop, tablet)
func (s *Session) GetDeviceType() string {
	if s.DeviceInfo == nil {
		return "unknown"
	}

	// Intentar obtener de diferentes campos comunes
	if userAgent := s.GetDeviceInfoValue("userAgent"); userAgent != nil {
		ua := userAgent.(string)
		// Lógica simple para detectar tipo de dispositivo
		// En producción, usarías una librería como ua-parser
		return detectDeviceType(ua)
	}

	if platform := s.GetDeviceInfoValue("platform"); platform != nil {
		return platform.(string)
	}

	return "unknown"
}

// GetBrowser obtiene el navegador desde device_info
func (s *Session) GetBrowser() string {
	if s.DeviceInfo == nil {
		return "unknown"
	}

	if browser := s.GetDeviceInfoValue("browser"); browser != nil {
		return browser.(string)
	}

	if userAgent := s.GetDeviceInfoValue("userAgent"); userAgent != nil {
		ua := userAgent.(string)
		return detectBrowser(ua)
	}

	return "unknown"
}

// GetOS obtiene el sistema operativo desde device_info
func (s *Session) GetOS() string {
	if s.DeviceInfo == nil {
		return "unknown"
	}

	if os := s.GetDeviceInfoValue("os"); os != nil {
		return os.(string)
	}

	return "unknown"
}

// Validate verifica que la sesión sea válida
func (s *Session) Validate() error {
	if s.UserID == 0 {
		return errors.New("user_id is required")
	}
	if s.RefreshTokenHash == "" {
		return errors.New("refresh_token_hash is required")
	}
	if s.ExpiresAt.IsZero() {
		return errors.New("expires_at is required")
	}
	if s.ExpiresAt.Before(s.CreatedAt) {
		return errors.New("expires_at cannot be before created_at")
	}
	return nil
}

// TimeUntilExpiry obtiene el tiempo hasta que expire la sesión
func (s *Session) TimeUntilExpiry() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return s.ExpiresAt.Sub(time.Now())
}

// GetClientInfo obtiene información completa del cliente
func (s *Session) GetClientInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if s.IPAddress != nil {
		info["ip"] = *s.IPAddress
	}

	if s.UserAgent != nil {
		info["user_agent"] = *s.UserAgent
	}

	if s.DeviceInfo != nil {
		info["device"] = *s.DeviceInfo
	}

	info["device_type"] = s.GetDeviceType()
	info["browser"] = s.GetBrowser()
	info["os"] = s.GetOS()

	return info
}

// Helper functions para detección básica (simplificadas)
func detectDeviceType(userAgent string) string {
	// Implementación simplificada - en producción usar una librería
	switch {
	case strings.Contains(userAgent, "Mobile"):
		return "mobile"
	case strings.Contains(userAgent, "Tablet"):
		return "tablet"
	default:
		return "desktop"
	}
}

func detectBrowser(userAgent string) string {
	// Implementación simplificada
	switch {
	case strings.Contains(userAgent, "Chrome"):
		return "chrome"
	case strings.Contains(userAgent, "Firefox"):
		return "firefox"
	case strings.Contains(userAgent, "Safari"):
		return "safari"
	case strings.Contains(userAgent, "Edge"):
		return "edge"
	default:
		return "unknown"
	}
}
