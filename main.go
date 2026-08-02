package main

import (
	"main/database"
	"main/scraper"
)

func main() {
	botAgent := "MyCustomScraperBot/1.0"
	seedUrls := []string{"https://www.gutenberg.org/", "https://www.wikipedia.org/"}
	config := &scraper.ScrapeConfig{MaxLevel: 2, MaxWorkers: 100}
	wordsToUrls := scraper.ScrapeUrls(botAgent, seedUrls, config)
	database.WriteToDatabase(wordsToUrls)
}