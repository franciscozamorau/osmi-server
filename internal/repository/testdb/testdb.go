// internal/repository/testdb/testdb.go
package testdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDB representa una base de datos de prueba
type TestDB struct {
	DB *sqlx.DB
}

// New crea una nueva conexión a la base de datos de prueba
// Usa la misma estructura que tu connection.go pero con base de datos "osmidb_test"
func New(t *testing.T) (*TestDB, func()) {
	t.Helper()

	// Configuración - usar base de datos de prueba separada
	connStr := "host=localhost port=5432 user=osmi password=osmi1405 dbname=osmidb_test sslmode=disable"

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Configurar pool como en producción
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Limpiar tablas antes de empezar (opcional, depende de tu flujo)
	cleanup := func() {
		// Aquí puedes limpiar datos si es necesario
		// Por ahora, simplemente cerramos la conexión
		db.Close()
	}

	return &TestDB{DB: db}, cleanup
}

// TruncateTables limpia todas las tablas después de cada test
func (tdb *TestDB) TruncateTables(t *testing.T) {
	t.Helper()

	tables := []string{
		"ticketing.tickets",
		"ticketing.ticket_types",
		"ticketing.event_categories",
		"ticketing.categories",
		"ticketing.events",
		"crm.customers",
		"auth.users",
		"auth.sessions",
	}

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			t.Logf("Warning: could not truncate %s: %v", table, err)
		}
	}
}
