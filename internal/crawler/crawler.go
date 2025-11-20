// Crawler retrieves links - title - size of several websites
package crawler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Crawler interface {
	ScrapeSite(url string) Site
}

type Site struct {
	Link  string
	Title string
	Links []string
}

func ScrapeSite(urlStr string) (*Site, error) {
	// var site Site

	res, err := http.Get(urlStr)
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

	return &Site{
		Link:  urlStr,
		Title: title,
		Links: links,
	}, nil
}

func Crawl(seedURL string, maxPages int, workers int) []Site {
	urlQueue := make(chan string, 100)
	results := make(chan Site, 100)

	visited := make(map[string]bool)

	var mu sync.Mutex

	var wg sync.WaitGroup

	for i := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for url := range urlQueue {
				fmt.Printf("Worker %d scraping: %s\n", id, url)
				site, err := ScrapeSite(url)
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

	urlQueue <- seedURL
	visited[seedURL] = true
	pagesScraped := 0

	var sites []Site
	for site := range results {
		sites = append(sites, site)
		pagesScraped++

		fmt.Printf("Scraped: %s (found %d links)\n", site.Title, len(site.Links))

		if pagesScraped >= maxPages {
			close(urlQueue)
			continue
		}

		for _, link := range site.Links {
			mu.Lock()
			if !visited[link] && shouldCrawl(link, seedURL) {
				visited[link] = true
				select {
				case urlQueue <- link:
				default:
				}
			}
			mu.Unlock()
		}

		mu.Lock()
		if len(urlQueue) == 0 && pagesScraped >= maxPages {
			close(urlQueue)
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
