package repository

import (
	"context"
	"log"
	"time"

	"github.com/FerrarioDev/concurrent-scraper/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return &SqliteRepository{db}
}

func (r *SqliteRepository) FetchAll(ctx context.Context) ([]*domain.Site, error) {
	return nil, nil
}

func (r *SqliteRepository) FetchByID(ctx context.Context, id int) (*domain.Site, error) {
	return nil, nil
}

func (r *SqliteRepository) Create(ctx context.Context, site *domain.SiteRequest) (*domain.SiteRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO sites (url, title, links, father_id) VALUES (:url, :title, :links, :father_id)`

	result, err := r.db.NamedExecContext(ctx, query, site)
	if err != nil {
		log.Printf("failed to insert site: %v", err)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get site id %v", err)
	}

	intID := int(id)
	site.ID = &intID

	return site, nil
}
