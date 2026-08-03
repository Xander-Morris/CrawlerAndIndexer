package main

import (
	"fmt"
	"main/database"
	"main/jobs"
)

func main() {
	botAgent := "MyCustomScraperBot/1.0"
	remoteOk := jobs.NewRemoteOK(botAgent)
	fetchedJobs, err := remoteOk.FetchJobs()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Fetched %d jobs from RemoteOK\n", len(fetchedJobs))

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
