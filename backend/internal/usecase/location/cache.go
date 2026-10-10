package location

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

func invalidate(ctx context.Context, cache domain.Cache[domain.CachedLocations], masterID int32) {
	_ = cache.Delete(ctx, fmt.Sprintf("locations:master:%d:active", masterID))
}