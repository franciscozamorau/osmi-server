package services

import (
	"context"
	"fmt"
	"time"

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

type CustomerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

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

	// Usar el repositorio real
	err := s.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer obtiene un cliente por su PublicID
func (s *CustomerService) GetCustomer(ctx context.Context, publicID string) (*entities.Customer, error) {
	if publicID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	// CORREGIDO: FindByPublicID → GetByPublicID
	customer, err := s.customerRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}
