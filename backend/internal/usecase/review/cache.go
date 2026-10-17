package review

import (
	"context"
	"fmt"
	"log"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

func summaryKey(target domain.ReviewTarget, targetID int32) string {
	return fmt.Sprintf("review:summary:%d:%d", target, targetID)
}

func invalidateSummary(ctx context.Context, cache domain.Cache[domain.CachedReviewSummary], target domain.ReviewTarget, targetID int32) {
	if err := cache.Delete(ctx, summaryKey(target, targetID)); err != nil {
		log.Printf("invalidate review summary cache: %v", err)
	}
}
