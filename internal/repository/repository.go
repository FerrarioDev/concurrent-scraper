// Repository is created to handle database and file operations
package repository

import (
	"context"

	"github.com/FerrarioDev/concurrent-scraper/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, site *domain.SiteRequest) (*domain.SiteRequest, error)
	FetchByID(ctx context.Context, id int) (*domain.Site, error)
	FetchAll(ctx context.Context) ([]*domain.Site, error)
}
