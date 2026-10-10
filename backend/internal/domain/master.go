package domain

import (
	"context"
	"time"
)

type ProficiencyLevel int32

const (
	ProficiencyUnspecified  ProficiencyLevel = 0
	ProficiencyBeginner     ProficiencyLevel = 1
	ProficiencyIntermediate ProficiencyLevel = 2
	ProficiencyAdvanced     ProficiencyLevel = 3
)

type Language struct {
	Name        string
	Proficiency ProficiencyLevel
}

type Certificate struct {
	Name         string
	Organization string
	Year         int32
	FileURL      string
}

type MasterProfile struct {
	ID					int32
	UserID              int32
	Slug                string
	Bio                 string
	Specialization      string
	YearsOfExperience   int32
	Languages           []Language
	Certificates        []Certificate
	IsAcceptingBookings bool
	DefaultLocationID   *int32
	AvgRating           float32
	ReviewsCount        int32
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type SortType int32

const (
	SortTypeUnspecified SortType = 0
	SortTypeRating      SortType = 1
	SortTypePrice       SortType = 2
	SortTypeExperience  SortType = 3
)

type SortOrder int32

const (
	SortOrderUnspecified SortOrder = 0
	SortOrderAsc         SortOrder = 1
	SortOrderDesc        SortOrder = 2
)

type MasterFilter struct {
	SortBy         SortType
	SortOrder      SortOrder
	MinPrice       *int32
	MaxPrice       *int32
	MinExperience  *int32
	Specialization string
	ServiceID      *int32
	MasterTimezone string
}

type MasterCardData struct {
	Profile       *MasterProfile
	User          *User
	MinPrice      int32
	TotalServices int32
}

type CachedMaster struct {
	Profile *MasterProfile `json:"profile"`
	User    *CachedUser    `json:"user"`
}

type MasterRepository interface {
	GetBySlug(ctx context.Context, slug string) (*MasterProfile, *User, error)
	GetByUserID(ctx context.Context, userID int32) (*MasterProfile, error)
	Create(ctx context.Context, p *MasterProfile) (*MasterProfile, error)
	Update(ctx context.Context, p *MasterProfile) (*MasterProfile, error)
	UpdateSlug(ctx context.Context, id int32, slug string) error
	SlugExists(ctx context.Context, slug string, excludeID int32) (bool, error)

	ReplaceLanguages(ctx context.Context, masterID int32, languages []Language) error
	ReplaceCertificates(ctx context.Context, masterID int32, certs []Certificate) error

	ListCards(ctx context.Context, f MasterFilter, cursor *ListCursor, limit int32) ([]*MasterCardData, error)
}