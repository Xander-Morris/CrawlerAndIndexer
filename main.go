package main

import (
	"main/scraper"
)

func main() {
	botAgent := "MyCustomScraperBot/1.0"
	seedUrls := []string{"https://www.gutenberg.org/", "https://www.wikipedia.org/"}
	config := &scraper.ScrapeConfig{MaxLevel: 3, MaxWorkers: 5}
	scraper.StartRecursiveScrape(botAgent, seedUrls, config)
}