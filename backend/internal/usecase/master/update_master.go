package master

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type UpdateMasterInput struct {
	Bio               *string
	Specialization    *string
	YearsOfExperience *int32
	Languages         []domain.Language
	Certificates      []domain.Certificate
	HasLanguages      bool
	HasCertificates   bool
}

type UpdateMasterOutput struct {
	Profile *domain.MasterProfile
}

type UpdateMasterUseCase struct {
	deps Deps
}

func NewUpdateMasterUseCase(deps Deps) *UpdateMasterUseCase {
	return &UpdateMasterUseCase{deps: deps}
}

func (uc *UpdateMasterUseCase) Execute(ctx context.Context, in UpdateMasterInput) (*UpdateMasterOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	profile, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Bio != nil {
		profile.Bio = *in.Bio
	}
	if in.Specialization != nil {
		profile.Specialization = *in.Specialization
	}
	if in.YearsOfExperience != nil {
		if *in.YearsOfExperience < 0 {
			return nil, domain.ErrInvalidInput
		}
		profile.YearsOfExperience = *in.YearsOfExperience
	}

	updated, err := uc.deps.Masters.Update(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("update master: %w", err)
	}

	if in.HasLanguages {
		if err := uc.deps.Masters.ReplaceLanguages(ctx, updated.ID, in.Languages); err != nil {
			return nil, fmt.Errorf("replace languages: %w", err)
		}
		updated.Languages = in.Languages
	}

	if in.HasCertificates {
		if err := uc.deps.Masters.ReplaceCertificates(ctx, updated.ID, in.Certificates); err != nil {
			return nil, fmt.Errorf("replace certificates: %w", err)
		}
		updated.Certificates = in.Certificates
	}

	invalidateMasterCache(ctx, uc.deps.Cache, updated.Slug, userID)

	return &UpdateMasterOutput{Profile: updated}, nil
}