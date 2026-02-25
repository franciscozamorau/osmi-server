package errors

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TransactionError representa un error de transacción
type TransactionError struct {
	Operation string
	Message   string
	Cause     error
}

// Error implementa la interfaz error
func (e *TransactionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("transaction error during %s: %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("transaction error during %s: %s", e.Operation, e.Message)
}

// SimpleTransactionManager maneja transacciones de forma simple
type SimpleTransactionManager struct {
	errorHandler *PostgresErrorHandler
}

// NewSimpleTransactionManager crea un nuevo gestor simple
func NewSimpleTransactionManager(errorHandler *PostgresErrorHandler) *SimpleTransactionManager {
	return &SimpleTransactionManager{
		errorHandler: errorHandler,
	}
}

// Execute ejecuta una función con transacción
func (tm *SimpleTransactionManager) Execute(ctx context.Context, pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, fn func(tx pgx.Tx) error) error {
	// Iniciar transacción
	tx, err := pool.Begin(ctx)
	if err != nil {
		return &TransactionError{
			Operation: "begin",
			Message:   "failed to begin transaction",
			Cause:     err,
		}
	}

	// Ejecutar función
	err = fn(tx)
	if err != nil {
		// Rollback en caso de error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return &TransactionError{
				Operation: "rollback",
				Message:   fmt.Sprintf("execution failed: %v, rollback also failed: %v", err, rbErr),
				Cause:     err,
			}
		}
		return &TransactionError{
			Operation: "execute",
			Message:   "transaction failed",
			Cause:     err,
		}
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return &TransactionError{
			Operation: "commit",
			Message:   "failed to commit transaction",
			Cause:     err,
		}
	}

	return nil
}

// ExecuteReadOnly ejecuta una función de solo lectura (sin transacción)
func (tm *SimpleTransactionManager) ExecuteReadOnly(ctx context.Context, pool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, fn func(rows pgx.Rows) error, sql string, args ...interface{}) error {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return &TransactionError{
			Operation: "query",
			Message:   "failed to execute query",
			Cause:     err,
		}
	}
	defer rows.Close()

	return fn(rows)
}

// ShouldRetry determina si se debe reintentar
func (tm *SimpleTransactionManager) ShouldRetry(err error) bool {
	if tm.errorHandler == nil {
		return false
	}

	// Solo reintentar para errores de conexión
	return tm.errorHandler.IsConnectionError(err)
}
