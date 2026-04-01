// internal/api/dto/customer/filter.go
package customer

type CustomerFilter struct {
	Search          string `json:"search,omitempty"`
	Country         string `json:"country,omitempty"`
	IsActive        *bool  `json:"is_active,omitempty"`
	IsVIP           *bool  `json:"is_vip,omitempty"`
	CustomerSegment string `json:"customer_segment,omitempty"`
	DateFrom        string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo          string `json:"date_to,omitempty" validate:"omitempty,date"`
}
