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
	var qrCodeData sql.NullString
	var attendeeName sql.NullString
	var attendeeEmail sql.NullString
	var attendeePhone sql.NullString
	var checkedInAt sql.NullTime
	var checkinMethod sql.NullString
	var checkinLocation sql.NullString
	var reservedAt sql.NullTime
	var reservationExpiresAt sql.NullTime
	var transferToken sql.NullString
	var transferredAt sql.NullTime
	var lastValidatedAt sql.NullTime
	var soldAt sql.NullTime
	var cancelledAt sql.NullTime
	var refundedAt sql.NullTime

	err := row.Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.TicketTypeID,
		&ticket.EventID,
		&ticket.CustomerID,
		&ticket.OrderID,
		&ticket.Code,
		&ticket.SecretHash,
		&qrCodeData,
		&ticket.Status,
		&ticket.FinalPrice,
		&ticket.Currency,
		&ticket.TaxAmount,
		&attendeeName,
		&attendeeEmail,
		&attendeePhone,
		&checkedInAt,
		&ticket.CheckedInBy,
		&checkinMethod,
		&checkinLocation,
		&reservedAt,
		&ticket.ReservedBy,
		&reservationExpiresAt,
		&transferToken,
		&ticket.TransferredFrom,
		&transferredAt,
		&ticket.ValidationCount,
		&lastValidatedAt,
		&soldAt,
		&cancelledAt,
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

	// Convertir Null types a pointers
	ticket.QRCodeData = ts.ConvertSQLNullable(qrCodeData)
	ticket.AttendeeName = ts.ConvertSQLNullable(attendeeName)
	ticket.AttendeeEmail = ts.ConvertSQLNullable(attendeeEmail)
	ticket.AttendeePhone = ts.ConvertSQLNullable(attendeePhone)
	ticket.CheckedInAt = ts.ConvertSQLNullableTime(checkedInAt)
	ticket.CheckinMethod = ts.ConvertSQLNullable(checkinMethod)
	ticket.CheckinLocation = ts.ConvertSQLNullable(checkinLocation)
	ticket.ReservedAt = ts.ConvertSQLNullableTime(reservedAt)
	ticket.ReservationExpiresAt = ts.ConvertSQLNullableTime(reservationExpiresAt)
	ticket.TransferToken = ts.ConvertSQLNullable(transferToken)
	ticket.TransferredAt = ts.ConvertSQLNullableTime(transferredAt)
	ticket.LastValidatedAt = ts.ConvertSQLNullableTime(lastValidatedAt)
	ticket.SoldAt = ts.ConvertSQLNullableTime(soldAt)
	ticket.CancelledAt = ts.ConvertSQLNullableTime(cancelledAt)
	ticket.RefundedAt = ts.ConvertSQLNullableTime(refundedAt)

	return &ticket, nil
}
