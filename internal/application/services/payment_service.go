package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/payment"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type PaymentService struct {
	orderRepo      repository.OrderRepository
	ticketRepo     repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	stripeClient   *payment.StripeClient
	webhookSecret  string
}

func NewPaymentService(
	orderRepo repository.OrderRepository,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	stripeClient *payment.StripeClient,
	webhookSecret string,
) *PaymentService {
	return &PaymentService{
		orderRepo:      orderRepo,
		ticketRepo:     ticketRepo,
		ticketTypeRepo: ticketTypeRepo,
		stripeClient:   stripeClient,
		webhookSecret:  webhookSecret,
	}
}

// HandleWebhook - SOLO marca payment_status = "paid" (IDEMPOTENTE)
func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	// Verificar firma del webhook
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	// Solo nos interesan eventos de pago exitoso
	if event.Type != "payment_intent.succeeded" {
		return nil
	}

	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to parse payment intent: %w", err)
	}

	orderID := paymentIntent.Metadata["order_id"]
	if orderID == "" {
		return fmt.Errorf("order_id not found in payment intent metadata")
	}

	// 🔥 IDEMPOTENCIA: Verificar si ya fue procesado
	order, err := s.orderRepo.FindByPublicID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Si ya está paid, no hacer nada (idempotencia)
	if order.PaymentStatus == "paid" {
		return nil
	}

	// 🔥 SOLO actualizar payment_status, NADA MÁS
	order.PaymentStatus = "paid"
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// ProcessPaidOrder - Procesa una orden pagada (lo hace un worker o endpoint interno)
func (s *PaymentService) ProcessPaidOrder(ctx context.Context, orderID string) error {
	// Iniciar transacción
	tx, err := s.ticketRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Bloquear la orden para evitar procesamiento concurrente
	order, err := s.orderRepo.FindByPublicIDForUpdate(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// 🔥 Idempotencia: si ya está completed, salir
	if order.Status == "completed" {
		return tx.Commit(ctx)
	}

	// 🔥 Validar que el pago esté marcado como paid
	if order.PaymentStatus != "paid" {
		return fmt.Errorf("order payment not confirmed yet")
	}

	// Validar que esté pendiente
	if order.Status != "pending" {
		return fmt.Errorf("order cannot be processed, current status: %s", order.Status)
	}

	// Obtener items de la orden
	items, err := s.orderRepo.GetItems(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// Procesar cada ticket
	for _, item := range items {
		ticket, err := s.ticketRepo.GetByID(ctx, item.TicketID)
		if err != nil {
			return fmt.Errorf("ticket not found: %w", err)
		}

		// Validar que el ticket esté reservado
		if ticket.Status != "reserved" {
			continue
		}

		// Convertir ticket a sold
		now := time.Now()
		ticket.Status = "sold"
		ticket.SoldAt = &now
		ticket.ReservedAt = nil
		ticket.ReservationExpiresAt = nil
		ticket.UpdatedAt = now

		err = s.ticketRepo.UpdateTx(ctx, tx, ticket)
		if err != nil {
			return fmt.Errorf("failed to update ticket: %w", err)
		}

		// Confirmar reserva en inventario
		err = s.ticketTypeRepo.ConfirmReservationTx(ctx, tx, ticket.TicketTypeID, 1)
		if err != nil {
			return fmt.Errorf("failed to confirm reservation: %w", err)
		}
	}

	// Marcar orden como completed
	order.Status = "completed"
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return tx.Commit(ctx)
}
