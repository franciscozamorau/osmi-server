package main

import (
	"context"
	"log"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/config"
	"github.com/franciscozamorau/osmi-server/internal/database"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres"
)

func main() {
	cfg := config.Load()

	if err := database.Init(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Inicializar repositorios
	ticketTypeRepo := postgres.NewTicketTypeRepository(database.Pool)
	ticketRepo := postgres.NewTicketRepository(database.Pool)
	eventRepo := postgres.NewEventRepository(database.Pool)
	customerRepo := postgres.NewCustomerRepository(database.Pool)

	ticketService := services.NewTicketService(
		ticketRepo, ticketTypeRepo, eventRepo, customerRepo, nil,
	)

	log.Println("🚀 Worker de expiración de reservas iniciado")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		count, err := ticketService.ReleaseExpiredReservations(ctx)
		if err != nil {
			log.Printf("❌ Error liberando reservas: %v", err)
		} else if count > 0 {
			log.Printf("✅ Liberadas %d reservas expiradas", count)
		}

		cancel()
	}
}
