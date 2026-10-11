package catalog

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

func invalidateCategories(ctx context.Context, cache domain.Cache[domain.CachedCategories]) {
	_ = cache.Delete(ctx, "catalog:categories")
}

func invalidateServices(ctx context.Context, cache domain.Cache[domain.CachedDefaultServices], categoryID *int32) {
	if categoryID == nil {
		_, _ = cache.DeleteByPattern(ctx, "catalog:default_services:*")
		return
	}
	_ = cache.Delete(ctx, fmt.Sprintf("catalog:default_services:%d", *categoryID))
	_ = cache.Delete(ctx, "catalog:default_services:all")
}
