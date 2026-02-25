// internal/application/handlers/grpc/customer_handler.go
package grpc

import (
	"context"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
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

// CreateCustomer maneja la creación de un nuevo cliente
func (h *CustomerHandler) CreateCustomer(ctx context.Context, req *osmi.CreateCustomerRequest) (*osmi.CustomerResponse, error) {
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

	// Usar solo los campos que existen en CustomerResponse según el proto
	return &osmi.CustomerResponse{
		Id:           int32(customer.ID),
		PublicId:     customer.PublicID,
		Name:         customer.FullName,
		Email:        customer.Email,
		Phone:        helpers.SafeStringPtr(customer.Phone),
		CustomerType: "guest", // Valor por defecto, ajustar según necesidad
		IsVip:        customer.IsVIP,
		TotalSpent:   customer.TotalSpent,
		TotalOrders:  int32(customer.TotalOrders),
		CreatedAt:    timestamppb.New(customer.CreatedAt),
		UpdatedAt:    timestamppb.New(customer.UpdatedAt),
	}, nil
}

// GetCustomer obtiene un cliente por diferentes criterios de búsqueda
func (h *CustomerHandler) GetCustomer(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerResponse, error) {
	var customerID string

	// Manejar diferentes formas de búsqueda
	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_PublicId:
		customerID = lookup.PublicId
	case *osmi.CustomerLookup_Id:
		return nil, status.Error(codes.Unimplemented, "search by numeric ID not implemented")
	case *osmi.CustomerLookup_Email:
		return nil, status.Error(codes.Unimplemented, "search by email not implemented")
	default:
		return nil, status.Error(codes.InvalidArgument, "no valid lookup provided")
	}

	if customerID == "" {
		return nil, status.Error(codes.InvalidArgument, "customer ID is required")
	}

	customer, err := h.customerService.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Usar solo los campos que existen en CustomerResponse
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

	// TODO: Implementar cuando el servicio lo soporte
	return nil, status.Error(codes.Unimplemented, "UpdateCustomer not implemented yet")
}

// ListCustomers lista clientes con filtros y paginación
func (h *CustomerHandler) ListCustomers(ctx context.Context, req *osmi.ListCustomersRequest) (*osmi.CustomerListResponse, error) {
	// TODO: Implementar cuando el servicio lo soporte
	return nil, status.Error(codes.Unimplemented, "ListCustomers not implemented yet")
}

// GetCustomerStats obtiene estadísticas de clientes
func (h *CustomerHandler) GetCustomerStats(ctx context.Context, req *osmi.Empty) (*osmi.CustomerStatsResponse, error) {
	// TODO: Implementar cuando el servicio lo soporte
	return nil, status.Error(codes.Unimplemented, "GetCustomerStats not implemented yet")
}
