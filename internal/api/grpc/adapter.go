package grpc

import (
	"time"

	"github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToCreateCustomerRequest convierte protobuf a DTO
func ProtoToCreateCustomerRequest(req *pb.CustomerRequest) *dto.CreateCustomerRequest {
	return &dto.CreateCustomerRequest{
		FullName:        req.Name,
		Email:           req.Email,
		Phone:           req.Phone,
		CompanyName:     req.CompanyName,
		AddressLine1:    req.AddressLine1,
		AddressLine2:    req.AddressLine2,
		City:            req.City,
		State:           req.State,
		PostalCode:      req.PostalCode,
		Country:         req.Country,
		TaxID:           req.TaxId,
		TaxIDType:       req.TaxIdType,
		TaxName:         req.TaxName,
		RequiresInvoice: req.RequiresInvoice,
	}
}

// CustomerToProto convierte entidad Customer a protobuf
func CustomerToProto(customer *entities.Customer) *pb.CustomerResponse {
	return &pb.CustomerResponse{
		Id:              int32(customer.ID),
		PublicId:        customer.PublicID,
		Name:            customer.FullName,
		Email:           customer.Email,
		Phone:           safeStringPtr(customer.Phone),
		CompanyName:     safeStringPtr(customer.CompanyName),
		AddressLine1:    safeStringPtr(customer.AddressLine1),
		AddressLine2:    safeStringPtr(customer.AddressLine2),
		City:            safeStringPtr(customer.City),
		State:           safeStringPtr(customer.State),
		PostalCode:      safeStringPtr(customer.PostalCode),
		Country:         safeStringPtr(customer.Country),
		TaxId:           safeStringPtr(customer.TaxID),
		TaxIdType:       safeStringPtr(customer.TaxIDType),
		TaxName:         safeStringPtr(customer.TaxName),
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
		UpdatedAt:       timestamppb.New(customer.UpdatedAt),
	}
}

// Helper functions
func safeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeTimeProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}
