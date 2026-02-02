package types

import (
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TicketConverter maneja conversiones específicas para tickets
type TicketConverter struct {
	*Converter
}

// NewTicketConverter crea un nuevo TicketConverter
func NewTicketConverter() *TicketConverter {
	return &TicketConverter{
		Converter: NewConverter(),
	}
}

// TicketCode convierte código de ticket
func (tc *TicketConverter) TicketCode(code string) pgtype.Text {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return pgtype.Text{Valid: false}
	}
	return tc.Text(code)
}

// TicketStatus convierte estado de ticket
func (tc *TicketConverter) TicketStatus(status string) pgtype.Text {
	validStatuses := map[string]bool{
		"available":   true,
		"reserved":    true,
		"sold":        true,
		"used":        true,
		"cancelled":   true,
		"transferred": true,
		"refunded":    true,
		"expired":     true,
		"pending":     true,
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if !validStatuses[status] {
		return pgtype.Text{Valid: false}
	}
	return tc.Text(status)
}

// TicketType convierte tipo de ticket
func (tc *TicketConverter) TicketType(ticketType string) pgtype.Text {
	validTypes := map[string]bool{
		"general":    true,
		"vip":        true,
		"premium":    true,
		"student":    true,
		"senior":     true,
		"group":      true,
		"early_bird": true,
		"late":       true,
	}

	ticketType = strings.ToLower(strings.TrimSpace(ticketType))
	if !validTypes[ticketType] {
		return pgtype.Text{Valid: false}
	}
	return tc.Text(ticketType)
}

// TicketPrice convierte precio de ticket
func (tc *TicketConverter) TicketPrice(price float64, currency string) pgtype.Text {
	if price < 0 {
		return pgtype.Text{Valid: false}
	}

	formatted := formatPrice(price, currency)
	return tc.Text(formatted)
}

// TicketPricePtr convierte *precio de ticket
func (tc *TicketConverter) TicketPricePtr(price *float64, currency string) pgtype.Text {
	if price == nil {
		return pgtype.Text{Valid: false}
	}
	return tc.TicketPrice(*price, currency)
}

// TicketSeat convierte información de asiento
func (tc *TicketConverter) TicketSeat(section, row, number string) pgtype.Text {
	if section == "" && row == "" && number == "" {
		return pgtype.Text{Valid: false}
	}

	seatInfo := strings.Join([]string{section, row, number}, "-")
	return tc.Text(seatInfo)
}

// TicketEventDate convierte fecha de evento
func (tc *TicketConverter) TicketEventDate(t time.Time) pgtype.Timestamp {
	return tc.Timestamp(t)
}

// TicketEventDatePtr convierte *fecha de evento
func (tc *TicketConverter) TicketEventDatePtr(t *time.Time) pgtype.Timestamp {
	return tc.TimestampPtr(t)
}

// TicketQRCode convierte datos de QR code
func (tc *TicketConverter) TicketQRCode(data string) pgtype.Text {
	if data == "" {
		return pgtype.Text{Valid: false}
	}
	return tc.Text(data)
}

// TicketPurchaseDate convierte fecha de compra
func (tc *TicketConverter) TicketPurchaseDate(t time.Time) pgtype.Timestamp {
	return tc.Timestamp(t)
}

// TicketPurchaseDatePtr convierte *fecha de compra
func (tc *TicketConverter) TicketPurchaseDatePtr(t *time.Time) pgtype.Timestamp {
	return tc.TimestampPtr(t)
}

// TicketValidUntil convierte fecha de validez
func (tc *TicketConverter) TicketValidUntil(t time.Time) pgtype.Timestamp {
	return tc.Timestamp(t)
}

// TicketValidUntilPtr convierte *fecha de validez
func (tc *TicketConverter) TicketValidUntilPtr(t *time.Time) pgtype.Timestamp {
	return tc.TimestampPtr(t)
}

// formatPrice formatea precio con moneda
func formatPrice(price float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "USD", "CAD", "AUD":
		return "$" + strconv.FormatFloat(price, 'f', 2, 64)
	case "EUR":
		return "€" + strconv.FormatFloat(price, 'f', 2, 64)
	case "GBP":
		return "£" + strconv.FormatFloat(price, 'f', 2, 64)
	default:
		return strconv.FormatFloat(price, 'f', 2, 64) + " " + currency
	}
}
