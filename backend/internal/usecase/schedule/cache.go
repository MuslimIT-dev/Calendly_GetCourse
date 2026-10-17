package schedule

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

func invalidateRules(ctx context.Context, cache domain.Cache[domain.CachedScheduleRules], locationID int32) {
	_ = cache.Delete(ctx, fmt.Sprintf("schedule:rules:%d", locationID))
}