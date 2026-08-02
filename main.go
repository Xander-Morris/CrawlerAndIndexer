package main

import (
	"main/scraper"
)

func main() {
	botAgent := "MyCustomScraperBot/1.0"
	seedUrls := []string{"https://www.gutenberg.org/", "https://www.wikipedia.org/"}
	config := &scraper.ScrapeConfig{MaxLevel: 2, MaxWorkers: 100, MaxInMemoryEntries: 1000}
	scraper.ScrapeUrls(botAgent, seedUrls, config)
}
