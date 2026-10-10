package domain

import (
	"context"
	"time"
)

type Location struct {
	ID         int32
	MasterID   int32
	Name       string
	Address    string
	Timezone   string
	IsOnline   bool
	MeetingURL string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CachedLocations struct {
	MasterID  int32       `json:"master_id"`
	Locations []*Location `json:"locations"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type LocationRepository interface {
	ListByMaster(ctx context.Context, masterID int32, onlyActive bool) ([]*Location, error)
	GetByID(ctx context.Context, id int32) (*Location, error)
	Create(ctx context.Context, l *Location) (*Location, error)
	Update(ctx context.Context, l *Location) (*Location, error)
	Delete(ctx context.Context, id int32) error
	BelongsToMaster(ctx context.Context, id, masterID int32) (bool, error)
	ExistsActiveByMaster(ctx context.Context, masterID int32) (bool, error)
}