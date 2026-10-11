package worker

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/notification"
)

type ReviewEvent struct {
	ReviewID    int64  `json:"review_id"`
	AuthorID    int32  `json:"author_id"`
	TargetID    int32  `json:"target_id"`
	TargetType  int32  `json:"target_type"`
	Rating      int32  `json:"rating"`
	Comment     string `json:"comment"`
	ReferenceID string `json:"reference_id"`
	Reply       string `json:"reply"`
}

type ReviewsConsumer struct {
	sendUC  *notificationuc.SendUseCase
	masters masterOwnerLookup
}

func NewReviewsConsumer(sendUC *notificationuc.SendUseCase, masters masterOwnerLookup) *ReviewsConsumer {
	return &ReviewsConsumer{sendUC: sendUC, masters: masters}
}

func (c *ReviewsConsumer) Consume(ctx context.Context, event *ReviewEvent) error {
	if event == nil || event.ReviewID <= 0 {
		return fmt.Errorf("invalid review event")
	}
	link := fmt.Sprintf("/reviews/%d", event.ReviewID)
	userID := event.AuthorID
	notificationType, title, body := domain.NotifReviewReplied, "Ответ на ваш отзыв", event.Reply
	if event.Reply == "" {
		if event.TargetType != int32(domain.ReviewTargetMaster) {
			return nil
		}
		var err error
		userID, err = c.masters.GetUserIDByMasterID(ctx, event.TargetID)
		if err != nil {
			return fmt.Errorf("resolve review target user: %w", err)
		}
		notificationType, title = domain.NotifReviewCreated, "Новый отзыв"
		body = fmt.Sprintf("Вам оставили отзыв: %d из 5", event.Rating)
	}
	if userID <= 0 {
		return fmt.Errorf("invalid review notification recipient")
	}
	_, err := c.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID: userID, Type: notificationType, Channel: domain.ChannelInApp,
		Title: title, Body: body, Link: link,
		Metadata: map[string]any{"review_id": event.ReviewID, "rating": event.Rating},
	})
	return err
}
