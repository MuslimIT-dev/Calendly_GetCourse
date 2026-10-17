package worker

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
)

type ReviewEvent struct {
	ReviewID    int32  `json:"review_id"`
	AuthorID    int32  `json:"author_id"`
	TargetID    int32  `json:"target_id"`
	TargetType  int32  `json:"target_type"`
	Rating      int32  `json:"rating"`
	Comment     string `json:"comment"`
	ReferenceID string `json:"reference_id"`
	Reply       string `json:"reply"`
}

type ReviewsConsumer struct {
	sendUC *notificationuc.SendUseCase
	masters domain.MasterRepository
}

func NewReviewsConsumer(sendUC *notificationuc.SendUseCase, masters domain.MasterRepository) *ReviewsConsumer {
	return &ReviewsConsumer{sendUC: sendUC, masters: masters}
}

func (c *ReviewsConsumer) Consume(ctx context.Context, e *ReviewEvent) error {
	if domain.ReviewTarget(e.TargetType) == domain.ReviewTargetMaster {
		_ = c.masters.UpdateRating(ctx, e.TargetID)
	}

	link := fmt.Sprintf("/reviews/%d", e.ReviewID)

	if e.Reply == "" {
		_, _ = c.sendUC.Execute(ctx, notificationuc.SendInput{
			UserID:  e.TargetID,
			Type:    domain.NotifReviewCreated,
			Channel: domain.ChannelInApp,
			Title:   "Новый отзыв",
			Body:    fmt.Sprintf("Вам оставили отзыв: %d из 5", e.Rating),
			Link:    link,
			Metadata: map[string]any{
				"review_id": e.ReviewID,
				"rating":    e.Rating,
			},
		})
		return nil
	}

	_, _ = c.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID:  e.AuthorID,
		Type:    domain.NotifReviewReplied,
		Channel: domain.ChannelInApp,
		Title:   "Ответ на ваш отзыв",
		Body:    e.Reply,
		Link:    link,
	})

	return nil
}