package scraper

import (
	"fmt"
	"sync"
	"time"
)

func recursiveScrape(botAgent string, guard *RobotsGuard, visited *sync.Map, wordToUrls *sync.Map, seedUrls []string, config *ScrapeConfig, level int) {
	if level >= config.MaxLevel {
		fmt.Println("Max level!")
		return
	}

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
			wg.Go(func() { extractFromPage(botAgent, pageURL, wordToUrls, results) })
		}

		wg.Wait()
		fmt.Println("Done with worker batch!")
		time.Sleep(0.5 * 100000000)
	}

	time.Sleep(1 * 100000000)

	fmt.Println("Starting to get those combined next urls!")

	var combinedNextUrls []string

	for i := 1; i <= len(seedUrls); i++ {
		nextUrls := <-results
		combinedNextUrls = append(combinedNextUrls, nextUrls.nextURLs...)
	}

	fmt.Println(len(combinedNextUrls))

	close(results)
	recursiveScrape(botAgent, guard, visited, wordToUrls, combinedNextUrls, config, level + 1)
}

func StartRecursiveScrape(botAgent string, seedUrls []string, config *ScrapeConfig) {
	guard := &RobotsGuard{userAgent: botAgent}
	var visited sync.Map
	var wordToUrls sync.Map

	recursiveScrape(botAgent, guard, &visited, &wordToUrls, seedUrls, config, 1)

	for word, sites := range wordToUrls.Range {
		fmt.Printf("Word %s has sites %s", word, sites)
	}
}
