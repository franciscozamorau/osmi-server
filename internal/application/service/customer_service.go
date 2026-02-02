package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type CustomerService struct {
	customerRepo repository.CustomerRepository
	userRepo     repository.UserRepository
}

func NewCustomerService(
	customerRepo repository.CustomerRepository,
	userRepo repository.UserRepository,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		userRepo:     userRepo,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*entities.Customer, error) {
	// Validar que el email no exista
	existing, _ := s.customerRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("customer with this email already exists")
	}

	// Verificar si hay userID y validar que exista
	var userID *int64
	if req.UserID != "" {
		user, err := s.userRepo.FindByPublicID(ctx, req.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}
		userID = &user.ID
	}

	// Crear cliente
	customer := &entities.Customer{
		PublicID:        uuid.New().String(),
		UserID:          userID,
		FullName:        req.FullName,
		Email:           req.Email,
		Phone:           &req.Phone,
		CompanyName:     &req.CompanyName,
		TaxID:           &req.TaxID,
		TaxIDType:       &req.TaxIDType,
		RequiresInvoice: req.RequiresInvoice,
		AddressLine1:    &req.AddressLine1,
		AddressLine2:    &req.AddressLine2,
		City:            &req.City,
		State:           &req.State,
		PostalCode:      &req.PostalCode,
		Country:         &req.Country,
		IsActive:        true,
		CustomerSegment: "new",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, customerID string) (*entities.Customer, error) {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, customerID string, req *dto.UpdateCustomerRequest) (*entities.Customer, error) {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Actualizar campos
	if req.FullName != "" {
		customer.FullName = req.FullName
	}
	if req.Phone != "" {
		customer.Phone = &req.Phone
	}
	if req.CompanyName != "" {
		customer.CompanyName = &req.CompanyName
	}
	if req.TaxID != "" {
		customer.TaxID = &req.TaxID
	}
	if req.TaxIDType != "" {
		customer.TaxIDType = &req.TaxIDType
	}
	if req.RequiresInvoice != nil {
		customer.RequiresInvoice = *req.RequiresInvoice
	}
	if req.AddressLine1 != "" {
		customer.AddressLine1 = &req.AddressLine1
	}
	if req.AddressLine2 != "" {
		customer.AddressLine2 = &req.AddressLine2
	}
	if req.City != "" {
		customer.City = &req.City
	}
	if req.State != "" {
		customer.State = &req.State
	}
	if req.PostalCode != "" {
		customer.PostalCode = &req.PostalCode
	}
	if req.Country != "" {
		customer.Country = &req.Country
	}

	customer.UpdatedAt = time.Now()

	err = s.customerRepo.Update(ctx, customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) ListCustomers(ctx context.Context, filter dto.CustomerFilter, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	customers, err := s.customerRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.customerRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (s *CustomerService) GetCustomerStats(ctx context.Context, customerID string) (*dto.CustomerStatsResponse, error) {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	stats := &dto.CustomerStatsResponse{
		TotalCustomers:         1,
		ActiveCustomers:        1,
		VIPCustomers:           0,
		NewCustomersLast30Days: 0,
		TotalRevenue:           customer.TotalSpent,
		AvgLifetimeValue:       customer.LifetimeValue,
		TopCountries:           []dto.CountryStats{},
	}

	if customer.IsVIP {
		stats.VIPCustomers = 1
	}

	return stats, nil
}

func (s *CustomerService) UpdateCustomerSegment(ctx context.Context, customerID string, segment string) error {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return errors.New("customer not found")
	}

	customer.CustomerSegment = segment
	customer.UpdatedAt = time.Now()

	return s.customerRepo.Update(ctx, customer)
}

func (s *CustomerService) ToggleVIPStatus(ctx context.Context, customerID string, vip bool) error {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return errors.New("customer not found")
	}

	customer.IsVIP = vip
	now := time.Now()
	if vip {
		customer.VIPSince = &now
	} else {
		customer.VIPSince = nil
	}
	customer.UpdatedAt = now

	return s.customerRepo.Update(ctx, customer)
}

func (s *CustomerService) UpdateLoyaltyPoints(ctx context.Context, customerID string, points int32) error {
	customer, err := s.customerRepo.FindByPublicID(ctx, customerID)
	if err != nil {
		return errors.New("customer not found")
	}

	return s.customerRepo.UpdateLoyaltyPoints(ctx, customer.ID, points)
}

func (s *CustomerService) MergeCustomers(ctx context.Context, primaryCustomerID, secondaryCustomerID string) error {
	primary, err := s.customerRepo.FindByPublicID(ctx, primaryCustomerID)
	if err != nil {
		return errors.New("primary customer not found")
	}

	secondary, err := s.customerRepo.FindByPublicID(ctx, secondaryCustomerID)
	if err != nil {
		return errors.New("secondary customer not found")
	}

	// TODO: Implementar lógica de merge
	// 1. Transferir órdenes del secundario al primario
	// 2. Transferir tickets
	// 3. Consolidar estadísticas
	// 4. Marcar secundario como inactivo

	secondary.IsActive = false
	secondary.UpdatedAt = time.Now()

	return s.customerRepo.Update(ctx, secondary)
}
