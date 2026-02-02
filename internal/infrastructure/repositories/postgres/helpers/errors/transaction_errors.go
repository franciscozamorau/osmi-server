package errors

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// TransactionError representa un error de transacción
type TransactionError struct {
	Operation string
	Message   string
	Cause     error
	Retries   int
}

// Error implementa la interfaz error
func (e *TransactionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("transaction error during %s: %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("transaction error during %s: %s", e.Operation, e.Message)
}

// Unwrap devuelve el error original
func (e *TransactionError) Unwrap() error {
	return e.Cause
}

// TransactionManager maneja transacciones
type TransactionManager struct {
	maxRetries   int
	retryDelay   time.Duration
	errorHandler *PostgresErrorHandler
}

// TransactionOptions opciones para transacciones
type TransactionOptions struct {
	IsolationLevel pgx.TxIsoLevel
	AccessMode     pgx.TxAccessMode
	MaxRetries     int
	RetryDelay     time.Duration
	ReadOnly       bool
}

// DefaultTransactionOptions opciones por defecto
var DefaultTransactionOptions = TransactionOptions{
	IsolationLevel: pgx.ReadCommitted,
	AccessMode:     pgx.ReadWrite,
	MaxRetries:     3,
	RetryDelay:     100 * time.Millisecond,
	ReadOnly:       false,
}

// NewTransactionManager crea un nuevo TransactionManager
func NewTransactionManager(errorHandler *PostgresErrorHandler) *TransactionManager {
	return &TransactionManager{
		maxRetries:   3,
		retryDelay:   100 * time.Millisecond,
		errorHandler: errorHandler,
	}
}

// WithRetries configura número de reintentos
func (tm *TransactionManager) WithRetries(maxRetries int) *TransactionManager {
	tm.maxRetries = maxRetries
	return tm
}

// WithRetryDelay configura delay entre reintentos
func (tm *TransactionManager) WithRetryDelay(delay time.Duration) *TransactionManager {
	tm.retryDelay = delay
	return tm
}

// ExecuteInTransaction ejecuta una función dentro de una transacción
func (tm *TransactionManager) ExecuteInTransaction(
	ctx context.Context,
	db interface {
		ExecuteQuery(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	},
	fn func(tx interface{}) error,
	opts ...TransactionOptions,
) error {
	options := DefaultTransactionOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	var lastErr error

	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		if attempt > 0 {
			// Esperar antes de reintentar
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(tm.retryDelay * time.Duration(attempt)):
			}
		}

		// Iniciar transacción
		tx, err := beginTransaction(ctx, db, options)
		if err != nil {
			lastErr = err
			if tm.shouldRetry(err) {
				continue
			}
			return &TransactionError{
				Operation: "begin_transaction",
				Message:   "failed to begin transaction",
				Cause:     err,
				Retries:   attempt,
			}
		}

		// Ejecutar función
		err = fn(tx)
		if err != nil {
			// Rollback en caso de error
			rollbackErr := rollbackTransaction(ctx, tx)
			if rollbackErr != nil {
				lastErr = fmt.Errorf("execution error: %v, rollback error: %v", err, rollbackErr)
			} else {
				lastErr = err
			}

			if tm.shouldRetry(err) {
				continue
			}

			return &TransactionError{
				Operation: "execute",
				Message:   "failed to execute transaction",
				Cause:     lastErr,
				Retries:   attempt,
			}
		}

		// Commit
		err = commitTransaction(ctx, tx)
		if err != nil {
			lastErr = err
			if tm.shouldRetry(err) {
				continue
			}
			return &TransactionError{
				Operation: "commit",
				Message:   "failed to commit transaction",
				Cause:     err,
				Retries:   attempt,
			}
		}

		return nil
	}

	return &TransactionError{
		Operation: "execute",
		Message:   fmt.Sprintf("transaction failed after %d attempts", options.MaxRetries),
		Cause:     lastErr,
		Retries:   options.MaxRetries,
	}
}

// ExecuteReadOnly ejecuta una función de solo lectura
func (tm *TransactionManager) ExecuteReadOnly(
	ctx context.Context,
	db interface{},
	fn func(tx interface{}) error,
) error {
	options := TransactionOptions{
		IsolationLevel: pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
		MaxRetries:     1, // No reintentar para solo lectura
		ReadOnly:       true,
	}

	return tm.ExecuteInTransaction(ctx, db, fn, options)
}

// ExecuteWithIsolation ejecuta con nivel de aislamiento específico
func (tm *TransactionManager) ExecuteWithIsolation(
	ctx context.Context,
	db interface{},
	fn func(tx interface{}) error,
	isolationLevel pgx.TxIsoLevel,
) error {
	options := DefaultTransactionOptions
	options.IsolationLevel = isolationLevel

	return tm.ExecuteInTransaction(ctx, db, fn, options)
}

// shouldRetry determina si se debe reintentar
func (tm *TransactionManager) shouldRetry(err error) bool {
	if tm.errorHandler == nil {
		return false
	}

	return tm.errorHandler.ShouldRetry(err)
}

// beginTransaction inicia una transacción
func beginTransaction(ctx context.Context, db interface{}, opts TransactionOptions) (interface{}, error) {
	// Implementación dependiente del driver de base de datos
	// Esto es un placeholder - en implementación real usarías pgxpool.Pool.BeginTx
	return nil, nil
}

// commitTransaction hace commit de una transacción
func commitTransaction(ctx context.Context, tx interface{}) error {
	// Implementación dependiente del driver
	return nil
}

// rollbackTransaction hace rollback de una transacción
func rollbackTransaction(ctx context.Context, tx interface{}) error {
	// Implementación dependiente del driver
	return nil
}

// TransactionFunc tipo para funciones de transacción
type TransactionFunc func(tx interface{}) error

// NestedTransaction maneja transacciones anidadas
func (tm *TransactionManager) NestedTransaction(
	ctx context.Context,
	tx interface{},
	fn TransactionFunc,
) error {
	// Para PostgreSQL, usamos savepoints para transacciones anidadas
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())

	// Crear savepoint
	if err := createSavepoint(ctx, tx, savepointName); err != nil {
		return &TransactionError{
			Operation: "create_savepoint",
			Message:   "failed to create savepoint",
			Cause:     err,
		}
	}

	// Ejecutar función
	err := fn(tx)
	if err != nil {
		// Rollback al savepoint
		if rollbackErr := rollbackToSavepoint(ctx, tx, savepointName); rollbackErr != nil {
			return &TransactionError{
				Operation: "rollback_savepoint",
				Message:   "failed to rollback to savepoint",
				Cause:     fmt.Errorf("execution error: %v, rollback error: %v", err, rollbackErr),
			}
		}
		return err
	}

	// Liberar savepoint
	if err := releaseSavepoint(ctx, tx, savepointName); err != nil {
		return &TransactionError{
			Operation: "release_savepoint",
			Message:   "failed to release savepoint",
			Cause:     err,
		}
	}

	return nil
}

// createSavepoint crea un savepoint
func createSavepoint(ctx context.Context, tx interface{}, name string) error {
	// Implementación: tx.Exec(ctx, "SAVEPOINT "+name)
	return nil
}

// rollbackToSavepoint hace rollback a un savepoint
func rollbackToSavepoint(ctx context.Context, tx interface{}, name string) error {
	// Implementación: tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+name)
	return nil
}

// releaseSavepoint libera un savepoint
func releaseSavepoint(ctx context.Context, tx interface{}, name string) error {
	// Implementación: tx.Exec(ctx, "RELEASE SAVEPOINT "+name)
	return nil
}

// BatchTransaction ejecuta operaciones en batch
func (tm *TransactionManager) BatchTransaction(
	ctx context.Context,
	db interface{},
	operations []TransactionFunc,
	batchSize int,
) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(operations); i += batchSize {
		end := i + batchSize
		if end > len(operations) {
			end = len(operations)
		}

		batch := operations[i:end]

		err := tm.ExecuteInTransaction(ctx, db, func(tx interface{}) error {
			for _, op := range batch {
				if err := op(tx); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// TransactionStats estadísticas de transacciones
type TransactionStats struct {
	TotalTransactions int64
	Successful        int64
	Failed            int64
	Retried           int64
	AvgDuration       time.Duration
	MaxDuration       time.Duration
}

// TransactionMonitor monitor de transacciones
type TransactionMonitor struct {
	stats TransactionStats
}

// NewTransactionMonitor crea un nuevo monitor
func NewTransactionMonitor() *TransactionMonitor {
	return &TransactionMonitor{}
}

// RecordTransaction registra una transacción
func (tm *TransactionMonitor) RecordTransaction(success bool, duration time.Duration, retries int) {
	// Implementación de monitoreo
}

// GetStats devuelve estadísticas
func (tm *TransactionMonitor) GetStats() TransactionStats {
	return tm.stats
}
