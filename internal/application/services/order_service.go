package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	orderdto "github.com/franciscozamorau/osmi-server/internal/api/dto/order"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo      repository.OrderRepository
	customerRepo   repository.CustomerRepository
	ticketTypeRepo repository.TicketTypeRepository
	ticketRepo     repository.TicketRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	customerRepo repository.CustomerRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	ticketRepo repository.TicketRepository,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		customerRepo:   customerRepo,
		ticketTypeRepo: ticketTypeRepo,
		ticketRepo:     ticketRepo,
	}
}

// CreateOrder crea una orden con los items seleccionados
func (s *OrderService) CreateOrder(ctx context.Context, req *orderdto.CreateOrderRequest) (*entities.Order, []*entities.Ticket, error) {
	customer, err := s.customerRepo.GetByPublicID(ctx, req.CustomerID)
	if err != nil {
		return nil, nil, fmt.Errorf("customer not found: %w", err)
	}

	tx, err := s.ticketRepo.BeginTx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var totalAmount float64
	var tickets []*entities.Ticket

	for _, item := range req.Items {
		ticketType, err := s.ticketTypeRepo.FindByPublicID(ctx, item.TicketTypeID)
		if err != nil {
			return nil, nil, fmt.Errorf("ticket type not found: %w", err)
		}

		available, err := s.ticketTypeRepo.CheckAvailability(ctx, ticketType.ID, item.Quantity)
		if err != nil || !available {
			return nil, nil, errors.New("not enough tickets available")
		}

		for i := 0; i < item.Quantity; i++ {
			ticket := &entities.Ticket{
				PublicID:             uuid.New().String(),
				TicketTypeID:         ticketType.ID,
				EventID:              ticketType.EventID,
				CustomerID:           &customer.ID,
				Code:                 fmt.Sprintf("ORD-%d-%d-%s", ticketType.EventID, ticketType.ID, uuid.New().String()[:8]),
				SecretHash:           uuid.New().String(),
				Status:               "reserved",
				FinalPrice:           ticketType.BasePrice,
				Currency:             ticketType.Currency,
				TaxAmount:            ticketType.BasePrice * ticketType.TaxRate,
				ReservedAt:           timePtr(time.Now()),
				ReservationExpiresAt: timePtr(time.Now().Add(15 * time.Minute)),
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}

			err = s.ticketRepo.CreateTx(ctx, tx, ticket)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create ticket: %w", err)
			}

			tickets = append(tickets, ticket)
			totalAmount += ticket.FinalPrice

			err = s.ticketTypeRepo.ReserveTicketsTx(ctx, tx, ticketType.ID, 1)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	paymentMethodStr := ""
	order := &entities.Order{
		CustomerID:       &customer.ID,
		CustomerEmail:    customer.Email,
		CustomerName:     &customer.FullName,
		Subtotal:         totalAmount,
		TaxAmount:        0,
		ServiceFeeAmount: 0,
		DiscountAmount:   0,
		TotalAmount:      totalAmount,
		Currency:         "MXN",
		Status:           "pending",
		OrderType:        "ticket",
		PaymentMethod:    &paymentMethodStr,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, ticket := range tickets {
		ticket.OrderID = &order.ID
		err = s.ticketRepo.UpdateTx(ctx, tx, ticket)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to associate ticket to order: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, tickets, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// generateTicketCode genera un código único para el ticket
func (s *OrderService) generateTicketCode(eventID, ticketTypeID int64, attempt int) string {
	return fmt.Sprintf("ORD-%d-%d-%s", eventID, ticketTypeID, uuid.New().String()[:8])
}
