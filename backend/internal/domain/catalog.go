package domain

import (
	"context"
	"time"
)

type Category struct {
	ID        int32
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DefaultService struct {
	ID           int32
	CategoryID   int32
	Slug         string
	Name         string
	Description  string
	Price        int32
	Currency     string
	DurationMins int32
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CachedCategories struct {
	Categories []*Category `json:"categories"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type CachedDefaultServices struct {
	CategoryID *int32            `json:"category_id,omitempty"`
	Services   []*DefaultService `json:"services"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type CatalogRepository interface {
	ListCategories(ctx context.Context) ([]*Category, error)
	GetCategory(ctx context.Context, id int32) (*Category, error)
	CreateCategory(ctx context.Context, c *Category) (*Category, error)
	UpdateCategory(ctx context.Context, c *Category) (*Category, error)
	DeleteCategory(ctx context.Context, id int32) error
	CategoryHasServices(ctx context.Context, id int32) (bool, error)

	ListDefaultServices(ctx context.Context, categoryID *int32) ([]*DefaultService, error)
	GetDefaultService(ctx context.Context, id int32) (*DefaultService, error)
	CreateDefaultService(ctx context.Context, s *DefaultService) (*DefaultService, error)
	UpdateDefaultService(ctx context.Context, s *DefaultService) (*DefaultService, error)
	DeleteDefaultService(ctx context.Context, id int32) error
	DefaultServiceSlugExists(ctx context.Context, slug string, excludeID int32) (bool, error)
}
