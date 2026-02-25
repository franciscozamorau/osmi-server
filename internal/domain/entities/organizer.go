package entities

import (
	"errors"
	"time"
)

// Organizer organizador de eventos
// Mapea exactamente la tabla ticketing.organizers
type Organizer struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`

	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`
	LogoURL     *string `json:"logo_url,omitempty" db:"logo_url"`

	// Información legal
	LegalName *string `json:"legal_name,omitempty" db:"legal_name"`
	TaxID     *string `json:"tax_id,omitempty" db:"tax_id"`
	TaxIDType *string `json:"tax_id_type,omitempty" db:"tax_id_type"`
	Country   *string `json:"country,omitempty" db:"country"`

	// Contacto
	ContactEmail string  `json:"contact_email" db:"contact_email"`
	ContactPhone *string `json:"contact_phone,omitempty" db:"contact_phone"`

	// Dirección
	AddressLine1 *string `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty" db:"address_line2"`
	City         *string `json:"city,omitempty" db:"city"`
	State        *string `json:"state,omitempty" db:"state"`
	PostalCode   *string `json:"postal_code,omitempty" db:"postal_code"`

	// Verificación - CAMPO renombrado para evitar conflicto con método
	IsVerifiedField    bool   `json:"is_verified" db:"is_verified"`
	IsActive           bool   `json:"is_active" db:"is_active"`
	VerificationStatus string `json:"verification_status" db:"verification_status"` // pending, verified, rejected

	// Estadísticas
	TotalEvents      int     `json:"total_events" db:"total_events"`
	TotalTicketsSold int64   `json:"total_tickets_sold" db:"total_tickets_sold"`
	OrganizerRating  float64 `json:"organizer_rating,omitempty" db:"organizer_rating"`
	RatingCount      int     `json:"rating_count" db:"rating_count"`

	// Redes sociales (JSONB)
	SocialLinks *map[string]string `json:"social_links,omitempty" db:"social_links,type:jsonb"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Organizer

// IsVerified verifica si el organizador está verificado
func (o *Organizer) IsVerified() bool {
	return o.IsVerifiedField && o.VerificationStatus == "verified"
}

// IsPendingVerification verifica si el organizador está pendiente de verificación
func (o *Organizer) IsPendingVerification() bool {
	return o.VerificationStatus == "pending"
}

// IsRejected verifica si el organizador fue rechazado
func (o *Organizer) IsRejected() bool {
	return o.VerificationStatus == "rejected"
}

// CanCreateEvents verifica si el organizador puede crear eventos
func (o *Organizer) CanCreateEvents() bool {
	return o.IsActive && o.IsVerified()
}

// HasCompleteProfile verifica si el organizador tiene perfil completo
func (o *Organizer) HasCompleteProfile() bool {
	return o.Name != "" &&
		o.Slug != "" &&
		o.ContactEmail != "" &&
		o.LegalName != nil &&
		o.TaxID != nil
}

// HasAddress verifica si el organizador tiene dirección
func (o *Organizer) HasAddress() bool {
	return o.AddressLine1 != nil &&
		o.City != nil &&
		o.Country != nil
}

// GetFullAddress obtiene la dirección completa formateada
func (o *Organizer) GetFullAddress() string {
	if !o.HasAddress() {
		return ""
	}

	address := *o.AddressLine1
	if o.AddressLine2 != nil && *o.AddressLine2 != "" {
		address += ", " + *o.AddressLine2
	}
	if o.City != nil {
		address += ", " + *o.City
	}
	if o.State != nil && *o.State != "" {
		address += ", " + *o.State
	}
	if o.PostalCode != nil && *o.PostalCode != "" {
		address += " " + *o.PostalCode
	}
	if o.Country != nil {
		address += ", " + *o.Country
	}

	return address
}

// UpdateStats actualiza las estadísticas del organizador
func (o *Organizer) UpdateStats(eventCount int, ticketsSold int64) {
	o.TotalEvents += eventCount
	o.TotalTicketsSold += ticketsSold
}

// AddRating añade una calificación y recalcula el promedio
func (o *Organizer) AddRating(rating float64) {
	if rating < 0 || rating > 5 {
		return
	}

	// Calcular nuevo promedio ponderado
	totalScore := o.OrganizerRating * float64(o.RatingCount)
	totalScore += rating
	o.RatingCount++
	o.OrganizerRating = totalScore / float64(o.RatingCount)
}

// SetSocialLink establece un enlace a red social
func (o *Organizer) SetSocialLink(platform string, url string) {
	if o.SocialLinks == nil {
		o.SocialLinks = &map[string]string{}
	}
	(*o.SocialLinks)[platform] = url
}

// GetSocialLink obtiene un enlace a red social
func (o *Organizer) GetSocialLink(platform string) string {
	if o.SocialLinks == nil {
		return ""
	}
	return (*o.SocialLinks)[platform]
}

// DeleteSocialLink elimina un enlace a red social
func (o *Organizer) DeleteSocialLink(platform string) {
	if o.SocialLinks == nil {
		return
	}
	delete(*o.SocialLinks, platform)
	if len(*o.SocialLinks) == 0 {
		o.SocialLinks = nil
	}
}

// Validate verifica que el organizador sea válido
func (o *Organizer) Validate() error {
	if o.Name == "" {
		return errors.New("name is required")
	}
	if o.Slug == "" {
		return errors.New("slug is required")
	}
	if o.ContactEmail == "" {
		return errors.New("contact_email is required")
	}
	if o.VerificationStatus == "" {
		o.VerificationStatus = "pending"
	}
	return nil
}

// Verify marca el organizador como verificado
func (o *Organizer) Verify() {
	o.IsVerifiedField = true
	o.VerificationStatus = "verified"
	o.UpdatedAt = time.Now()
}

// Reject marca el organizador como rechazado
func (o *Organizer) Reject() {
	o.IsVerifiedField = false
	o.VerificationStatus = "rejected"
	o.UpdatedAt = time.Now()
}

// Activate activa el organizador
func (o *Organizer) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}

// Deactivate desactiva el organizador
func (o *Organizer) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}
