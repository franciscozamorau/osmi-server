// internal/application/handlers/grpc/customer_handler.go
package grpc

import (
	"context"
	"strconv"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CustomerHandler struct {
	osmi.UnimplementedOsmiServiceServer
	customerService *services.CustomerService
}

func NewCustomerHandler(customerService *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// ============================================================================
// MÉTODOS IMPLEMENTADOS
// ============================================================================

// CreateCustomer maneja la creación de un nuevo cliente
func (h *CustomerHandler) CreateCustomer(ctx context.Context, req *osmi.CreateCustomerRequest) (*osmi.CustomerResponse, error) {
	// Validación básica
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// Convertir a request compatible con el servicio
	createReq := &services.CreateCustomerRequest{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}

	customer, err := h.customerService.CreateCustomer(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Determinar customer type basado en el request o valor por defecto
	customerType := req.CustomerType
	if customerType == "" {
		customerType = "guest"
	}

	return &osmi.CustomerResponse{
		Id:           int32(customer.ID),
		PublicId:     customer.PublicID,
		Name:         customer.FullName,
		Email:        customer.Email,
		Phone:        helpers.SafeStringPtr(customer.Phone),
		CustomerType: customerType,
		IsVip:        customer.IsVIP,
		TotalSpent:   customer.TotalSpent,
		TotalOrders:  int32(customer.TotalOrders),
		CreatedAt:    timestamppb.New(customer.CreatedAt),
		UpdatedAt:    timestamppb.New(customer.UpdatedAt),
	}, nil
}

// GetCustomer obtiene un cliente por diferentes criterios de búsqueda
func (h *CustomerHandler) GetCustomer(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerResponse, error) {
	var (
		customer *entities.Customer
		err      error
	)

	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_PublicId:
		if lookup.PublicId == "" {
			return nil, status.Error(codes.InvalidArgument, "public_id cannot be empty")
		}
		customer, err = h.customerService.GetCustomer(ctx, lookup.PublicId)

	case *osmi.CustomerLookup_Id:
		// Convertir int32 a string para el servicio (el servicio espera string)
		customerID := strconv.FormatInt(int64(lookup.Id), 10)
		customer, err = h.customerService.GetCustomer(ctx, customerID)

	case *osmi.CustomerLookup_Email:
		// Por ahora, no implementado
		return nil, status.Error(codes.Unimplemented, "search by email not implemented")

	default:
		return nil, status.Error(codes.InvalidArgument, "no valid lookup provided")
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &osmi.CustomerResponse{
		Id:           int32(customer.ID),
		PublicId:     customer.PublicID,
		Name:         customer.FullName,
		Email:        customer.Email,
		Phone:        helpers.SafeStringPtr(customer.Phone),
		CustomerType: customer.CustomerSegment,
		IsVip:        customer.IsVIP,
		TotalSpent:   customer.TotalSpent,
		TotalOrders:  int32(customer.TotalOrders),
		CreatedAt:    timestamppb.New(customer.CreatedAt),
		UpdatedAt:    timestamppb.New(customer.UpdatedAt),
	}, nil
}

// UpdateCustomer actualiza la información de un cliente
func (h *CustomerHandler) UpdateCustomer(ctx context.Context, req *osmi.UpdateCustomerRequest) (*osmi.CustomerResponse, error) {
	// Validar que se proporcione el ID
	if req.PublicId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer public_id is required")
	}

	// Convertir protobuf a DTO
	updateReq := &services.UpdateCustomerRequest{
		Name:         req.Name,
		Phone:        req.Phone,
		CompanyName:  req.CompanyName,
		IsVIP:        req.IsVip,
		CustomerType: req.CustomerType,
	}

	customer, err := h.customerService.UpdateCustomer(ctx, req.PublicId, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &osmi.CustomerResponse{
		Id:           int32(customer.ID),
		PublicId:     customer.PublicID,
		Name:         customer.FullName,
		Email:        customer.Email,
		Phone:        helpers.SafeStringPtr(customer.Phone),
		CustomerType: customer.CustomerSegment,
		IsVip:        customer.IsVIP,
		TotalSpent:   customer.TotalSpent,
		TotalOrders:  int32(customer.TotalOrders),
		CreatedAt:    timestamppb.New(customer.CreatedAt),
		UpdatedAt:    timestamppb.New(customer.UpdatedAt),
	}, nil
}

// ListCustomers lista clientes con filtros y paginación
func (h *CustomerHandler) ListCustomers(ctx context.Context, req *osmi.ListCustomersRequest) (*osmi.CustomerListResponse, error) {
	// Convertir filtros
	filter := &dto.CustomerFilter{
		Search:          req.Search,
		Country:         req.Country,
		IsActive:        &req.IsActive,
		IsVIP:           &req.IsVip,
		CustomerSegment: req.CustomerSegment,
		DateFrom:        req.DateFrom,
		DateTo:          req.DateTo,
	}

	// Paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}

	// Llamar al servicio
	customers, total, err := h.customerService.ListCustomers(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir a respuesta
	pbCustomers := make([]*osmi.CustomerResponse, len(customers))
	for i, customer := range customers {
		pbCustomers[i] = &osmi.CustomerResponse{
			Id:           int32(customer.ID),
			PublicId:     customer.PublicID,
			Name:         customer.FullName,
			Email:        customer.Email,
			Phone:        helpers.SafeStringPtr(customer.Phone),
			CustomerType: customer.CustomerSegment,
			IsVip:        customer.IsVIP,
			TotalSpent:   customer.TotalSpent,
			TotalOrders:  int32(customer.TotalOrders),
			CreatedAt:    timestamppb.New(customer.CreatedAt),
			UpdatedAt:    timestamppb.New(customer.UpdatedAt),
		}
	}

	// Calcular total de páginas
	totalPages := int32(0)
	if pagination.PageSize > 0 {
		totalPages = int32((int(total) + pagination.PageSize - 1) / pagination.PageSize)
	}

	return &osmi.CustomerListResponse{
		Customers:  pbCustomers,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: totalPages,
	}, nil
}

// GetCustomerStats obtiene estadísticas de clientes
func (h *CustomerHandler) GetCustomerStats(ctx context.Context, req *osmi.Empty) (*osmi.CustomerStatsResponse, error) {
	// Llamar al servicio
	stats, err := h.customerService.GetCustomerStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir a respuesta
	topCountries := make([]*osmi.CountryStats, len(stats.TopCountries))
	for i, country := range stats.TopCountries {
		topCountries[i] = &osmi.CountryStats{
			Country: country.Country,
			Count:   int64(country.Count),
			Revenue: country.Revenue,
		}
	}

	return &osmi.CustomerStatsResponse{
		TotalCustomers:          stats.TotalCustomers,
		ActiveCustomers:         stats.ActiveCustomers,
		VipCustomers:            stats.VIPCustomers,
		NewCustomersLast_30Days: stats.NewCustomersLast30Days,
		TotalRevenue:            stats.TotalRevenue,
		AvgLifetimeValue:        stats.AvgLifetimeValue,
		TopCountries:            topCountries,
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
