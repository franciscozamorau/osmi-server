package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	paymentdto "github.com/franciscozamorau/osmi-server/internal/api/dto/payment"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/payment"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type PaymentService struct {
	paymentRepo    repository.PaymentRepository
	orderRepo      repository.OrderRepository
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	stripeClient   *payment.StripeClient
	webhookSecret  string
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	stripeClient *payment.StripeClient,
	webhookSecret string,
) *PaymentService {
	return &PaymentService{
		paymentRepo:    paymentRepo,
		orderRepo:      orderRepo,
		ticketRepo:     ticketRepo,
		ticketTypeRepo: ticketTypeRepo,
		stripeClient:   stripeClient,
		webhookSecret:  webhookSecret,
	}
}

// CreatePayment crea un nuevo pago usando TU DTO y devuelve TU DTO de respuesta
func (s *PaymentService) CreatePayment(ctx context.Context, req *paymentdto.CreatePaymentRequest) (*paymentdto.PaymentProcessingResponse, error) {
	// 1. Obtener la orden
	order, err := s.orderRepo.FindByPublicID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// 2. Validar que la orden esté pendiente
	if order.Status != "pending" {
		return nil, fmt.Errorf("order is not pending, current status: %s", order.Status)
	}

	// 3. Mapear proveedor (Stripe = 1 por ahora)
	providerID := int16(1)

	// 4. Crear entidad Payment
	now := time.Now()
	payment := &entities.Payment{
		OrderID:       order.ID,
		ProviderID:    providerID,
		Amount:        order.TotalAmount,
		Currency:      req.Currency,
		ExchangeRate:  1.0,
		Status:        "pending",
		PaymentMethod: &req.PaymentMethod,
		Attempts:      0,
		MaxAttempts:   3,
		IPAddress:     nil,
		UserAgent:     nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := payment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid payment: %w", err)
	}

	// 5. Guardar en BD
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 6. Crear PaymentIntent en Stripe
	amountCents := int64(order.TotalAmount * 100)
	pi, err := s.stripeClient.CreatePaymentIntent(amountCents, req.Currency, order.PublicID)
	if err != nil {
		payment.Status = "failed"
		_ = s.paymentRepo.Update(ctx, payment)
		return nil, fmt.Errorf("failed to create Stripe payment intent: %w", err)
	}

	// 7. Actualizar payment con datos de Stripe
	payment.ProviderTransactionID = &pi.ID
	payment.Status = "processing"

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment with Stripe data: %w", err)
	}

	// 8. Devolver respuesta con client_secret para el frontend
	paymentID := fmt.Sprintf("%d", payment.ID)
	return &paymentdto.PaymentProcessingResponse{
		PaymentID:      paymentID,
		Status:         payment.Status,
		RequiresAction: true,
		ActionType:     strPtr("stripe_sdk"),
		ProviderInstructions: map[string]interface{}{
			"client_secret":     pi.ClientSecret,
			"payment_intent_id": pi.ID,
		},
	}, nil
}

// Helper para crear punteros a string
func strPtr(s string) *string {
	return &s
}

// GetPayment obtiene un pago por ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*entities.Payment, error) {
	var id int64
	if _, err := fmt.Sscanf(paymentID, "%d", &id); err == nil {
		return s.paymentRepo.FindByID(ctx, id)
	}
	return s.paymentRepo.FindByTransactionID(ctx, paymentID)
}

// HandleWebhook - SOLO marca payment_status = "paid" (IDEMPOTENTE)
func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	if event.Type != "payment_intent.succeeded" {
		return nil
	}

	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	// Buscar payment por transaction_id
	payment, err := s.paymentRepo.FindByTransactionID(ctx, paymentIntent.ID)
	if err != nil {
		return fmt.Errorf("payment not found for transaction: %s", paymentIntent.ID)
	}

	// Idempotencia: si ya está completed o refunded, salir
	if payment.Status == "completed" || payment.Status == "refunded" {
		return nil
	}

	// Actualizar payment
	payment.Status = "completed"
	now := time.Now()
	payment.ProcessedAt = &now

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Actualizar orden (marcar payment_status = paid)
	order, err := s.orderRepo.FindByID(ctx, payment.OrderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	order.PaymentStatus = "paid"
	order.UpdatedAt = now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order payment status: %w", err)
	}

	return nil
}

// ProcessPaidOrder - Procesa una orden pagada (lo hace un worker o endpoint interno)
func (s *PaymentService) ProcessPaidOrder(ctx context.Context, orderID string) error {
	tx, err := s.ticketRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	order, err := s.orderRepo.FindByPublicIDForUpdate(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.Status == "completed" {
		return tx.Commit(ctx)
	}

	if order.PaymentStatus != "paid" {
		return fmt.Errorf("order payment not confirmed yet")
	}

	if order.Status != "pending" {
		return fmt.Errorf("order cannot be processed, current status: %s", order.Status)
	}

	items, err := s.orderRepo.GetItems(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	for _, item := range items {
		ticket, err := s.ticketRepo.GetByID(ctx, item.TicketID)
		if err != nil {
			return fmt.Errorf("ticket not found: %w", err)
		}

		if ticket.Status != "reserved" {
			continue
		}

		now := time.Now()
		ticket.Status = "sold"
		ticket.SoldAt = &now
		ticket.ReservedAt = nil
		ticket.ReservationExpiresAt = nil
		ticket.UpdatedAt = now

		if err := s.ticketRepo.UpdateTx(ctx, tx, ticket); err != nil {
			return fmt.Errorf("failed to update ticket: %w", err)
		}

		if err := s.ticketTypeRepo.ConfirmReservationTx(ctx, tx, ticket.TicketTypeID, 1); err != nil {
			return fmt.Errorf("failed to confirm reservation: %w", err)
		}
	}

	order.Status = "completed"
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return tx.Commit(ctx)
}

// CreatePaymentIntent crea un PaymentIntent de Stripe para el frontend
func (s *PaymentService) CreatePaymentIntent(
	ctx context.Context,
	req *paymentdto.CreatePaymentIntentRequest,
) (*paymentdto.CreatePaymentIntentResponse, error) {

	order, err := s.orderRepo.FindByPublicID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Debe seguir pendiente
	if order.Status != "pending" {
		return nil, fmt.Errorf(
			"order is not pending, current status: %s",
			order.Status,
		)
	}

	// La reserva creada por CreateOrder debe seguir viva
	if order.PaymentStatus == "paid" {
		return nil, fmt.Errorf("order already paid")
	}

	currency := req.Currency
	if currency == "" {
		currency = "MXN"
	}

	amountCents := int64(order.TotalAmount * 100)

	pi, err := s.stripeClient.CreatePaymentIntent(
		amountCents,
		currency,
		order.PublicID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create stripe payment intent: %w",
			err,
		)
	}

	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf(
			"failed to update order: %w",
			err,
		)
	}

	return &paymentdto.CreatePaymentIntentResponse{
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
		Amount:          pi.Amount,
		Currency:        string(pi.Currency),
	}, nil
}
