package services

import (
	"context"
	"fmt"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type CustomerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*entities.Customer, error) {
	// Validar request
	if req.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Crear entidad Customer
	customer := &entities.Customer{
		PublicID:                 uuid.New().String(),
		FullName:                 req.FullName,
		Email:                    req.Email,
		Phone:                    req.Phone,
		CompanyName:              req.CompanyName,
		AddressLine1:             req.AddressLine1,
		AddressLine2:             req.AddressLine2,
		City:                     req.City,
		State:                    req.State,
		PostalCode:               req.PostalCode,
		Country:                  req.Country,
		TaxID:                    req.TaxID,
		TaxIDType:                req.TaxIDType,
		TaxName:                  req.TaxName,
		RequiresInvoice:          req.RequiresInvoice,
		CommunicationPreferences: make(map[string]interface{}),
		TotalSpent:               0,
		TotalOrders:              0,
		TotalTickets:             0,
		AvgOrderValue:            0,
		IsActive:                 true,
		IsVIP:                    false,
		CustomerSegment:          "new",
		LifetimeValue:            0,
	}

	// Usar el repositorio real
	err := s.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, publicID string) (*entities.Customer, error) {
	if publicID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	customer, err := s.customerRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, publicID string, req *dto.UpdateCustomerRequest) (*entities.Customer, error) {
	// Primero obtener el cliente existente
	customer, err := s.customerRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Actualizar campos permitidos
	if req.FullName != "" {
		customer.FullName = req.FullName
	}
	if req.Phone != nil {
		customer.Phone = req.Phone
	}
	if req.CompanyName != nil {
		customer.CompanyName = req.CompanyName
	}
	if req.AddressLine1 != nil {
		customer.AddressLine1 = req.AddressLine1
	}
	if req.City != nil {
		customer.City = req.City
	}
	if req.State != nil {
		customer.State = req.State
	}
	if req.PostalCode != nil {
		customer.PostalCode = req.PostalCode
	}
	if req.Country != nil {
		customer.Country = req.Country
	}

	// Guardar cambios
	err = s.customerRepo.Update(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

func (s *CustomerService) ListCustomers(ctx context.Context, filter dto.CustomerFilter, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	return s.customerRepo.List(ctx, filter, pagination)
}

func (s *CustomerService) GetCustomerStats(ctx context.Context, publicID string) (*dto.CustomerStatsResponse, error) {
	// Implementación básica - puedes expandirla
	customer, err := s.customerRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	return &dto.CustomerStatsResponse{
		TotalCustomers:         1,
		ActiveCustomers:        1,
		VIPCustomers:           0,
		NewCustomersLast30Days: 0,
		TotalRevenue:           customer.TotalSpent,
		AvgLifetimeValue:       customer.LifetimeValue,
		TopCountries:           []dto.CountryStats{},
	}, nil
}
