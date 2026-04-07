package repository

import (
	"context"
	"errors"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

type CategoryFilter struct {
	IDs        []int64
	PublicIDs  []string
	EventID    *string // UUID del evento
	ParentID   *int64
	IsActive   *bool
	IsFeatured *bool
	SearchTerm *string
	Slug       *string
	MinLevel   *int
	MaxLevel   *int
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}

type CategoryNode struct {
	*entities.Category
	Children []*CategoryNode `json:"children,omitempty"`
}

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryDuplicateSlug = errors.New("category slug already exists")
	ErrCategoryDuplicateName = errors.New("category name already exists for this event")
	ErrCategoryHasChildren   = errors.New("category has children, cannot delete")
	ErrInvalidParent         = errors.New("invalid parent category")
)

type CategoryRepository interface {
	Create(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id int64) error

	Find(ctx context.Context, filter *CategoryFilter) ([]*entities.Category, int64, error)
	GetByID(ctx context.Context, id int64) (*entities.Category, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Category, error)
	GetByEventID(ctx context.Context, eventID string, isActive *bool) ([]*entities.Category, error)

	Exists(ctx context.Context, id int64) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	GetTree(ctx context.Context, rootID *int64) ([]*CategoryNode, error)

	IncrementEventCount(ctx context.Context, categoryID int64) error
	DecrementEventCount(ctx context.Context, categoryID int64) error
	UpdateEventStats(ctx context.Context, categoryID int64, ticketSold int64, revenue float64) error
}
