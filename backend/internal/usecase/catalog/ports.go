package catalog

import (
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Catalog         domain.CatalogRepository
	CategoriesCache domain.Cache[domain.CachedCategories]
	ServicesCache   domain.Cache[domain.CachedDefaultServices]
	CacheTTL        time.Duration
}
