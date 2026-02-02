package validations

import (
	"strings"
	"time"
)

// IsValidTicketStatus valida estado de ticket
func IsValidTicketStatus(status string) bool {
	validStatuses := []string{
		"available", "reserved", "sold", "used",
		"cancelled", "transferred", "refunded",
		"pending", "expired", "blocked", "activated",
	}

	status = strings.ToLower(strings.TrimSpace(status))
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidTicketStatusWithReason valida estado de ticket con razón
func IsValidTicketStatusWithReason(status string) (bool, string) {
	status = strings.ToLower(strings.TrimSpace(status))

	if status == "" {
		return false, "ticket status cannot be empty"
	}

	if !IsValidTicketStatus(status) {
		validStatuses := []string{
			"available", "reserved", "sold", "used",
			"cancelled", "transferred", "refunded",
			"pending", "expired", "blocked", "activated",
		}
		return false, "invalid ticket status. Must be one of: " + strings.Join(validStatuses, ", ")
	}

	return true, ""
}

// IsValidTicketType valida tipo de ticket
func IsValidTicketType(ticketType string) bool {
	validTypes := []string{
		"general", "vip", "premium", "student", "senior",
		"group", "early_bird", "late", "family", "child",
		"corporate", "sponsor", "media", "complimentary",
	}

	ticketType = strings.ToLower(strings.TrimSpace(ticketType))
	for _, validType := range validTypes {
		if ticketType == validType {
			return true
		}
	}
	return false
}

// IsValidCustomerType valida tipo de cliente
func IsValidCustomerType(customerType string) bool {
	validTypes := []string{
		"registered", "guest", "corporate", "vip",
		"student", "senior", "wholesale", "retail",
		"reseller", "affiliate", "partner",
	}

	customerType = strings.ToLower(strings.TrimSpace(customerType))
	for _, validType := range validTypes {
		if customerType == validType {
			return true
		}
	}
	return false
}

// IsValidEventStatus valida estado de evento
func IsValidEventStatus(status string) bool {
	validStatuses := []string{
		"draft", "published", "cancelled", "completed",
		"sold_out", "ongoing", "postponed", "archived",
		"private", "public", "hidden",
	}

	status = strings.ToLower(strings.TrimSpace(status))
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidEventType valida tipo de evento
func IsValidEventType(eventType string) bool {
	validTypes := []string{
		"concert", "conference", "workshop", "seminar",
		"festival", "exhibition", "sports", "theater",
		"movie", "party", "networking", "charity",
		"corporate", "educational", "religious",
	}

	eventType = strings.ToLower(strings.TrimSpace(eventType))
	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

// IsValidUserRole valida rol de usuario
func IsValidUserRole(role string) bool {
	validRoles := []string{
		"admin", "customer", "organizer", "guest",
		"staff", "moderator", "super_admin", "vendor",
		"support", "manager", "editor", "viewer",
	}

	role = strings.ToLower(strings.TrimSpace(role))
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsValidOrderStatus valida estado de pedido
func IsValidOrderStatus(status string) bool {
	validStatuses := []string{
		"pending", "confirmed", "processing", "shipped",
		"delivered", "cancelled", "refunded", "failed",
		"partially_refunded", "on_hold", "completed",
	}

	status = strings.ToLower(strings.TrimSpace(status))
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPaymentStatus valida estado de pago
func IsValidPaymentStatus(status string) bool {
	validStatuses := []string{
		"pending", "processing", "completed", "failed",
		"refunded", "cancelled", "partially_refunded",
		"authorized", "captured", "voided",
	}

	status = strings.ToLower(strings.TrimSpace(status))
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPaymentMethod valida método de pago
func IsValidPaymentMethod(method string) bool {
	validMethods := []string{
		"credit_card", "debit_card", "paypal", "stripe",
		"bank_transfer", "cash", "check", "mobile_payment",
		"crypto", "gift_card", "invoice", "subscription",
	}

	method = strings.ToLower(strings.TrimSpace(method))
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// IsValidDiscountType valida tipo de descuento
func IsValidDiscountType(discountType string) bool {
	validTypes := []string{
		"percentage", "fixed", "buy_one_get_one", "seasonal",
		"early_bird", "group", "student", "senior", "vip",
		"promo_code", "loyalty", "referral",
	}

	discountType = strings.ToLower(strings.TrimSpace(discountType))
	for _, validType := range validTypes {
		if discountType == validType {
			return true
		}
	}
	return false
}

// IsValidNotificationType valida tipo de notificación
func IsValidNotificationType(notificationType string) bool {
	validTypes := []string{
		"email", "sms", "push", "in_app", "webhook",
		"slack", "teams", "discord", "whatsapp",
	}

	notificationType = strings.ToLower(strings.TrimSpace(notificationType))
	for _, validType := range validTypes {
		if notificationType == validType {
			return true
		}
	}
	return false
}

// IsValidVenueType valida tipo de venue
func IsValidVenueType(venueType string) bool {
	validTypes := []string{
		"stadium", "arena", "theater", "auditorium",
		"convention_center", "hotel", "restaurant",
		"club", "bar", "outdoor", "museum", "gallery",
		"park", "beach", "stadium", "racetrack",
	}

	venueType = strings.ToLower(strings.TrimSpace(venueType))
	for _, validType := range validTypes {
		if venueType == validType {
			return true
		}
	}
	return false
}

// IsValidSeatType valida tipo de asiento
func IsValidSeatType(seatType string) bool {
	validTypes := []string{
		"general", "premium", "vip", "box", "balcony",
		"orchestra", "mezzanine", "standing", "table",
		"booth", "couch", "barstool",
	}

	seatType = strings.ToLower(strings.TrimSpace(seatType))
	for _, validType := range validTypes {
		if seatType == validType {
			return true
		}
	}
	return false
}

// ValidateTicketPrice valida precio de ticket
func ValidateTicketPrice(price float64, currency string) (bool, string) {
	if price < 0 {
		return false, "ticket price cannot be negative"
	}

	if price > 1000000 {
		return false, "ticket price cannot exceed 1,000,000"
	}

	if currency == "" {
		return false, "currency is required"
	}

	if !IsValidCurrencyCode(currency) {
		return false, "invalid currency code"
	}

	return true, ""
}

// ValidateEventDates valida fechas de evento
func ValidateEventDates(start, end time.Time) (bool, string) {
	if start.IsZero() {
		return false, "event start date is required"
	}

	if end.IsZero() {
		return false, "event end date is required"
	}

	if !end.After(start) {
		return false, "event end date must be after start date"
	}

	// Validar que no sea en el pasado (para nuevos eventos)
	if start.Before(time.Now()) {
		return false, "event cannot start in the past"
	}

	// Validar duración máxima (30 días)
	maxDuration := 30 * 24 * time.Hour
	if end.Sub(start) > maxDuration {
		return false, "event duration cannot exceed 30 days"
	}

	return true, ""
}

// ValidateTicketQuantity valida cantidad de tickets
func ValidateTicketQuantity(quantity int, available int) (bool, string) {
	if quantity <= 0 {
		return false, "quantity must be greater than 0"
	}

	if quantity > 20 {
		return false, "cannot purchase more than 20 tickets at once"
	}

	if quantity > available {
		return false, "not enough tickets available"
	}

	return true, ""
}

// ValidateAgeRequirement valida requisito de edad
func ValidateAgeRequirement(dateOfBirth time.Time, minAge int) (bool, string) {
	if dateOfBirth.IsZero() {
		return false, "date of birth is required"
	}

	age := time.Now().Year() - dateOfBirth.Year()
	if time.Now().YearDay() < dateOfBirth.YearDay() {
		age--
	}

	if age < minAge {
		return false, "must be at least 18 years old"
	}

	return true, ""
}

// ValidatePromoCode valida código promocional
func ValidatePromoCode(code string, expiry time.Time) (bool, string) {
	if code == "" {
		return false, "promo code cannot be empty"
	}

	if len(code) < 4 || len(code) > 20 {
		return false, "promo code must be between 4 and 20 characters"
	}

	if !expiry.IsZero() && expiry.Before(time.Now()) {
		return false, "promo code has expired"
	}

	return true, ""
}

// ValidateReservationWindow valida ventana de reservación
func ValidateReservationWindow(reservationTime, eventTime time.Time, maxHours int) (bool, string) {
	if reservationTime.IsZero() || eventTime.IsZero() {
		return false, "invalid times provided"
	}

	hoursUntilEvent := eventTime.Sub(reservationTime).Hours()
	if hoursUntilEvent < 0 {
		return false, "cannot reserve for past events"
	}

	if hoursUntilEvent > float64(maxHours*24) {
		return false, "reservation window is too far in advance"
	}

	return true, ""
}
