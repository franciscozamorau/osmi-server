// internal/api/grpc/adapter.go
package grpc

import (
	"github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToCreateCustomerRequest convierte protobuf a DTO
func ProtoToCreateCustomerRequest(req *pb.CreateCustomerRequest) *dto.CreateCustomerRequest {
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
		Phone:           SafeStringPtr(customer.Phone),
		CompanyName:     SafeStringPtr(customer.CompanyName),
		AddressLine1:    SafeStringPtr(customer.AddressLine1),
		AddressLine2:    SafeStringPtr(customer.AddressLine2),
		City:            SafeStringPtr(customer.City),
		State:           SafeStringPtr(customer.State),
		PostalCode:      SafeStringPtr(customer.PostalCode),
		Country:         SafeStringPtr(customer.Country),
		TaxId:           SafeStringPtr(customer.TaxID),
		TaxIdType:       SafeStringPtr(customer.TaxIDType),
		TaxName:         SafeStringPtr(customer.TaxName),
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
