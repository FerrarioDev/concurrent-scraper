package tests

import (
	"reflect"
	"testing"

	"github.com/FerrarioDev/concurrent-scraper/internal/crawler"
)

func TestScrapeSite(t *testing.T) {
	got, err := crawler.ScrapeSite("https://books.toscrape.com/")
	if err != nil {
		t.Error(err)
	}
	want := crawler.Site{
		Link:  "https://books.toscrape.com/",
		Title: "All products | Books to Scrape - Sandbox",
		Links: []string{},
	}

	if !reflect.DeepEqual(got, &want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
