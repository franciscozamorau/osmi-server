// internal/domain/repository/category_repository.go
package repository

import (
	"context"
	"errors"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// CategoryFilter encapsula TODOS los criterios de búsqueda para categorías.
// Es el corazón de la flexibilidad.
type CategoryFilter struct {
	// Filtros exactos
	IDs        []int64  `db:"ids"`          // Para GetByIDs
	PublicIDs  []string `db:"public_uuids"` // Filtrar por múltiples UUIDs
	ParentID   *int64   `db:"parent_id"`    // nil = todas, 0 = raíces (parent_id IS NULL)
	IsActive   *bool    `db:"is_active"`
	IsFeatured *bool    `db:"is_featured"`

	// Filtros de texto
	SearchTerm *string // Busca en name, slug, description
	Slug       *string `db:"slug"` // Para GetBySlug, pero en filtro

	// Filtros de rango
	MinLevel *int
	MaxLevel *int

	// Paginación y ordenamiento (Siempre presentes para List)
	Limit     int
	Offset    int
	SortBy    string // "name", "created_at", "total_events", "sort_order", etc.
	SortOrder string // "asc", "desc"
}

// CategoryNode representa una categoría con sus hijos para árboles jerárquicos.
// Sigue siendo útil como estructura de retorno.
type CategoryNode struct {
	*entities.Category
	Children []*CategoryNode `json:"children,omitempty"`
}

// Errores específicos del repositorio (Excelentes, los dejamos)
var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryDuplicateSlug = errors.New("category slug already exists")
	ErrCategoryHasChildren   = errors.New("category has children, cannot delete")
	ErrInvalidParent         = errors.New("invalid parent category")
)

type CategoryRepository interface {
	// --- Operaciones de Escritura ---
	Create(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id int64) error

	// --- Operaciones de Lectura (Flexibles) ---
	// El método principal para obtener múltiples categorías.
	// Acepta un filtro que puede ser tan simple o complejo como se necesite.
	// Devuelve las categorías y el total de registros (ignorando paginación).
	Find(ctx context.Context, filter *CategoryFilter) ([]*entities.Category, int64, error)

	// Atajos para casos de uso muy comunes que no necesitan un filtro complejo.
	// Internamente, estos métodos llaman a Find con un filtro específico.
	GetByID(ctx context.Context, id int64) (*entities.Category, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Category, error)

	// --- Operaciones de Verificación ---
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// --- Operaciones de Jerarquía (Lógica de negocio compleja que merece su propio método) ---
	// GetTree merece un método propio porque la construcción del árbol es compleja
	// y no se puede hacer eficientemente con un solo Find.
	GetTree(ctx context.Context, rootID *int64) ([]*CategoryNode, error)

	// --- Operaciones de Negocio (Actualizaciones específicas) ---
	// Estas son acciones de negocio muy claras, no simples actualizaciones de campo.
	IncrementEventCount(ctx context.Context, categoryID int64) error
	DecrementEventCount(ctx context.Context, categoryID int64) error
	UpdateEventStats(ctx context.Context, categoryID int64, ticketSold int64, revenue float64) error

	// --- Relaciones con Eventos ---
	// Estas son específicas y no se cubren con el filtro de categoría.
	GetEventCategories(ctx context.Context, eventID int64) ([]*entities.Category, error)
	AddEventToCategory(ctx context.Context, eventID, categoryID int64, isPrimary bool) error
	RemoveEventFromCategory(ctx context.Context, eventID, categoryID int64) error
	GetPrimaryCategoryForEvent(ctx context.Context, eventID int64) (*entities.Category, error)
}
