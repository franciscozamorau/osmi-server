package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// Init inicializa la conexión a la base de datos usando pgxpool
func Init() error {
	connStr := getConnectionString()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Configurar el pool de conexiones
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 2 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verificar la conexión con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Printf("✅ Database connected successfully (connections: %d)", config.MaxConns)
	return nil
}

// getConnectionString obtiene la cadena de conexión desde variables de entorno
func getConnectionString() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "osmi")
	password := getEnv("DB_PASSWORD", "osmi1405")
	dbname := getEnv("DB_NAME", "osmidb")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Close cierra la conexión a la base de datos
func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("✅ Database connection closed")
	}
}

// HealthCheck verifica el estado de la base de datos
func HealthCheck() error {
	if Pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return Pool.Ping(ctx)
}

// GetStats devuelve estadísticas del pool de conexiones
func GetStats() *pgxpool.Stat {
	if Pool == nil {
		return nil
	}
	return Pool.Stat()
}
