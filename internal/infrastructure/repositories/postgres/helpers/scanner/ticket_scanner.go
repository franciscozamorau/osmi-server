package scanner

import (
	"database/sql"
	"fmt"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"

	"github.com/jackc/pgx/v5"
)

// TicketScanner escanea resultados específicos de tickets
type TicketScanner struct {
	*RowScanner
}

// NewTicketScanner crea un nuevo TicketScanner
func NewTicketScanner() *TicketScanner {
	return &TicketScanner{
		RowScanner: NewRowScanner(),
	}
}

// ScanTicket escanea una fila completa a entidad Ticket
func (ts *TicketScanner) ScanTicket(row pgx.Row) (*entities.Ticket, error) {
	var ticket entities.Ticket
	var code sql.NullString
	var seatInfo sql.NullString
	var qrCode sql.NullString
	var purchaseDate sql.NullTime
	var validUntil sql.NullTime
	var usedAt sql.NullTime
	var cancelledAt sql.NullTime
	var transferredAt sql.NullTime
	var refundedAt sql.NullTime

	err := row.Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.EventID,
		&ticket.UserID,
		&ticket.OrderID,
		&code,
		&ticket.Type,
		&ticket.Status,
		&ticket.Price,
		&ticket.Currency,
		&seatInfo,
		&qrCode,
		&purchaseDate,
		&validUntil,
		&usedAt,
		&cancelledAt,
		&transferredAt,
		&refundedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to scan ticket: %w", err)
	}

	ticket.Code = ts.ConvertSQLNullable(code)
	ticket.SeatInfo = ts.ConvertSQLNullable(seatInfo)
	ticket.QRCode = ts.ConvertSQLNullable(qrCode)
	ticket.PurchaseDate = ts.ConvertSQLNullableTime(purchaseDate)
	ticket.ValidUntil = ts.ConvertSQLNullableTime(validUntil)
	ticket.UsedAt = ts.ConvertSQLNullableTime(usedAt)
	ticket.CancelledAt = ts.ConvertSQLNullableTime(cancelledAt)
	ticket.TransferredAt = ts.ConvertSQLNullableTime(transferredAt)
	ticket.RefundedAt = ts.ConvertSQLNullableTime(refundedAt)

	return &ticket, nil
}

// ScanTicketBasic escanea campos básicos de ticket
func (ts *TicketScanner) ScanTicketBasic(row pgx.Row) (*entities.TicketBasic, error) {
	var ticket entities.TicketBasic
	var code sql.NullString
	var validUntil sql.NullTime

	err := row.Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.EventID,
		&code,
		&ticket.Type,
		&ticket.Status,
		&ticket.Price,
		&ticket.Currency,
		&validUntil,
		&ticket.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to scan ticket basic: %w", err)
	}

	ticket.Code = ts.ConvertSQLNullable(code)
	ticket.ValidUntil = ts.ConvertSQLNullableTime(validUntil)

	return &ticket, nil
}

// ScanTicketWithEvent escanea ticket con información de evento
func (ts *TicketScanner) ScanTicketWithEvent(row pgx.Row) (*entities.TicketWithEvent, error) {
	var ticket entities.TicketWithEvent
	var eventName sql.NullString
	var eventDate sql.NullTime
	var venueName sql.NullString
	var ticketCode sql.NullString
	var validUntil sql.NullTime

	err := row.Scan(
		&ticket.TicketID,
		&ticket.PublicID,
		&ticket.EventID,
		&eventName,
		&eventDate,
		&venueName,
		&ticketCode,
		&ticket.Type,
		&ticket.Status,
		&ticket.Price,
		&ticket.Currency,
		&validUntil,
		&ticket.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to scan ticket with event: %w", err)
	}

	ticket.EventName = ts.ConvertSQLNullable(eventName)
	ticket.EventDate = ts.ConvertSQLNullableTime(eventDate)
	ticket.VenueName = ts.ConvertSQLNullable(venueName)
	ticket.Code = ts.ConvertSQLNullable(ticketCode)
	ticket.ValidUntil = ts.ConvertSQLNullableTime(validUntil)

	return &ticket, nil
}

// ScanTicketStats escanea estadísticas de tickets
func (ts *TicketScanner) ScanTicketStats(row pgx.Row) (*entities.TicketStats, error) {
	var stats entities.TicketStats
	var lastSale sql.NullTime

	err := row.Scan(
		&stats.EventID,
		&stats.TotalTickets,
		&stats.Available,
		&stats.Sold,
		&stats.Used,
		&stats.Cancelled,
		&stats.Refunded,
		&stats.TotalRevenue,
		&lastSale,
		&stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan ticket stats: %w", err)
	}

	stats.LastSale = ts.ConvertSQLNullableTime(lastSale)

	return &stats, nil
}

// ScanTicketValidation escanea datos para validación de ticket
func (ts *TicketScanner) ScanTicketValidation(row pgx.Row) (*entities.TicketValidation, error) {
	var validation entities.TicketValidation
	var validUntil sql.NullTime
	var usedAt sql.NullTime
	var eventDate sql.NullTime

	err := row.Scan(
		&validation.TicketID,
		&validation.PublicID,
		&validation.Code,
		&validation.Status,
		&validUntil,
		&usedAt,
		&validation.EventID,
		&eventDate,
		&validation.VenueID,
		&validation.UserID,
		&validation.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
		return nil, fmt.Errorf("failed to scan ticket validation: %w", err)
	}

	validation.ValidUntil = ts.ConvertSQLNullableTime(validUntil)
	validation.UsedAt = ts.ConvertSQLNullableTime(usedAt)
	validation.EventDate = ts.ConvertSQLNullableTime(eventDate)

	return &validation, nil
}
