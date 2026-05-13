// internal/application/handlers/grpc/payment_handler.go
package grpc

import (
	"context"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	paymentdto "github.com/franciscozamorau/osmi-server/internal/api/dto/payment"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentHandler struct {
	osmi.UnimplementedOsmiServiceServer
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePayment crea un nuevo pago usando TU DTO existente
func (h *PaymentHandler) CreatePayment(ctx context.Context, req *osmi.CreatePaymentRequest) (*osmi.PaymentProcessingResponse, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}
	if req.Currency == "" {
		req.Currency = "MXN"
	}
	if req.PaymentMethod == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_method is required")
	}
	if req.PaymentProvider == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_provider is required")
	}

	// 🔥 Convertir map[string]string a map[string]interface{}
	paymentMethodDetails := make(map[string]interface{})
	for k, v := range req.PaymentMethodDetails {
		paymentMethodDetails[k] = v
	}

	createReq := &paymentdto.CreatePaymentRequest{
		OrderID:              req.OrderId,
		Amount:               req.Amount,
		Currency:             req.Currency,
		PaymentMethod:        req.PaymentMethod,
		PaymentProvider:      req.PaymentProvider,
		PaymentMethodDetails: paymentMethodDetails,
		SaveCard:             req.SaveCard,
	}

	resp, err := h.paymentService.CreatePayment(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 🔥 Convertir ProviderInstructions a map[string]string
	providerInstructions := make(map[string]string)
	for k, v := range resp.ProviderInstructions {
		if str, ok := v.(string); ok {
			providerInstructions[k] = str
		}
	}

	// 🔥 Manejar ActionURL (posible nil)
	actionURL := ""
	if resp.ActionURL != nil {
		actionURL = *resp.ActionURL
	}

	// 🔥 Manejar ActionType (posible nil)
	actionType := ""
	if resp.ActionType != nil {
		actionType = *resp.ActionType
	}

	// 🔥 Manejar EstimatedCompletion (posible nil)
	var estimatedCompletion *timestamppb.Timestamp
	if resp.EstimatedCompletion != nil {
		estimatedCompletion = timestamppb.New(*resp.EstimatedCompletion)
	}

	return &osmi.PaymentProcessingResponse{
		PaymentId:            resp.PaymentID,
		Status:               resp.Status,
		RequiresAction:       resp.RequiresAction,
		ActionUrl:            actionURL,
		ActionType:           actionType,
		ProviderInstructions: providerInstructions,
		NextSteps:            resp.NextSteps,
		EstimatedCompletion:  estimatedCompletion,
	}, nil
}

// ProcessOrder procesa una orden pagada
func (h *PaymentHandler) ProcessOrder(ctx context.Context, req *osmi.ProcessOrderRequest) (*osmi.Empty, error) {
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	err := h.paymentService.ProcessPaidOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &osmi.Empty{}, nil
}

// CreatePaymentIntent crea un PaymentIntent de Stripe CON reserva temporal de stock
func (h *PaymentHandler) CreatePaymentIntent(
	ctx context.Context,
	req *osmi.CreatePaymentIntentRequest,
) (*osmi.PaymentIntentResponse, error) {

	if req.OrderId == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"order_id is required",
		)
	}

	if req.Currency == "" {
		req.Currency = "MXN"
	}

	createReq := &paymentdto.CreatePaymentIntentRequest{
		OrderID:  req.OrderId,
		Currency: req.Currency,
	}

	resp, err := h.paymentService.CreatePaymentIntent(
		ctx,
		createReq,
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &osmi.PaymentIntentResponse{
		ClientSecret:    resp.ClientSecret,
		PaymentIntentId: resp.PaymentIntentID,
		Amount:          resp.Amount,
		Currency:        resp.Currency,
	}, nil
}
