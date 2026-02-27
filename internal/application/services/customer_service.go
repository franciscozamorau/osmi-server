// internal/application/services/customer_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

// CreateCustomerRequest - Versión compatible con handler
type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UpdateCustomerRequest - DTO para actualizar cliente
type UpdateCustomerRequest struct {
	Name         *string `json:"name,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	CompanyName  *string `json:"company_name,omitempty"`
	IsVIP        *bool   `json:"is_vip,omitempty"`
	CustomerType *string `json:"customer_type,omitempty"`
	Address      *string `json:"address,omitempty"`
}

type CustomerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

// ============================================================================
// MÉTODOS EXISTENTES
// ============================================================================

// CreateCustomer crea un nuevo cliente
func (s *CustomerService) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*entities.Customer, error) {
	// Validar request
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Crear entidad Customer
	now := time.Now()
	phonePtr := &req.Phone
	if req.Phone == "" {
		phonePtr = nil
	}

	customer := &entities.Customer{
		PublicID:        uuid.New().String(),
		FullName:        req.Name,
		Email:           req.Email,
		Phone:           phonePtr,
		TotalSpent:      0,
		TotalOrders:     0,
		TotalTickets:    0,
		AvgOrderValue:   0,
		IsActive:        true,
		IsVIP:           false,
		CustomerSegment: "new",
		LifetimeValue:   0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer obtiene un cliente por su PublicID
func (s *CustomerService) GetCustomer(ctx context.Context, publicID string) (*entities.Customer, error) {
	if publicID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	customer, err := s.customerRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// ============================================================================
// NUEVOS MÉTODOS (IMPLEMENTADOS)
// ============================================================================

// UpdateCustomer actualiza la información de un cliente
func (s *CustomerService) UpdateCustomer(ctx context.Context, publicID string, req *UpdateCustomerRequest) (*entities.Customer, error) {
	// Obtener el cliente existente
	customer, err := s.customerRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Actualizar campos si se proporcionan
	if req.Name != nil {
		customer.FullName = *req.Name
	}
	if req.Phone != nil {
		customer.Phone = req.Phone
	}
	if req.CompanyName != nil {
		customer.CompanyName = req.CompanyName
	}
	if req.IsVIP != nil {
		customer.IsVIP = *req.IsVIP
	}
	if req.CustomerType != nil {
		customer.CustomerSegment = *req.CustomerType
	}

	customer.UpdatedAt = time.Now()

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// ListCustomers lista clientes con filtros y paginación
func (s *CustomerService) ListCustomers(ctx context.Context, filter *dto.CustomerFilter, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	// Convertir filtro DTO a filtro del repositorio
	repoFilter := &repository.CustomerFilter{
		Limit:  pagination.PageSize,
		Offset: (pagination.Page - 1) * pagination.PageSize,
	}

	if filter != nil {
		if filter.IsActive != nil {
			repoFilter.IsActive = filter.IsActive
		}
		if filter.IsVIP != nil {
			repoFilter.IsVIP = filter.IsVIP
		}
		if filter.Country != "" {
			repoFilter.Country = &filter.Country
		}
		if filter.CustomerSegment != "" {
			repoFilter.CustomerSegment = &filter.CustomerSegment
		}
		if filter.Search != "" {
			repoFilter.SearchTerm = &filter.Search
		}
		if filter.DateFrom != "" {
			// Convertir string a time.Time si es necesario
		}
		if filter.DateTo != "" {
			// Convertir string a time.Time si es necesario
		}
	}

	return s.customerRepo.Find(ctx, repoFilter)
}

// GetCustomerStats obtiene estadísticas globales de clientes
func (s *CustomerService) GetCustomerStats(ctx context.Context) (*dto.CustomerStatsResponse, error) {
	// Usar el método del repositorio
	stats, err := s.customerRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}

	// Convertir a DTO
	return &dto.CustomerStatsResponse{
		TotalCustomers:         stats.TotalCustomers,
		ActiveCustomers:        stats.ActiveCustomers,
		VIPCustomers:           stats.VIPCustomers,
		NewCustomersLast30Days: stats.NewCustomersLast30Days,
		TotalRevenue:           stats.TotalRevenue,
		AvgLifetimeValue:       stats.AvgLifetimeValue,
		TopCountries:           convertCountryStats(stats.TopCountries),
	}, nil
}

// convertCountryStats convierte []repository.CountryStat a []dto.CountryStats
func convertCountryStats(stats []repository.CountryStat) []dto.CountryStats {
	result := make([]dto.CountryStats, len(stats))
	for i, stat := range stats {
		result[i] = dto.CountryStats{
			Country: stat.Country,
			Count:   stat.Count,
			Revenue: stat.Revenue,
		}
	}
	return result
}
