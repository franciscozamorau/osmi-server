package grpc

import (
	"context"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	orderdto "github.com/franciscozamorau/osmi-server/internal/api/dto/order"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	osmi.UnimplementedOsmiServiceServer
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *osmi.CreateOrderRequest) (*osmi.OrderResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one item is required")
	}

	// 🔥 Convertir proto items a DTO items
	items := make([]orderdto.CreateOrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = orderdto.CreateOrderItemRequest{
			TicketTypeID: item.TicketTypeId,
			Quantity:     int(item.Quantity),
		}
	}

	createReq := &orderdto.CreateOrderRequest{
		CustomerID: req.CustomerId,
		Items:      items,
	}

	order, tickets, err := h.orderService.CreateOrder(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbTickets := make([]*osmi.TicketResponse, len(tickets))
	for i, t := range tickets {
		pbTickets[i] = &osmi.TicketResponse{
			TicketId: t.PublicID,
			Status:   t.Status,
			Price:    t.FinalPrice,
		}
	}

	return &osmi.OrderResponse{
		PublicId:    order.PublicID,
		CustomerId:  req.CustomerId,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
		Tickets:     pbTickets,
		CreatedAt:   timestamppb.New(order.CreatedAt),
	}, nil
}
