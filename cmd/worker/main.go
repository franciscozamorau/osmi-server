package main

import (
	"context"
	"log"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/database"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("🚀 Worker de expiración de reservas iniciado")
	log.Println("⏱️  Se ejecutará cada 5 minutos")

	// Ejecutar inmediatamente al iniciar
	runExpiration()

	// Luego cada 5 minutos
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		runExpiration()
	}
}

func runExpiration() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Marcar reservas expiradas como 'expired'
	updateTicketsQuery := `
		UPDATE ticketing.tickets 
		SET status = 'expired', 
		    reservation_expires_at = NULL,
		    updated_at = NOW()
		WHERE status = 'reserved' 
		  AND reservation_expires_at < NOW()
	`
	result, err := database.Pool.Exec(ctx, updateTicketsQuery)
	if err != nil {
		log.Printf("❌ Error marcando tickets expirados: %v", err)
		return
	}

	count := result.RowsAffected()

	if count > 0 {
		log.Printf("📝 Marcados %d tickets como expired", count)

		// 2. Recalcular contadores
		recalcQuery := `
			UPDATE ticketing.ticket_types tt
			SET 
			    reserved_quantity = COALESCE(r.real_reserved, 0),
			    sold_quantity = COALESCE(r.real_sold, 0)
			FROM (
			    SELECT 
			        ticket_type_id,
			        COUNT(*) FILTER (WHERE status = 'reserved') AS real_reserved,
			        COUNT(*) FILTER (WHERE status IN ('sold', 'checked_in')) AS real_sold
			    FROM ticketing.tickets
			    GROUP BY ticket_type_id
			) r
			WHERE tt.id = r.ticket_type_id
		`
		_, err = database.Pool.Exec(ctx, recalcQuery)
		if err != nil {
			log.Printf("❌ Error recalculando contadores: %v", err)
		} else {
			log.Printf("✅ Liberadas %d reservas expiradas (contadores actualizados)", count)
		}
	}
}
