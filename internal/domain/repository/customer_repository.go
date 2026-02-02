package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// CustomerRepository define las operaciones para clientes
type CustomerRepository interface {
	// CRUD básico
	Create(ctx context.Context, customer *entities.Customer) error
	FindByID(ctx context.Context, id int64) (*entities.Customer, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Customer, error)
	FindByEmail(ctx context.Context, email string) (*entities.Customer, error)
	FindByUserID(ctx context.Context, userID int64) (*entities.Customer, error)
	Update(ctx context.Context, customer *entities.Customer) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter dto.CustomerFilter, pagination dto.Pagination) ([]*entities.Customer, int64, error)
	FindByName(ctx context.Context, name string, limit int) ([]*entities.Customer, error)
	FindByType(ctx context.Context, customerType string, pagination dto.Pagination) ([]*entities.Customer, int64, error)
	FindByCountry(ctx context.Context, country string, pagination dto.Pagination) ([]*entities.Customer, int64, error)
	Search(ctx context.Context, term string, limit int) ([]*entities.Customer, error)

	// Operaciones específicas
	UpdateStats(ctx context.Context, customerID int64, amount float64) error
	UpdateLoyaltyPoints(ctx context.Context, customerID int64, points int32) error
	UpdateVerification(ctx context.Context, customerID int64, verified bool) error
	UpdatePreferences(ctx context.Context, customerID int64, preferences map[string]interface{}) error
	SetVIP(ctx context.Context, customerID int64, isVIP bool) error
	UpdateInvoiceSettings(ctx context.Context, customerID int64, requiresInvoice bool, taxID, taxName string) error

	// Estadísticas
	GetStats(ctx context.Context) (*dto.CustomerStatsResponse, error)
	GetVIPCustomers(ctx context.Context) ([]*entities.Customer, error)
	CountByType(ctx context.Context, customerType string) (int64, error)
	GetTotalSpent(ctx context.Context, customerID int64) (float64, error)
	GetPurchaseHistory(ctx context.Context, customerID int64, limit int) ([]*dto.PurchaseRecord, error)
}
