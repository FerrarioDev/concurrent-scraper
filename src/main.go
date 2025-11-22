package main

import (
	"context"
	"flag"
	"time"

	"github.com/FerrarioDev/concurrent-scraper/internal/crawler"
	"github.com/FerrarioDev/concurrent-scraper/internal/domain"
	"github.com/FerrarioDev/concurrent-scraper/internal/repository"
	"github.com/FerrarioDev/concurrent-scraper/internal/utils"
)

func main() {
	baseURL := flag.String("url", "required", "starting URL (required)")
	workers := flag.Int("concurrency", 20, "max concurrent workers (default 20)")
	maxPages := flag.Int("max-pages", 10, "maximum number of pages to crawl (default unlimited)")
	// rateLimit := flag.Int("rate-limit", 10, "global requests per second (default 10)")
	timeout := flag.Int("timeout", 15, "per-request timeout (default 15s)")
	// depth := flag.Int("depth", 0, "maximum crawl depth (default infinite)")
	// sameDomain := flag.Bool("same-domain", true, "only follow links within the starting domain (default true)")

	flag.Parse()

	params := domain.Params{
		BaseURL:  *baseURL,
		Workers:  *workers,
		MaxPages: *maxPages,
		Timeout:  time.Duration(*timeout),
	}

	db := utils.InitDB()
	sqlite := repository.NewSqliteRepository(db)

	crawlerService := crawler.NewCrawlerService(sqlite)

	crawlerService.Crawl(context.Background(), params)
}
