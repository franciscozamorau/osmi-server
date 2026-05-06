package postgres

import (
	"context"
	"errors"
	"time"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	paymentdto "github.com/franciscozamorau/osmi-server/internal/api/dto/payment"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create inserta un nuevo pago
func (r *PaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	query := `
		INSERT INTO billing.payments (
			order_id, provider_id, amount, currency, exchange_rate,
			status, payment_method, attempts, max_attempts,
			ip_address, user_agent, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		payment.OrderID, payment.ProviderID, payment.Amount, payment.Currency, payment.ExchangeRate,
		payment.Status, payment.PaymentMethod, payment.Attempts, payment.MaxAttempts,
		payment.IPAddress, payment.UserAgent,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

	return err
}

// FindByID obtiene un pago por ID
func (r *PaymentRepository) FindByID(ctx context.Context, id int64) (*entities.Payment, error) {
	query := `
		SELECT id, order_id, provider_id, provider_transaction_id, provider_session_id,
			amount, currency, exchange_rate, status, payment_method, payment_method_details,
			attempts, max_attempts, next_retry_at, last_error, error_code,
			ip_address, user_agent, processed_at, refunded_at, cancelled_at,
			created_at, updated_at
		FROM billing.payments
		WHERE id = $1
	`

	var p entities.Payment
	var providerTransactionID, providerSessionID, lastError, errorCode *string
	var paymentMethodDetails map[string]interface{}
	var nextRetryAt, processedAt, refundedAt, cancelledAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.OrderID, &p.ProviderID, &providerTransactionID, &providerSessionID,
		&p.Amount, &p.Currency, &p.ExchangeRate, &p.Status, &p.PaymentMethod, &paymentMethodDetails,
		&p.Attempts, &p.MaxAttempts, &nextRetryAt, &lastError, &errorCode,
		&p.IPAddress, &p.UserAgent, &processedAt, &refundedAt, &cancelledAt,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	p.ProviderTransactionID = providerTransactionID
	p.ProviderSessionID = providerSessionID
	p.PaymentMethodDetails = &paymentMethodDetails
	p.NextRetryAt = nextRetryAt
	p.LastError = lastError
	p.ErrorCode = errorCode
	p.ProcessedAt = processedAt
	p.RefundedAt = refundedAt
	p.CancelledAt = cancelledAt

	return &p, nil
}

// FindByOrderID obtiene pagos de una orden
func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID int64) ([]*entities.Payment, error) {
	query := `
		SELECT id, order_id, provider_id, provider_transaction_id, provider_session_id,
			amount, currency, exchange_rate, status, payment_method, payment_method_details,
			attempts, max_attempts, next_retry_at, last_error, error_code,
			ip_address, user_agent, processed_at, refunded_at, cancelled_at,
			created_at, updated_at
		FROM billing.payments
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*entities.Payment
	for rows.Next() {
		var p entities.Payment
		var providerTransactionID, providerSessionID, lastError, errorCode *string
		var paymentMethodDetails map[string]interface{}
		var nextRetryAt, processedAt, refundedAt, cancelledAt *time.Time

		err = rows.Scan(
			&p.ID, &p.OrderID, &p.ProviderID, &providerTransactionID, &providerSessionID,
			&p.Amount, &p.Currency, &p.ExchangeRate, &p.Status, &p.PaymentMethod, &paymentMethodDetails,
			&p.Attempts, &p.MaxAttempts, &nextRetryAt, &lastError, &errorCode,
			&p.IPAddress, &p.UserAgent, &processedAt, &refundedAt, &cancelledAt,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		p.ProviderTransactionID = providerTransactionID
		p.ProviderSessionID = providerSessionID
		p.PaymentMethodDetails = &paymentMethodDetails
		p.NextRetryAt = nextRetryAt
		p.LastError = lastError
		p.ErrorCode = errorCode
		p.ProcessedAt = processedAt
		p.RefundedAt = refundedAt
		p.CancelledAt = cancelledAt

		payments = append(payments, &p)
	}

	return payments, nil
}

// UpdateStatus actualiza el estado de un pago
func (r *PaymentRepository) UpdateStatus(ctx context.Context, paymentID int64, status string, providerData map[string]interface{}) error {
	query := `
		UPDATE billing.payments
		SET status = $1,
		    provider_transaction_id = COALESCE($2, provider_transaction_id),
		    provider_session_id = COALESCE($3, provider_session_id),
		    updated_at = NOW()
		WHERE id = $4
	`

	var transactionID, sessionID *string
	if providerData != nil {
		if tid, ok := providerData["transaction_id"].(string); ok {
			transactionID = &tid
		}
		if sid, ok := providerData["session_id"].(string); ok {
			sessionID = &sid
		}
	}

	_, err := r.db.Exec(ctx, query, status, transactionID, sessionID, paymentID)
	return err
}

// Update actualiza un pago completo
func (r *PaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	query := `
		UPDATE billing.payments SET
			provider_transaction_id = $1,
			provider_session_id = $2,
			status = $3,
			attempts = $4,
			next_retry_at = $5,
			last_error = $6,
			error_code = $7,
			processed_at = $8,
			refunded_at = $9,
			cancelled_at = $10,
			updated_at = NOW()
		WHERE id = $11
	`

	_, err := r.db.Exec(ctx, query,
		payment.ProviderTransactionID, payment.ProviderSessionID,
		payment.Status, payment.Attempts, payment.NextRetryAt,
		payment.LastError, payment.ErrorCode,
		payment.ProcessedAt, payment.RefundedAt, payment.CancelledAt,
		payment.ID,
	)

	return err
}

// FindByTransactionID obtiene un pago por transaction_id del proveedor
func (r *PaymentRepository) FindByTransactionID(ctx context.Context, transactionID string) (*entities.Payment, error) {
	query := `
		SELECT id, order_id, provider_id, provider_transaction_id, provider_session_id,
			amount, currency, exchange_rate, status, payment_method, payment_method_details,
			attempts, max_attempts, next_retry_at, last_error, error_code,
			ip_address, user_agent, processed_at, refunded_at, cancelled_at,
			created_at, updated_at
		FROM billing.payments
		WHERE provider_transaction_id = $1
	`

	var p entities.Payment
	var providerSessionID, lastError, errorCode *string
	var paymentMethodDetails map[string]interface{}
	var nextRetryAt, processedAt, refundedAt, cancelledAt *time.Time

	err := r.db.QueryRow(ctx, query, transactionID).Scan(
		&p.ID, &p.OrderID, &p.ProviderID, &p.ProviderTransactionID, &providerSessionID,
		&p.Amount, &p.Currency, &p.ExchangeRate, &p.Status, &p.PaymentMethod, &paymentMethodDetails,
		&p.Attempts, &p.MaxAttempts, &nextRetryAt, &lastError, &errorCode,
		&p.IPAddress, &p.UserAgent, &processedAt, &refundedAt, &cancelledAt,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repository.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	p.ProviderSessionID = providerSessionID
	p.PaymentMethodDetails = &paymentMethodDetails
	p.NextRetryAt = nextRetryAt
	p.LastError = lastError
	p.ErrorCode = errorCode
	p.ProcessedAt = processedAt
	p.RefundedAt = refundedAt
	p.CancelledAt = cancelledAt

	return &p, nil
}

// FindByOrder obtiene todos los pagos de una orden
func (r *PaymentRepository) FindByOrder(ctx context.Context, orderID int64) ([]*entities.Payment, error) {
	query := `
        SELECT id, order_id, provider_id, provider_transaction_id, provider_session_id,
            amount, currency, exchange_rate, status, payment_method, payment_method_details,
            attempts, max_attempts, next_retry_at, last_error, error_code,
            ip_address, user_agent, processed_at, refunded_at, cancelled_at,
            created_at, updated_at
        FROM billing.payments
        WHERE order_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*entities.Payment
	for rows.Next() {
		var p entities.Payment
		var providerTransactionID, providerSessionID, lastError, errorCode *string
		var paymentMethodDetails map[string]interface{}
		var nextRetryAt, processedAt, refundedAt, cancelledAt *time.Time

		err = rows.Scan(
			&p.ID, &p.OrderID, &p.ProviderID, &providerTransactionID, &providerSessionID,
			&p.Amount, &p.Currency, &p.ExchangeRate, &p.Status, &p.PaymentMethod, &paymentMethodDetails,
			&p.Attempts, &p.MaxAttempts, &nextRetryAt, &lastError, &errorCode,
			&p.IPAddress, &p.UserAgent, &processedAt, &refundedAt, &cancelledAt,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		p.ProviderTransactionID = providerTransactionID
		p.ProviderSessionID = providerSessionID
		p.PaymentMethodDetails = &paymentMethodDetails
		p.NextRetryAt = nextRetryAt
		p.LastError = lastError
		p.ErrorCode = errorCode
		p.ProcessedAt = processedAt
		p.RefundedAt = refundedAt
		p.CancelledAt = cancelledAt

		payments = append(payments, &p)
	}

	return payments, nil
}

// ============================================================================
// MÉTODOS REQUERIDOS POR LA INTERFAZ (STUBS - PENDIENTES DE IMPLEMENTAR)
// ============================================================================

func (r *PaymentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM billing.payments WHERE id = $1`, id)
	return err
}

func (r *PaymentRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Payment, error) {
	// Si tu tabla no tiene public_uuid, puedes omitir este método o implementarlo
	return nil, repository.ErrPaymentNotFound
}

func (r *PaymentRepository) List(ctx context.Context, filter paymentdto.PaymentFilter, pagination commondto.Pagination) ([]*entities.Payment, int64, error) {
	return nil, 0, nil
}

func (r *PaymentRepository) FindByCustomer(ctx context.Context, customerID int64, pagination commondto.Pagination) ([]*entities.Payment, int64, error) {
	return nil, 0, nil
}

func (r *PaymentRepository) FindByStatus(ctx context.Context, status string, pagination commondto.Pagination) ([]*entities.Payment, int64, error) {
	return nil, 0, nil
}

func (r *PaymentRepository) FindByProvider(ctx context.Context, providerID int64, pagination commondto.Pagination) ([]*entities.Payment, int64, error) {
	return nil, 0, nil
}

func (r *PaymentRepository) FindFailedPayments(ctx context.Context, hours int) ([]*entities.Payment, error) {
	return nil, nil
}

func (r *PaymentRepository) FindPendingPayments(ctx context.Context) ([]*entities.Payment, error) {
	return nil, nil
}

func (r *PaymentRepository) MarkAsProcessed(ctx context.Context, paymentID int64, processedAt string) error {
	return nil
}

func (r *PaymentRepository) MarkAsRefunded(ctx context.Context, paymentID int64, refundID int64) error {
	return nil
}

func (r *PaymentRepository) MarkAsFailed(ctx context.Context, paymentID int64, errorMessage string, errorCode string) error {
	return nil
}

func (r *PaymentRepository) IncrementAttempts(ctx context.Context, paymentID int64) error {
	return nil
}

func (r *PaymentRepository) SetNextRetry(ctx context.Context, paymentID int64, nextRetryAt string) error {
	return nil
}

func (r *PaymentRepository) RecordProviderResponse(ctx context.Context, paymentID int64, response map[string]interface{}) error {
	return nil
}

func (r *PaymentRepository) UpdatePaymentMethod(ctx context.Context, paymentID int64, method string, details map[string]interface{}) error {
	return nil
}

func (r *PaymentRepository) GetStats(ctx context.Context, filter paymentdto.PaymentFilter) (*paymentdto.PaymentStatsResponse, error) {
	return nil, nil
}

func (r *PaymentRepository) GetProviderStats(ctx context.Context, providerID int64) (*paymentdto.ProviderStats, error) {
	return nil, nil
}

func (r *PaymentRepository) GetDailyPaymentVolume(ctx context.Context, days int) ([]*paymentdto.DailyVolume, error) {
	return nil, nil
}

func (r *PaymentRepository) GetSuccessRate(ctx context.Context, providerID *int64) (float64, error) {
	return 0, nil
}

func (r *PaymentRepository) GetAverageProcessingTime(ctx context.Context) (float64, error) {
	return 0, nil
}

func (r *PaymentRepository) GetTotalProcessedAmount(ctx context.Context, currency string) (float64, error) {
	return 0, nil
}
