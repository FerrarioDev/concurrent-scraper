// Crawler retrieves links - title - size of several websites
package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/FerrarioDev/concurrent-scraper/internal/domain"
	"github.com/FerrarioDev/concurrent-scraper/internal/repository"
	"github.com/PuerkitoBio/goquery"
)

type Crawler interface {
	ScrapeSite(ctx context.Context, timeout time.Duration, urlStr string, fatherID *int) (*domain.Site, error)
	Crawl(ctx context.Context, params domain.Params) []domain.Site
}

type CrawlerService struct {
	repo repository.Repository
}

func NewCrawlerService(repo repository.Repository) Crawler {
	return &CrawlerService{repo}
}

func (s *CrawlerService) ScrapeSite(ctx context.Context, timeout time.Duration, urlStr string, fatherID *int) (*domain.Site, error) {
	// var site Site
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	baseURL, _ := url.Parse(urlStr)
	var links []string

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			linkURL, err := baseURL.Parse(href)
			if err == nil {
				links = append(links, linkURL.String())
			}
		}
	})

	title := strings.TrimSpace(doc.Find("title").Text())

	siteReq := domain.SiteRequest{
		URL:      urlStr,
		Title:    title,
		Links:    len(links),
		FatherID: fatherID,
	}

	siteRes, err := s.repo.Create(ctx, &siteReq)
	if err != nil {
		log.Printf("Error saving to DB: %v", err)
		return nil, err
	}

	return &domain.Site{
		ID:    *siteRes.ID,
		Link:  urlStr,
		Title: title,
		Links: links,
	}, nil
}

func (s *CrawlerService) Crawl(ctx context.Context, params domain.Params) []domain.Site {
	urlQueue := make(chan string, 100)
	results := make(chan domain.Site, 100)

	visited := make(map[string]bool)
	urlToID := make(map[string]*int)

	var mu sync.Mutex
	var wg sync.WaitGroup

	queueClosed := false

	for i := range params.Workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for url := range urlQueue {
				fmt.Printf("Worker %d scraping: %s\n", id, url)

				mu.Lock()
				parentID := urlToID[url]
				mu.Unlock()

				site, err := s.ScrapeSite(ctx, params.Timeout, url, parentID)
				if err != nil {
					log.Printf("Error scraping %s: %v", url, err)
					continue
				}
				results <- *site
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	urlQueue <- params.BaseURL
	visited[params.BaseURL] = true
	urlToID[params.BaseURL] = nil
	pagesScraped := 0
	var sites []domain.Site

	for site := range results {
		sites = append(sites, site)
		pagesScraped++

		fmt.Printf("Scraped: %s (found %d links)\n", site.Title, len(site.Links))

		if pagesScraped >= params.MaxPages {
			mu.Lock()
			if !queueClosed {
				close(urlQueue)
				queueClosed = true
			}
			mu.Unlock()
			continue
		}

		for _, link := range site.Links {
			mu.Lock()
			if !visited[link] && shouldCrawl(link, params.BaseURL) {
				visited[link] = true
				urlToID[link] = &site.ID // set site as parent for the child link
				select {
				case urlQueue <- link:
				default:
				}
			}
			mu.Unlock()
		}

		mu.Lock()
		if len(urlQueue) == 0 && pagesScraped >= params.MaxPages && !queueClosed {
			close(urlQueue)
			queueClosed = true
		}
		mu.Unlock()
	}
	return sites
}

func shouldCrawl(link, seedURL string) bool {
	linkURL, err1 := url.Parse(link)
	seedURLParsed, err2 := url.Parse(seedURL)

	if err1 != nil || err2 != nil {
		return false
	}

	return linkURL.Host == seedURLParsed.Host
}
