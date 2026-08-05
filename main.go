package main

import (
	"fmt"
	"main/database"
	"main/jobs"
)

func main() {
	botAgent := "MyCustomScraperBot/1.0"

	sources := []jobs.JobSource{
		jobs.NewRemoteOK(botAgent),
		jobs.NewRemotive(botAgent),
		jobs.NewArbeitnow(botAgent),
		jobs.NewJobicy(botAgent),
		jobs.NewHimalayas(botAgent),
		jobs.NewWeWorkRemotely(botAgent),
	}

	var fetchedJobs []jobs.Job

	for _, source := range sources {
		sourceJobs, err := source.FetchJobs()

		if err != nil {
			fmt.Printf("Failed to fetch jobs from %T: %v\n", source, err)
			continue
		}

		fmt.Printf("Fetched %d jobs from %T\n", len(sourceJobs), source)
		fetchedJobs = append(fetchedJobs, sourceJobs...)
	}

	if err := database.WriteToDatabase(fetchedJobs); err != nil {
		panic(err)
	}

	results, err := database.SearchForJobs(&database.JobSearchParams{
		SearchQuery:   "engineer",
		WorkplaceType: jobs.Remote,
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d matching jobs:\n", len(results))

	for _, job := range results {
		fmt.Printf("- %s at %s (%s)\n", job.Title, job.Company, job.URL)
	}
}
