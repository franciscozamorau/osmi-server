package repository

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// CustomerFilter encapsula TODOS los criterios de búsqueda para clientes
type CustomerFilter struct {
	// Filtros por ID
	IDs       []int64
	PublicIDs []string
	UserID    *int64
	Email     *string

	// Filtros de texto
	SearchTerm  *string // Busca en full_name, email, company_name, tax_id
	FullName    *string
	CompanyName *string
	Country     *string
	City        *string

	// Filtros booleanos
	IsActive        *bool
	IsVIP           *bool
	RequiresInvoice *bool

	// Filtros de segmento
	CustomerSegment *string

	// Filtros de rango de fechas
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	LastPurchaseFrom *time.Time
	LastPurchaseTo   *time.Time

	// Filtros de estadísticas
	MinTotalSpent  *float64
	MaxTotalSpent  *float64
	MinTotalOrders *int32
	MaxTotalOrders *int32

	// Paginación y ordenamiento
	Limit     int
	Offset    int
	SortBy    string // "created_at", "total_spent", "total_orders", "last_purchase_at"
	SortOrder string // "asc", "desc"
}

// Errores específicos del repositorio
var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrCustomerEmailExists   = errors.New("customer email already exists")
	ErrCustomerAlreadyLinked = errors.New("customer already linked to a user")
)

type CustomerRepository interface {
	// --- Operaciones de Escritura ---
	Create(ctx context.Context, customer *entities.Customer) error
	Update(ctx context.Context, customer *entities.Customer) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// --- Operaciones de Lectura (Flexibles) ---
	Find(ctx context.Context, filter *CustomerFilter) ([]*entities.Customer, int64, error)

	// Atajos
	GetByID(ctx context.Context, id int64) (*entities.Customer, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entities.Customer, error)
	GetByUserID(ctx context.Context, userID int64) (*entities.Customer, error)

	// --- Operaciones de Verificación ---
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// --- Operaciones de Estadísticas ---
	UpdateStats(ctx context.Context, customerID int64, amount float64) error
	UpdateLoyaltyPoints(ctx context.Context, customerID int64, points int32) error
	SetVIP(ctx context.Context, customerID int64, isVIP bool) error

	// --- Operaciones de Preferencias ---
	UpdatePreferences(ctx context.Context, customerID int64, preferences map[string]interface{}) error
	UpdateInvoiceSettings(ctx context.Context, customerID int64, requiresInvoice bool, taxID, taxName string) error

	// --- Estadísticas Agregadas ---
	GetStats(ctx context.Context) (*CustomerStats, error)
	GetVIPCustomers(ctx context.Context) ([]*entities.Customer, error)
}

// CustomerStats representa estadísticas agregadas de clientes
// Los tags `db:` son NECESARIOS para que sqlx pueda mapear los resultados de la query
type CustomerStats struct {
	TotalCustomers         int64         `db:"total_customers" json:"total_customers"`
	ActiveCustomers        int64         `db:"active_customers" json:"active_customers"`
	VIPCustomers           int64         `db:"vip_customers" json:"vip_customers"`
	NewCustomersLast30Days int64         `db:"new_customers_last_30_days" json:"new_customers_last_30_days"`
	TotalRevenue           float64       `db:"total_revenue" json:"total_revenue"`
	AvgLifetimeValue       float64       `db:"avg_lifetime_value" json:"avg_lifetime_value"`
	TopCountries           []CountryStat `json:"top_countries,omitempty"`
}

type CountryStat struct {
	Country string  `db:"country" json:"country"` // ← Añadido tag db: para consistencia
	Count   int64   `db:"count" json:"count"`     // ← Añadido tag db: para consistencia
	Revenue float64 `db:"revenue" json:"revenue"` // ← Añadido tag db: para consistencia
}
