package db

import (
	"context"
	"fmt"
	"log"
)

// SeedData inserta datos de prueba para desarrollo
func SeedData() error {
	log.Println("🌱 Seeding database with test data...")

	// Verificar si ya existen datos
	var customerCount int
	err := Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM customers").Scan(&customerCount)
	if err != nil {
		return fmt.Errorf("error checking existing data: %w", err)
	}

	if customerCount > 0 {
		log.Println("✅ Database already has data, skipping seed")
		return nil
	}

	// Insertar datos de prueba
	queries := []string{
		// Insertar eventos
		`
		INSERT INTO events (name, description, start_date, end_date, location, category, is_published)
		VALUES 
		('Concierto de Rock', 'Un increíble concierto de rock con las mejores bandas', 
		 CURRENT_TIMESTAMP + INTERVAL '10 days', CURRENT_TIMESTAMP + INTERVAL '11 days',
		 'Estadio Nacional', 'Música', true),
		('Conferencia Tech', 'La mayor conferencia de tecnología del año',
		 CURRENT_TIMESTAMP + INTERVAL '15 days', CURRENT_TIMESTAMP + INTERVAL '17 days', 
		 'Centro de Convenciones', 'Tecnología', true)
		`,

		// Insertar clientes
		`
		INSERT INTO customers (name, email, phone, is_verified)
		VALUES 
		('Juan Pérez', 'juan@example.com', '+1234567890', true),
		('María García', 'maria@example.com', '+0987654321', true),
		('Carlos López', 'carlos@example.com', '+1122334455', false)
		`,
	}

	for _, query := range queries {
		_, err := Pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("error seeding data: %w\nQuery: %s", err, query)
		}
	}

	log.Println("✅ Test data seeded successfully")
	return nil
}
