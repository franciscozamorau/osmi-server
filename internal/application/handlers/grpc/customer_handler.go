package grpc

import (
	"context"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
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

func (h *CustomerHandler) CreateCustomer(ctx context.Context, req *osmi.CustomerRequest) (*osmi.CustomerResponse, error) {
	// Convertir protobuf a DTO
	createReq := &dto.CreateCustomerRequest{
		UserID:          req.UserId,
		FullName:        req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		CompanyName:     req.CompanyName,
		TaxID:           req.TaxId,
		TaxIDType:       req.TaxIdType,
		RequiresInvoice: req.RequiresInvoice,
		AddressLine1:    req.AddressLine1,
		AddressLine2:    req.AddressLine2,
		City:            req.City,
		State:           req.State,
		PostalCode:      req.PostalCode,
		Country:         req.Country,
	}

	// Llamar al servicio
	customer, err := h.customerService.CreateCustomer(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.CustomerResponse{
		Id:              customer.PublicID,
		UserId:          safeStringID(customer.UserID),
		Name:            customer.FullName,
		Email:           customer.Email,
		Phone:           safeStringPtr(customer.Phone),
		CompanyName:     safeStringPtr(customer.CompanyName),
		TaxId:           safeStringPtr(customer.TaxID),
		TaxIdType:       safeStringPtr(customer.TaxIDType),
		RequiresInvoice: customer.RequiresInvoice,
		AddressLine1:    safeStringPtr(customer.AddressLine1),
		AddressLine2:    safeStringPtr(customer.AddressLine2),
		City:            safeStringPtr(customer.City),
		State:           safeStringPtr(customer.State),
		PostalCode:      safeStringPtr(customer.PostalCode),
		Country:         safeStringPtr(customer.Country),
		TotalSpent:      customer.TotalSpent,
		TotalOrders:     int32(customer.TotalOrders),
		TotalTickets:    int32(customer.TotalTickets),
		AvgOrderValue:   customer.AvgOrderValue,
		FirstOrderAt:    safeTimeProtoFromString(customer.FirstOrderAt),
		LastOrderAt:     safeTimeProtoFromString(customer.LastOrderAt),
		IsActive:        customer.IsActive,
		IsVip:           customer.IsVIP,
		VipSince:        safeTimeProto(customer.VIPSince),
		CustomerSegment: customer.CustomerSegment,
		LifetimeValue:   customer.LifetimeValue,
		CreatedAt:       timestamppb.New(customer.CreatedAt),
		UpdatedAt:       timestamppb.New(customer.UpdatedAt),
	}, nil
}

func (h *CustomerHandler) GetCustomer(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerResponse, error) {
	var customerID string

	// Manejar diferentes formas de búsqueda
	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_Id:
		customerID = lookup.Id
	case *osmi.CustomerLookup_Email:
		// TODO: Implementar búsqueda por email
		return nil, status.Error(codes.Unimplemented, "search by email not implemented")
	default:
		return nil, status.Error(codes.InvalidArgument, "no valid lookup provided")
	}

	// Llamar al servicio
	customer, err := h.customerService.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.CustomerResponse{
		Id:              customer.PublicID,
		UserId:          safeStringID(customer.UserID),
		Name:            customer.FullName,
		Email:           customer.Email,
		Phone:           safeStringPtr(customer.Phone),
		CompanyName:     safeStringPtr(customer.CompanyName),
		TaxId:           safeStringPtr(customer.TaxID),
		TaxIdType:       safeStringPtr(customer.TaxIDType),
		RequiresInvoice: customer.RequiresInvoice,
		AddressLine1:    safeStringPtr(customer.AddressLine1),
		AddressLine2:    safeStringPtr(customer.AddressLine2),
		City:            safeStringPtr(customer.City),
		State:           safeStringPtr(customer.State),
		PostalCode:      safeStringPtr(customer.PostalCode),
		Country:         safeStringPtr(customer.Country),
		TotalSpent:      customer.TotalSpent,
		TotalOrders:     int32(customer.TotalOrders),
		TotalTickets:    int32(customer.TotalTickets),
		AvgOrderValue:   customer.AvgOrderValue,
		FirstOrderAt:    safeTimeProtoFromString(customer.FirstOrderAt),
		LastOrderAt:     safeTimeProtoFromString(customer.LastOrderAt),
		IsActive:        customer.IsActive,
		IsVip:           customer.IsVIP,
		VipSince:        safeTimeProto(customer.VIPSince),
		CustomerSegment: customer.CustomerSegment,
		LifetimeValue:   customer.LifetimeValue,
		CreatedAt:       timestamppb.New(customer.CreatedAt),
		UpdatedAt:       timestamppb.New(customer.UpdatedAt),
	}, nil
}

func (h *CustomerHandler) UpdateCustomer(ctx context.Context, req *osmi.UpdateCustomerRequest) (*osmi.CustomerResponse, error) {
	// Convertir protobuf a DTO
	updateReq := &dto.UpdateCustomerRequest{
		FullName:        req.Name,
		Phone:           req.Phone,
		CompanyName:     req.CompanyName,
		TaxID:           req.TaxId,
		TaxIDType:       req.TaxIdType,
		RequiresInvoice: req.RequiresInvoice,
		AddressLine1:    req.AddressLine1,
		AddressLine2:    req.AddressLine2,
		City:            req.City,
		State:           req.State,
		PostalCode:      req.PostalCode,
		Country:         req.Country,
	}

	// Llamar al servicio
	customer, err := h.customerService.UpdateCustomer(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.CustomerResponse{
		Id:              customer.PublicID,
		UserId:          safeStringID(customer.UserID),
		Name:            customer.FullName,
		Email:           customer.Email,
		Phone:           safeStringPtr(customer.Phone),
		CompanyName:     safeStringPtr(customer.CompanyName),
		TaxId:           safeStringPtr(customer.TaxID),
		TaxIdType:       safeStringPtr(customer.TaxIDType),
		RequiresInvoice: customer.RequiresInvoice,
		AddressLine1:    safeStringPtr(customer.AddressLine1),
		AddressLine2:    safeStringPtr(customer.AddressLine2),
		City:            safeStringPtr(customer.City),
		State:           safeStringPtr(customer.State),
		PostalCode:      safeStringPtr(customer.PostalCode),
		Country:         safeStringPtr(customer.Country),
		TotalSpent:      customer.TotalSpent,
		TotalOrders:     int32(customer.TotalOrders),
		TotalTickets:    int32(customer.TotalTickets),
		AvgOrderValue:   customer.AvgOrderValue,
		FirstOrderAt:    safeTimeProtoFromString(customer.FirstOrderAt),
		LastOrderAt:     safeTimeProtoFromString(customer.LastOrderAt),
		IsActive:        customer.IsActive,
		IsVip:           customer.IsVIP,
		VipSince:        safeTimeProto(customer.VIPSince),
		CustomerSegment: customer.CustomerSegment,
		LifetimeValue:   customer.LifetimeValue,
		CreatedAt:       timestamppb.New(customer.CreatedAt),
		UpdatedAt:       timestamppb.New(customer.UpdatedAt),
	}, nil
}

func (h *CustomerHandler) ListCustomers(ctx context.Context, req *osmi.ListRequest) (*osmi.CustomerListResponse, error) {
	// Convertir protobuf a DTO de filtro
	filter := dto.CustomerFilter{
		Search:          req.Search,
		Country:         req.Country,
		IsActive:        req.IsActive,
		CustomerSegment: req.Segment,
		DateFrom:        req.DateFrom,
		DateTo:          req.DateTo,
	}

	// Convertir paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	// Llamar al servicio
	customers, total, err := h.customerService.ListCustomers(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir entidades a protobuf
	pbCustomers := make([]*osmi.CustomerResponse, 0, len(customers))
	for _, customer := range customers {
		pbCustomer := &osmi.CustomerResponse{
			Id:              customer.PublicID,
			UserId:          safeStringID(customer.UserID),
			Name:            customer.FullName,
			Email:           customer.Email,
			Phone:           safeStringPtr(customer.Phone),
			CompanyName:     safeStringPtr(customer.CompanyName),
			TaxId:           safeStringPtr(customer.TaxID),
			TaxIdType:       safeStringPtr(customer.TaxIDType),
			RequiresInvoice: customer.RequiresInvoice,
			TotalSpent:      customer.TotalSpent,
			TotalOrders:     int32(customer.TotalOrders),
			TotalTickets:    int32(customer.TotalTickets),
			AvgOrderValue:   customer.AvgOrderValue,
			IsActive:        customer.IsActive,
			IsVip:           customer.IsVIP,
			CustomerSegment: customer.CustomerSegment,
			LifetimeValue:   customer.LifetimeValue,
			CreatedAt:       timestamppb.New(customer.CreatedAt),
		}
		pbCustomers = append(pbCustomers, pbCustomer)
	}

	return &osmi.CustomerListResponse{
		Customers:  pbCustomers,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: int32((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	}, nil
}

func (h *CustomerHandler) GetCustomerStats(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerStatsResponse, error) {
	var customerID string

	// Obtener customerID
	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_Id:
		customerID = lookup.Id
	default:
		return nil, status.Error(codes.InvalidArgument, "customer ID required")
	}

	// Llamar al servicio
	stats, err := h.customerService.GetCustomerStats(ctx, customerID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir estadísticas a protobuf
	pbCountryStats := make([]*osmi.CountryStats, 0, len(stats.TopCountries))
	for _, country := range stats.TopCountries {
		pbCountryStats = append(pbCountryStats, &osmi.CountryStats{
			Country: country.Country,
			Count:   int32(country.Count),
			Revenue: country.Revenue,
		})
	}

	return &osmi.CustomerStatsResponse{
		TotalCustomers:         int32(stats.TotalCustomers),
		ActiveCustomers:        int32(stats.ActiveCustomers),
		VipCustomers:           int32(stats.VIPCustomers),
		NewCustomersLast30Days: int32(stats.NewCustomersLast30Days),
		TotalRevenue:           stats.TotalRevenue,
		AvgLifetimeValue:       stats.AvgLifetimeValue,
		TopCountries:           pbCountryStats,
	}, nil
}

// Helper functions
func safeStringID(id *int64) string {
	if id == nil {
		return ""
	}
	// TODO: Convertir int64 a string (UUID)
	return ""
}

func safeTimeProtoFromString(t *string) *timestamppb.Timestamp {
	if t == nil || *t == "" {
		return nil
	}
	// TODO: Parsear string a time.Time
	return nil
}
