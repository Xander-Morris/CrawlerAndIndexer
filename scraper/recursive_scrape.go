package scraper

import (
	"fmt"
	"sync"
	"time"
)

type ScrapeConfig struct {
	MaxLevel   int
	MaxWorkers int
}

func recursiveScrape(botAgent string, guard *RobotsGuard, visited *sync.Map, wordsToUrls *WordsToUrls, seedUrls []string, config *ScrapeConfig, level int) {
	fmt.Println("About to start a recursive scrape with seed urls!")
	results := make(chan ScrapeResults, len(seedUrls))

	for i := 0; i < len(seedUrls); {
		var wg sync.WaitGroup

		for j := 0; j < config.MaxWorkers; {
			if i >= len(seedUrls) {
				break
			}

			pageURL := seedUrls[i]
			i++
			_, exists := visited.Load(pageURL)

			if exists {
				continue
			}

			fmt.Printf("Fetching for %s\n", pageURL)
			allowed, err := guard.IsAllowed(pageURL)

			if !allowed || err != nil {
				continue
			}

			j++
			visited.Store(pageURL, "")
			wg.Go(func() { extractFromPage(botAgent, pageURL, wordsToUrls, results) })
		}

		wg.Wait()
		fmt.Println("Done with worker batch!")
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	if level >= config.MaxLevel {
		return
	}

	var combinedNextUrls []string

	for i := 1; i <= len(seedUrls); i++ {
		nextUrls := <-results
		combinedNextUrls = append(combinedNextUrls, nextUrls.nextURLs...)
	}

	recursiveScrape(botAgent, guard, visited, wordsToUrls, combinedNextUrls, config, level+1)
}

func ScrapeUrls(botAgent string, seedUrls []string, config *ScrapeConfig) *WordsToUrls {
	guard := &RobotsGuard{userAgent: botAgent}
	var visited sync.Map
	var wordsToUrls = NewWordsToUrls()

	recursiveScrape(botAgent, guard, &visited, wordsToUrls, seedUrls, config, 1)

	return wordsToUrls
}