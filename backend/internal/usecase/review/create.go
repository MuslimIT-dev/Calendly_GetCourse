package review

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CreateInput struct {
	Target      domain.ReviewTarget
	TargetID    int32
	ReferenceID string
	Rating      int32
	Comment     string
}

type CreateOutput struct {
	Review *domain.ReviewCard
}

type CreateUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewCreateUseCase(deps Deps, events EventPublisher) *CreateUseCase {
	return &CreateUseCase{deps: deps, events: events}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*CreateOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if in.TargetID <= 0 || in.ReferenceID == "" || in.Rating < 1 || in.Rating > 5 {
		return nil, domain.ErrInvalidInput
	}

	switch in.Target {
	case domain.ReviewTargetMaster:
		appointmentID, err := strconv.ParseInt(in.ReferenceID, 10, 64)
		if err != nil || appointmentID <= 0 {
			return nil, domain.ErrInvalidInput
		}
		card, err := uc.deps.Bookings.GetByID(ctx, appointmentID)
		if err != nil {
			if errors.Is(err, domain.ErrAppointmentNotFound) {
				return nil, domain.ErrReviewNotAllowed
			}
			return nil, err
		}
		if card == nil || card.Appointment == nil {
			return nil, domain.ErrReviewNotAllowed
		}
		appointment := card.Appointment
		if appointment.ClientID != userID || appointment.MasterID != in.TargetID {
			return nil, domain.ErrForbidden
		}
		if appointment.Status != domain.AppointmentCompleted {
			return nil, domain.ErrReviewNotAllowed
		}
	case domain.ReviewTargetCourse:
		return nil, domain.ErrReviewNotAllowed
	default:
		return nil, domain.ErrInvalidInput
	}

	exists, err := uc.deps.Reviews.ExistsForReference(ctx, userID, in.ReferenceID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrReviewAlreadyExists
	}

	review, err := uc.deps.Reviews.Create(ctx, &domain.Review{
		AuthorID:    userID,
		TargetID:    in.TargetID,
		Target:      in.Target,
		ReferenceID: in.ReferenceID,
		Rating:      in.Rating,
		Comment:     strings.TrimSpace(in.Comment),
		Status:      domain.ReviewPublished,
	})
	if err != nil {
		return nil, err
	}

	invalidateSummary(ctx, uc.deps.Cache, in.Target, in.TargetID)
	card, err := uc.deps.Reviews.GetByID(ctx, review.ID)
	if err != nil {
		return nil, err
	}

	if uc.events != nil {
		if err := publishEvent(ctx, uc.events, ReviewCreatedEvent{
			ReviewID:    review.ID,
			AuthorID:    review.AuthorID,
			TargetID:    review.TargetID,
			TargetType:  int32(review.Target),
			Rating:      review.Rating,
			Comment:     review.Comment,
			ReferenceID: review.ReferenceID,
		}); err != nil {
			log.Printf("publish review-created event: %v", err)
		}
	}

	return &CreateOutput{Review: card}, nil
}
