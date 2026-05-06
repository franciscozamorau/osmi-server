// cmd/worker/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/database"
)

const (
	workerInterval = 5 * time.Minute
	queryTimeout   = 2 * time.Minute
)

func main() {
	log.Println("🚀 OSMI Reservation Expiration Worker")
	log.Println("======================================")
	log.Printf("⏱️ Intervalo de ejecución: %s", workerInterval)

	if err := database.Init(); err != nil {
		log.Fatalf("❌ Failed to initialize database connection: %v", err)
	}
	defer database.Close()

	log.Println("✅ Database connected")

	// Primera ejecución inmediata al iniciar
	executeExpirationJob()

	// Ejecución recurrente
	ticker := time.NewTicker(workerInterval)
	defer ticker.Stop()

	for range ticker.C {
		executeExpirationJob()
	}
}

func executeExpirationJob() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	log.Println("🔄 Ejecutando limpieza de reservas expiradas...")

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		log.Printf("❌ Failed to start transaction: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	expiredCount, err := expireReservedTickets(ctx, tx)
	if err != nil {
		log.Printf("❌ Failed to expire reserved tickets: %v", err)
		return
	}

	if expiredCount == 0 {
		if err := tx.Commit(ctx); err != nil {
			log.Printf("❌ Failed to commit empty transaction: %v", err)
			return
		}

		log.Println("📭 No expired reservations found")
		return
	}

	if err := recalculateTicketTypeCounters(ctx, tx); err != nil {
		log.Printf("❌ Failed to recalculate ticket counters: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("❌ Failed to commit transaction: %v", err)
		return
	}

	log.Printf(
		"✅ Expired reservations processed successfully | released=%d | duration=%s",
		expiredCount,
		time.Since(start),
	)
}

func expireReservedTickets(
	ctx context.Context,
	tx interface {
		Exec(context.Context, string, ...interface{}) (interface {
			RowsAffected() int64
		}, error)
	},
) (int64, error) {
	const query = `
		UPDATE ticketing.tickets
		SET
			status = 'expired',
			reservation_expires_at = NULL,
			updated_at = NOW()
		WHERE status = 'reserved'
		  AND reservation_expires_at IS NOT NULL
		  AND reservation_expires_at < NOW()
	`

	result, err := tx.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func recalculateTicketTypeCounters(
	ctx context.Context,
	tx interface {
		Exec(context.Context, string, ...interface{}) (interface {
			RowsAffected() int64
		}, error)
	},
) error {
	const query = `
		UPDATE ticketing.ticket_types tt
		SET
			reserved_quantity = COALESCE(calc.real_reserved, 0),
			sold_quantity = COALESCE(calc.real_sold, 0),
			updated_at = NOW()
		FROM (
			SELECT
				ticket_type_id,
				COUNT(*) FILTER (
					WHERE status = 'reserved'
				) AS real_reserved,
				COUNT(*) FILTER (
					WHERE status IN ('sold', 'checked_in')
				) AS real_sold
			FROM ticketing.tickets
			GROUP BY ticket_type_id
		) calc
		WHERE tt.id = calc.ticket_type_id
	`

	_, err := tx.Exec(ctx, query)
	return err
}
