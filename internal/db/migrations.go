package db

import (
	"context"
	"fmt"
	"log"
)

// Migration representa una migración de base de datos
type Migration struct {
	ID        int
	Name      string
	Query     string
	AppliedAt string
}

// RunMigrations ejecuta migraciones pendientes
func RunMigrations() error {
	// Crear tabla de migraciones si no existe
	migrationTableQuery := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			query TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := Pool.Exec(context.Background(), migrationTableQuery)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Aquí puedes agregar migraciones futuras
	migrations := []Migration{
		{
			Name:  "initial_schema",
			Query: "-- Migración inicial ya aplicada en CreateTables",
		},
		// Ejemplo de migración futura:
		// {
		// 	Name: "add_user_preferences",
		// 	Query: `
		// 		ALTER TABLE customers
		// 		ADD COLUMN notification_preferences JSONB DEFAULT '{"email": true, "sms": false}'
		// 	`,
		// },
	}

	for _, migration := range migrations {
		if !isMigrationApplied(migration.Name) {
			log.Printf("🔄 Applying migration: %s", migration.Name)

			if migration.Query != "-- Migración inicial ya aplicada en CreateTables" {
				_, err := Pool.Exec(context.Background(), migration.Query)
				if err != nil {
					return fmt.Errorf("error applying migration %s: %w", migration.Name, err)
				}
			}

			err := markMigrationApplied(migration)
			if err != nil {
				return fmt.Errorf("error marking migration as applied: %w", err)
			}

			log.Printf("✅ Migration applied: %s", migration.Name)
		}
	}

	return nil
}

func isMigrationApplied(name string) bool {
	var count int
	err := Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM schema_migrations WHERE name = $1", name).Scan(&count)

	if err != nil {
		return false
	}
	return count > 0
}

func markMigrationApplied(migration Migration) error {
	_, err := Pool.Exec(context.Background(),
		"INSERT INTO schema_migrations (name, query) VALUES ($1, $2)",
		migration.Name, migration.Query)
	return err
}
