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

	// Configurar health checks
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
	// Primero intenta con DSN completo
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	// Si no, construye desde variables individuales
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

// CreateTables crea las tablas necesarias
func CreateTables() error {
	queries := []string{
		// Extensión para UUIDs
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		// Tabla de customers (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS customers (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(50),
			date_of_birth DATE,
			address JSONB,
			preferences JSONB,
			loyalty_points INTEGER DEFAULT 0,
			is_verified BOOLEAN NOT NULL DEFAULT false,
			verification_token VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
		`,

		// Tabla de eventos (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			short_description VARCHAR(500),
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			location VARCHAR(255) NOT NULL,
			venue_details TEXT,
			category VARCHAR(100),
			tags VARCHAR(100)[],
			is_active BOOLEAN NOT NULL DEFAULT true,
			is_published BOOLEAN NOT NULL DEFAULT false,
			image_url VARCHAR(512),
			banner_url VARCHAR(512),
			max_attendees INTEGER CHECK (max_attendees > 0),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
		`,

		// Tabla de categorías (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			price DECIMAL(12,2) NOT NULL CHECK (price >= 0),
			quantity_available INTEGER NOT NULL CHECK (quantity_available >= 0),
			quantity_sold INTEGER NOT NULL DEFAULT 0 CHECK (quantity_sold >= 0),
			max_tickets_per_order INTEGER NOT NULL DEFAULT 10 CHECK (max_tickets_per_order > 0),
			sales_start TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			sales_end TIMESTAMP,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			CHECK (quantity_sold <= quantity_available),
			CHECK (sales_end IS NULL OR sales_end > sales_start)
		)
		`,

		// Tabla de tickets (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS tickets (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			code VARCHAR(50) UNIQUE NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'reserved', 'sold', 'used', 'cancelled', 'transferred')),
			seat_number VARCHAR(20),
			qr_code_url VARCHAR(512),
			price DECIMAL(12,2) NOT NULL CHECK (price >= 0),
			used_at TIMESTAMP,
			transferred_from_ticket_id INTEGER REFERENCES tickets(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
		`,

		// Tabla de transacciones (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
			promotion_id INTEGER REFERENCES promotions(id) ON DELETE SET NULL,
			amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
			discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
			final_amount DECIMAL(12,2) NOT NULL CHECK (final_amount >= 0),
			currency VARCHAR(3) NOT NULL DEFAULT 'USD',
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded', 'cancelled')),
			stripe_session_id VARCHAR(255),
			stripe_payment_intent_id VARCHAR(255),
			payment_method VARCHAR(50),
			customer_email VARCHAR(255) NOT NULL,
			customer_name VARCHAR(255) NOT NULL,
			receipt_url VARCHAR(512),
			metadata JSONB,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			CHECK (final_amount = amount - discount_amount)
		)
		`,

		// Tabla de promociones (con IDENTITY)
		`
		CREATE TABLE IF NOT EXISTS promotions (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
			code VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
			discount_value DECIMAL(10,2) NOT NULL CHECK (discount_value >= 0),
			event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
			min_order_amount DECIMAL(10,2) CHECK (min_order_amount >= 0),
			max_discount_amount DECIMAL(10,2) CHECK (max_discount_amount >= 0),
			valid_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			valid_to TIMESTAMP NOT NULL,
			usage_limit INTEGER CHECK (usage_limit >= 0),
			usage_count INTEGER NOT NULL DEFAULT 0 CHECK (usage_count >= 0),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			CHECK (valid_to > valid_from),
			CHECK (usage_count <= usage_limit OR usage_limit IS NULL)
		)
		`,
	}

	for _, query := range queries {
		log.Printf("🔄 Executing query: %s", getQuerySummary(query))
		_, err := Pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("error creating table: %v\nQuery: %s", err, query)
		}
	}

	log.Println("✅ Main database tables created/verified successfully")

	// Crear tablas de auditoría por separado
	if err := CreateAuditTables(); err != nil {
		return fmt.Errorf("error creating audit tables: %w", err)
	}

	// Crear índices
	if err := CreateIndexes(); err != nil {
		return fmt.Errorf("error creating indexes: %w", err)
	}

	// Crear triggers
	if err := CreateTriggers(); err != nil {
		return fmt.Errorf("error creating triggers: %w", err)
	}

	return nil
}

// CreateAuditTables crea las tablas de auditoría separadamente
func CreateAuditTables() error {
	queries := []string{
		// Auditoría de clientes
		`
		CREATE TABLE IF NOT EXISTS customer_audit (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			customer_id INTEGER NOT NULL,
			public_id UUID NOT NULL,
			operation VARCHAR(10) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
			changed_by VARCHAR(100) NOT NULL DEFAULT current_user,
			changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			old_data JSONB,
			new_data JSONB,
			ip_address INET,
			user_agent TEXT
		)
		`,

		// Auditoría de transacciones
		`
		CREATE TABLE IF NOT EXISTS transaction_audit (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			transaction_id INTEGER NOT NULL,
			public_id UUID NOT NULL,
			operation VARCHAR(10) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
			changed_by VARCHAR(100) NOT NULL DEFAULT current_user,
			changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			old_data JSONB,
			new_data JSONB,
			ip_address INET
		)
		`,

		// Auditoría de tickets
		`
		CREATE TABLE IF NOT EXISTS ticket_audit (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			ticket_id INTEGER NOT NULL,
			public_id UUID NOT NULL,
			operation VARCHAR(10) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
			changed_by VARCHAR(100) NOT NULL DEFAULT current_user,
			changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			old_data JSONB,
			new_data JSONB
		)
		`,
	}

	for _, query := range queries {
		log.Printf("🔄 Creating audit table: %s", getQuerySummary(query))
		_, err := Pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("error creating audit table: %v", err)
		}
	}

	log.Println("✅ Audit tables created/verified successfully")
	return nil
}

// CreateIndexes crea los índices para mejor rendimiento
func CreateIndexes() error {
	indexes := []string{
		// Índices para customers
		`CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_public_id ON customers(public_id)`,

		// Índices para tickets
		`CREATE INDEX IF NOT EXISTS idx_tickets_code ON tickets(code)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_event_id ON tickets(event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_public_id ON tickets(public_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_category_id ON tickets(category_id)`,

		// Índices para events
		`CREATE INDEX IF NOT EXISTS idx_events_start_date ON events(start_date)`,
		`CREATE INDEX IF NOT EXISTS idx_events_is_active ON events(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_events_public_id ON events(public_id)`,
		`CREATE INDEX IF NOT EXISTS idx_events_category ON events(category)`,

		// Índices para categories
		`CREATE INDEX IF NOT EXISTS idx_categories_event_id ON categories(event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active)`,

		// Índices para transactions
		`CREATE INDEX IF NOT EXISTS idx_transactions_customer_id ON transactions(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at)`,

		// Índices para promotions
		`CREATE INDEX IF NOT EXISTS idx_promotions_code ON promotions(code)`,
		`CREATE INDEX IF NOT EXISTS idx_promotions_valid_to ON promotions(valid_to)`,

		// Índices para auditoría
		`CREATE INDEX IF NOT EXISTS idx_customer_audit_customer_id ON customer_audit(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_customer_audit_changed_at ON customer_audit(changed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_audit_transaction_id ON transaction_audit(transaction_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ticket_audit_ticket_id ON ticket_audit(ticket_id)`,

		// Índices para búsqueda de texto
		`CREATE INDEX IF NOT EXISTS idx_events_name_trgm ON events USING gin (name gin_trgm_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_customers_name_trgm ON customers USING gin (name gin_trgm_ops)`,
	}

	for _, index := range indexes {
		log.Printf("🔄 Creating index: %s", getQuerySummary(index))
		_, err := Pool.Exec(context.Background(), index)
		if err != nil {
			// Solo log el error pero no detengas la ejecución (los índices pueden existir)
			log.Printf("⚠️ Warning creating index (might already exist): %v", err)
		}
	}

	log.Println("✅ Database indexes created/verified successfully")
	return nil
}

// CreateTriggers crea los triggers necesarios
func CreateTriggers() error {
	// Primero crear las funciones de trigger
	triggerFunctions := []string{
		// Función para actualizar automáticamente updated_at
		`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
		    NEW.updated_at = CURRENT_TIMESTAMP;
		    RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		`,

		// Función para auditoría de customers
		`
		CREATE OR REPLACE FUNCTION audit_customer_changes()
		RETURNS TRIGGER AS $$
		BEGIN
		  IF TG_OP = 'INSERT' THEN
		    INSERT INTO customer_audit (customer_id, public_id, operation, new_data)
		    VALUES (NEW.id, NEW.public_id, TG_OP, to_jsonb(NEW));
		    RETURN NEW;
		  ELSIF TG_OP = 'UPDATE' THEN
		    INSERT INTO customer_audit (customer_id, public_id, operation, old_data, new_data)
		    VALUES (NEW.id, NEW.public_id, TG_OP, to_jsonb(OLD), to_jsonb(NEW));
		    RETURN NEW;
		  ELSIF TG_OP = 'DELETE' THEN
		    INSERT INTO customer_audit (customer_id, public_id, operation, old_data)
		    VALUES (OLD.id, OLD.public_id, TG_OP, to_jsonb(OLD));
		    RETURN OLD;
		  END IF;
		  RETURN NULL;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;
		`,

		// Función para auditoría de tickets
		`
		CREATE OR REPLACE FUNCTION audit_ticket_changes()
		RETURNS TRIGGER AS $$
		BEGIN
		  IF TG_OP = 'INSERT' THEN
		    INSERT INTO ticket_audit (ticket_id, public_id, operation, new_data)
		    VALUES (NEW.id, NEW.public_id, TG_OP, to_jsonb(NEW));
		    RETURN NEW;
		  ELSIF TG_OP = 'UPDATE' THEN
		    INSERT INTO ticket_audit (ticket_id, public_id, operation, old_data, new_data)
		    VALUES (NEW.id, NEW.public_id, TG_OP, to_jsonb(OLD), to_jsonb(NEW));
		    RETURN NEW;
		  ELSIF TG_OP = 'DELETE' THEN
		    INSERT INTO ticket_audit (ticket_id, public_id, operation, old_data)
		    VALUES (OLD.id, OLD.public_id, TG_OP, to_jsonb(OLD));
		    RETURN OLD;
		  END IF;
		  RETURN NULL;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;
		`,
	}

	for _, function := range triggerFunctions {
		log.Printf("🔄 Creating trigger function: %s", getQuerySummary(function))
		_, err := Pool.Exec(context.Background(), function)
		if err != nil {
			return fmt.Errorf("error creating trigger function: %v", err)
		}
	}

	// Luego crear los triggers
	triggers := []string{
		// Triggers para updated_at
		`CREATE OR REPLACE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE OR REPLACE TRIGGER update_events_updated_at BEFORE UPDATE ON events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE OR REPLACE TRIGGER update_tickets_updated_at BEFORE UPDATE ON tickets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE OR REPLACE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,

		// Triggers para auditoría
		`CREATE OR REPLACE TRIGGER trigger_audit_customer AFTER INSERT OR UPDATE OR DELETE ON customers FOR EACH ROW EXECUTE FUNCTION audit_customer_changes()`,
		`CREATE OR REPLACE TRIGGER trigger_audit_ticket AFTER INSERT OR UPDATE OR DELETE ON tickets FOR EACH ROW EXECUTE FUNCTION audit_ticket_changes()`,
	}

	for _, trigger := range triggers {
		log.Printf("🔄 Creating trigger: %s", getQuerySummary(trigger))
		_, err := Pool.Exec(context.Background(), trigger)
		if err != nil {
			// Los triggers pueden fallar si ya existen, solo log el error
			log.Printf("⚠️ Warning creating trigger (might already exist): %v", err)
		}
	}

	log.Println("✅ Database triggers created/verified successfully")
	return nil
}

// getQuerySummary devuelve un resumen de la query para logging
func getQuerySummary(query string) string {
	if len(query) > 100 {
		return query[:100] + "..."
	}
	return query
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
