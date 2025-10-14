package context

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Context keys
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	IPAddressKey contextKey = "ip_address"
	UserAgentKey contextKey = "user_agent"
)

// AuditContext contiene información de auditoría
type AuditContext struct {
	UserID    string
	IPAddress string
	UserAgent string
	Metadata  map[string]interface{}
}

// ExtractAuditContext extrae información de auditoría del contexto
func ExtractAuditContext(ctx context.Context) *AuditContext {
	auditCtx := &AuditContext{
		Metadata: make(map[string]interface{}),
	}

	// Extraer UserID
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		auditCtx.UserID = userID
	} else {
		auditCtx.UserID = "system" // Default para operaciones del sistema
	}

	// Extraer IP Address
	if ip, ok := ctx.Value(IPAddressKey).(string); ok {
		auditCtx.IPAddress = ip
	} else {
		auditCtx.IPAddress = "127.0.0.1" // Default
	}

	// Extraer User Agent
	if ua, ok := ctx.Value(UserAgentKey).(string); ok {
		auditCtx.UserAgent = ua
	} else {
		auditCtx.UserAgent = "osmi-server" // Default
	}

	return auditCtx
}

// WithUserID agrega UserID al contexto
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithIPAddress agrega IP Address al contexto
func WithIPAddress(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ip)
}

// WithUserAgent agrega User Agent al contexto
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

// ExtractFromHTTPRequest extrae información de auditoría de un HTTP request
func ExtractFromHTTPRequest(r *http.Request) context.Context {
	ctx := r.Context()

	// Extraer UserID del header (ejemplo, en realidad vendría del JWT)
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		ctx = WithUserID(ctx, userID)
	}

	// Extraer IP Address
	ip := getClientIP(r)
	ctx = WithIPAddress(ctx, ip)

	// Extraer User Agent
	userAgent := r.UserAgent()
	ctx = WithUserAgent(ctx, userAgent)

	return ctx
}

// getClientIP obtiene la IP real del cliente
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For puede contener múltiples IPs, tomar la primera
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback a RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
