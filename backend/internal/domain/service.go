package domain

import (
	"context"
	"time"
)

type ServiceLocation struct {
	LocationID   int32
	DurationMins int32
	Price        int32
	IsActive     bool
}

type Service struct {
	ID               int32
	MasterID         int32
	DefaultServiceID *int32
	Slug             string
	Name             string
	Description      string
	Price            int32
	Currency         string
	DurationMins     int32
	IsActive         bool
	Locations        []ServiceLocation
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CachedServices struct {
	MasterID  int32      `json:"master_id"`
	Services  []*Service `json:"services"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type DefaultServiceTemplate struct {
	ID           int32
	Slug         string
	Name         string
	Description  string
	Price        int32
	Currency     string
	DurationMins int32
}

type ServiceRepository interface {
	GetByID(ctx context.Context, id int32) (*Service, error)
	ListByMaster(ctx context.Context, masterID int32) ([]*Service, error)
	Create(ctx context.Context, s *Service) (*Service, error)
	Update(ctx context.Context, s *Service) (*Service, error)
	Delete(ctx context.Context, id, masterID int32) error
	ExistsByDefaultID(ctx context.Context, masterID, defaultServiceID int32) (bool, error)
	GetDefaultTemplate(ctx context.Context, id int32) (*DefaultServiceTemplate, error)
	ReplaceLocations(ctx context.Context, serviceID int32, locs []ServiceLocation) error
	UpdateMasterMinPrice(ctx context.Context, masterID int32) error
}
