package postgres

import (
	"context"
	"errors"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	orderdto "github.com/franciscozamorau/osmi-server/internal/api/dto/order"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// ============================================================================
// MÉTODOS BASE (IMPLEMENTADOS)
// ============================================================================

func (r *OrderRepository) Create(ctx context.Context, order *entities.Order) error {
	query := `
		INSERT INTO billing.orders (
			public_uuid, customer_id, customer_email, customer_name, customer_phone,
			subtotal, tax_amount, service_fee_amount, discount_amount, total_amount, currency,
			status, order_type, is_reservation, reservation_expires_at,
			payment_method, payment_provider_id,
			invoice_required, invoice_generated, invoice_number,
			promotion_code, promotion_id, metadata, notes,
			ip_address, user_agent,
			expires_at, paid_at, cancelled_at, refunded_at,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16,
			$17, $18, $19,
			$20, $21, $22, $23,
			$24, $25,
			$26, $27, $28, $29,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		order.CustomerID, order.CustomerEmail, order.CustomerName, order.CustomerPhone,
		order.Subtotal, order.TaxAmount, order.ServiceFeeAmount, order.DiscountAmount, order.TotalAmount, order.Currency,
		order.Status, order.OrderType, order.IsReservation, order.ReservationExpiresAt,
		order.PaymentMethod, order.PaymentProviderID,
		order.InvoiceRequired, order.InvoiceGenerated, order.InvoiceNumber,
		order.PromotionCode, order.PromotionID, order.Metadata, order.Notes,
		order.IPAddress, order.UserAgent,
		order.ExpiresAt, order.PaidAt, order.CancelledAt, order.RefundedAt,
	).Scan(&order.ID, &order.PublicID, &order.CreatedAt, &order.UpdatedAt)

	return err
}

func (r *OrderRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Order, error) {
	query := `
		SELECT id, public_uuid, customer_id, status, total_amount, currency,
			payment_method, created_at, updated_at
		FROM billing.orders
		WHERE public_uuid = $1
	`

	var order entities.Order
	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&order.ID, &order.PublicID, &order.CustomerID, &order.Status,
		&order.TotalAmount, &order.Currency, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrOrderNotFound
	}
	return &order, err
}

func (r *OrderRepository) GetByCustomerID(ctx context.Context, customerID int64) ([]*entities.Order, error) {
	query := `
		SELECT id, public_uuid, customer_id, status, total_amount, currency,
			payment_method, created_at, updated_at
		FROM billing.orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entities.Order
	for rows.Next() {
		var order entities.Order
		err = rows.Scan(
			&order.ID, &order.PublicID, &order.CustomerID, &order.Status,
			&order.TotalAmount, &order.Currency, &order.PaymentMethod,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	query := `UPDATE billing.orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, orderID)
	return err
}

func (r *OrderRepository) AddItem(ctx context.Context, item *entities.OrderItem) error {
	query := `
		INSERT INTO billing.order_items (
			order_id, ticket_type_id, ticket_id, quantity, unit_price, total_price
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query,
		item.OrderID, item.TicketTypeID, item.TicketID, item.Quantity,
		item.UnitPrice, item.TotalPrice,
	).Scan(&item.ID)
}

func (r *OrderRepository) GetItems(ctx context.Context, orderID int64) ([]*entities.OrderItem, error) {
	query := `
		SELECT id, order_id, ticket_type_id, ticket_id, quantity, unit_price, total_price
		FROM billing.order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entities.OrderItem
	for rows.Next() {
		var item entities.OrderItem
		err = rows.Scan(
			&item.ID, &item.OrderID, &item.TicketTypeID, &item.TicketID,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// ============================================================================
// MÉTODOS REQUERIDOS POR LA INTERFAZ (STUBS - SIN DUPLICADOS)
// ============================================================================

func (r *OrderRepository) FindByID(ctx context.Context, id int64) (*entities.Order, error) {
	return nil, repository.ErrOrderNotFound
}

func (r *OrderRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Order, error) {
	return r.GetByPublicID(ctx, publicID)
}

func (r *OrderRepository) Update(ctx context.Context, order *entities.Order) error {
	query := `
        UPDATE billing.orders SET
            status = $1,
            total_amount = $2,
            updated_at = NOW()
        WHERE public_uuid = $3
    `
	_, err := r.db.Exec(ctx, query, order.Status, order.TotalAmount, order.PublicID)
	return err
}

func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM billing.orders WHERE id = $1`, id)
	return err
}

func (r *OrderRepository) List(ctx context.Context, filter orderdto.OrderFilter, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) FindByCustomer(ctx context.Context, customerID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) FindByStatus(ctx context.Context, status string, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) FindByEvent(ctx context.Context, eventID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) FindByPaymentProvider(ctx context.Context, providerID int64, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) FindExpiredReservations(ctx context.Context) ([]*entities.Order, error) {
	return nil, nil
}

func (r *OrderRepository) Search(ctx context.Context, term string, filter orderdto.OrderFilter, pagination commondto.Pagination) ([]*entities.Order, int64, error) {
	return nil, 0, nil
}

func (r *OrderRepository) MarkAsPaid(ctx context.Context, orderID int64, paymentID int64, paidAt string) error {
	query := `UPDATE billing.orders SET status = 'completed', paid_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, orderID)
	return err
}

func (r *OrderRepository) MarkAsCancelled(ctx context.Context, orderID int64, reason string) error {
	query := `UPDATE billing.orders SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, orderID)
	return err
}

func (r *OrderRepository) MarkAsRefunded(ctx context.Context, orderID int64, refundID int64) error {
	query := `UPDATE billing.orders SET status = 'refunded', refunded_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, orderID)
	return err
}

func (r *OrderRepository) AddOrderItem(ctx context.Context, orderID int64, item *entities.OrderItem) error {
	item.OrderID = orderID
	return r.AddItem(ctx, item)
}

func (r *OrderRepository) UpdateOrderItems(ctx context.Context, orderID int64, items []*entities.OrderItem) error {
	return nil
}

func (r *OrderRepository) CalculateTotals(ctx context.Context, orderID int64) (*orderdto.OrderTotals, error) {
	return nil, nil
}

func (r *OrderRepository) ApplyPromotion(ctx context.Context, orderID int64, promotionCode string) error {
	return nil
}

func (r *OrderRepository) RemovePromotion(ctx context.Context, orderID int64) error {
	return nil
}

func (r *OrderRepository) GenerateInvoice(ctx context.Context, orderID int64) (string, error) {
	return "", nil
}

func (r *OrderRepository) CancelInvoice(ctx context.Context, orderID int64) error {
	return nil
}

func (r *OrderRepository) GetStats(ctx context.Context, filter orderdto.OrderFilter) (*orderdto.OrderStatsResponse, error) {
	return nil, nil
}

func (r *OrderRepository) GetCustomerOrderStats(ctx context.Context, customerID int64) (*orderdto.CustomerOrderStats, error) {
	return nil, nil
}

func (r *OrderRepository) GetEventOrderStats(ctx context.Context, eventID int64) (*orderdto.EventOrderStats, error) {
	return nil, nil
}

func (r *OrderRepository) GetDailyRevenue(ctx context.Context, days int) ([]*orderdto.DailyRevenue, error) {
	return nil, nil
}

func (r *OrderRepository) GetAverageOrderValue(ctx context.Context) (float64, error) {
	return 0, nil
}

func (r *OrderRepository) GetConversionRate(ctx context.Context) (float64, error) {
	return 0, nil
}
