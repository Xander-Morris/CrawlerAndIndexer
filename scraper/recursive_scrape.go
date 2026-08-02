package scraper

import (
	"fmt"
	"main/data_structures"
	"main/database"
	"sync"
	"time"
)

type ScrapeConfig struct {
	MaxLevel           int
	MaxWorkers         int
	MaxInMemoryEntries int
}

type ScrapeInstances struct {
	BotAgent    string 
	Guard       *RobotsRules
	Visited  	*sync.Map 
	Config      *ScrapeConfig 
	WordsToUrls *data_structures.WordsToUrls
	Results     chan ScrapeResults
}

func fillUpWorkers(scrapeInstances *ScrapeInstances, startQueue []string) ([]string, time.Duration, int) {
	var wg sync.WaitGroup
	var deferred []string
	waitForNext := time.Duration(0)
	dispatchedCount := 0
	queue := startQueue

	for j := 0; j < scrapeInstances.Config.MaxWorkers && len(queue) > 0; {
		pageURL := queue[0]
		queue = queue[1:]

		if _, exists := scrapeInstances.Visited.Load(pageURL); exists {
			continue
		}

		robotConfig, err := scrapeInstances.Guard.GetRobotConfig(pageURL)

		if err != nil || robotConfig == nil || !robotConfig.Allowed {
			continue
		}

		timeNow := time.Now().UnixNano()
		wait := robotConfig.CrawlDelay - time.Duration(timeNow-robotConfig.LastRequestTime)

		if wait > 0 {
			deferred = append(deferred, pageURL)

			if waitForNext == 0 || wait < waitForNext {
				waitForNext = wait
			}

			continue
		}

		fmt.Printf("Fetching for %s\n", pageURL)
		robotConfig.LastRequestTime = timeNow

		j++
		dispatchedCount++
		scrapeInstances.Visited.Store(pageURL, "")
		wg.Go(func() { extractFromPage(scrapeInstances, pageURL) })
	}

	wg.Wait()
	queue = append(queue, deferred...)

	return queue, waitForNext, dispatchedCount
}

func processSeedUrls(scrapeInstances *ScrapeInstances, seedUrls []string) int {
	queue := seedUrls
	dispatchedCount := 0

	for len(queue) > 0 {
		newQueue, waitForNext, newDispatchedCount := fillUpWorkers(scrapeInstances, queue)
		queue = newQueue
		dispatchedCount += newDispatchedCount

		if len(scrapeInstances.WordsToUrls.GetItems()) >= scrapeInstances.Config.MaxInMemoryEntries {
			fmt.Println("Flushing to database!")

			if err := database.WriteToDatabase(scrapeInstances.WordsToUrls); err != nil {
				fmt.Printf("Flush failed, will retry next batch: %v\n", err)
			} else {
				scrapeInstances.WordsToUrls = data_structures.NewWordsToUrls()
			}
		}

		if len(queue) > 0 {
			if waitForNext <= 0 {
				waitForNext = 100 * time.Millisecond
			}

			time.Sleep(waitForNext)
		}
	}

	return dispatchedCount
}

func startRecursiveScrape(scrapeInstances *ScrapeInstances, seedUrls []string, level int) {
	fmt.Printf("About to start a recursive scrape with seed urls: %s\n", seedUrls)
	scrapeInstances.Results = make(chan ScrapeResults, len(seedUrls))
	dispatchedCount := processSeedUrls(scrapeInstances, seedUrls)

	if level >= scrapeInstances.Config.MaxLevel {
		return
	}

	var combinedNextUrls []string

	for range dispatchedCount {
		nextUrls := <-scrapeInstances.Results
		combinedNextUrls = append(combinedNextUrls, nextUrls.nextURLs...)
	}

	startRecursiveScrape(scrapeInstances, combinedNextUrls, level+1)
}

func ScrapeUrls(botAgent string, seedUrls []string, config *ScrapeConfig) {
	scrapeInstances := &ScrapeInstances{
		BotAgent:    botAgent,
		Guard:       &RobotsRules{userAgent: botAgent},
		Visited:     &sync.Map{},
		Config:      config,
		WordsToUrls: data_structures.NewWordsToUrls(),
	}

	startRecursiveScrape(scrapeInstances, seedUrls, 1)
}